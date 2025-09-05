/*
Copyright © 2025 Furkan Pehlivan furkanpehlivan34@gmail.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package tui

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/pehlicd/crd-wizard/internal/k8s"
	"github.com/pehlicd/crd-wizard/internal/models"
)

type detailViewTab int

const (
	graphTab detailViewTab = iota
	definitionTab
	eventsTab
)

type detailModel struct {
	client        *k8s.Client
	crd           models.CRD
	instance      unstructured.Unstructured
	events        []corev1.Event
	graph         *models.ResourceGraph
	yamlContent   string
	eventsContent string
	graphContent  string
	viewport      viewport.Model
	spinner       spinner.Model
	activeTab     detailViewTab
	loading       bool
	err           error
	width, height int
}

type contentLoadedMsg struct {
	yamlStr string
	events  []corev1.Event
	graph   *models.ResourceGraph
}

func newDetailModel(client *k8s.Client, crd models.CRD, instance unstructured.Unstructured, width, height int) detailModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	vp := viewport.New(width-4, height-8)
	vp.Style = lipgloss.NewStyle().Margin(0, 1).Border(lipgloss.NormalBorder(), true).BorderForeground(lipgloss.Color("#7D56F4")).Align(lipgloss.Left)

	return detailModel{
		client:   client,
		crd:      crd,
		instance: instance,
		viewport: vp,
		spinner:  s,
		loading:  true,
		width:    width,
		height:   height,
	}
}

func (m detailModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		var yamlStr string
		var events []corev1.Event
		var graph *models.ResourceGraph
		var wg sync.WaitGroup
		var err1, err2, err3 error

		wg.Add(3)
		go func() {
			defer wg.Done()
			yamlBytes, err := yaml.Marshal(m.instance.Object)
			if err != nil {
				err1 = err
				return
			}
			yamlStr, err1 = highlightYAML(string(yamlBytes))
		}()
		go func() {
			defer wg.Done()
			events, err2 = m.client.GetEvents(context.Background(), m.crd.Name, string(m.instance.GetUID()))
		}()
		go func() {
			defer wg.Done()
			// Fetch the resource graph using the actual client method.
			graph, err3 = m.client.GetResourceGraph(context.Background(), string(m.instance.GetUID()))
		}()
		wg.Wait()

		if err1 != nil {
			return errMsg{err1}
		}
		if err2 != nil {
			return errMsg{err2}
		}
		if err3 != nil {
			return errMsg{err3}
		}

		return contentLoadedMsg{yamlStr: yamlStr, events: events, graph: graph}
	})
}

func (m detailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 8
	case contentLoadedMsg:
		m.loading = false
		m.yamlContent = msg.yamlStr
		m.events = msg.events
		m.graph = msg.graph
		m.eventsContent = m.formatEvents()
		m.graphContent = m.formatGraph()
		m.switchTabContent() // Set initial content based on active tab
	case errMsg:
		m.err = msg.err
		m.loading = false
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "b", "esc":
			return m, func() tea.Msg { return goBackMsg{} }
		case "tab", "right", "l":
			m.activeTab = (m.activeTab + 1) % 3
			m.switchTabContent()
		case "left", "h":
			m.activeTab--
			if m.activeTab < definitionTab {
				m.activeTab = graphTab
			}
			m.switchTabContent()
		}
	}

	if m.loading {
		m.spinner, cmd = m.spinner.Update(msg)
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *detailModel) switchTabContent() {
	switch m.activeTab {
	case definitionTab:
		m.viewport.SetContent(m.yamlContent)
	case eventsTab:
		m.viewport.SetContent(m.eventsContent)
	case graphTab:
		m.viewport.SetContent(m.graphContent)
	}
	m.viewport.GotoTop()
}

func (m detailModel) formatEvents() string {
	if len(m.events) == 0 {
		return "No events found for this resource."
	}
	var b strings.Builder
	for _, e := range m.events {
		t := e.LastTimestamp.Time
		if t.IsZero() {
			t = e.FirstTimestamp.Time
		}

		eventType := e.Type
		if e.Type == "Warning" {
			eventType = ErrStyle.Render(e.Type)
		}

		b.WriteString(fmt.Sprintf("%s  %s  %s  %s\n",
			k8s.HumanReadableAge(t),
			eventType,
			e.Reason,
			e.Message,
		))
	}
	return b.String()
}

func (m detailModel) formatGraph() string {
	if m.graph == nil || len(m.graph.Nodes) == 0 {
		return "No resource graph available."
	}

	var b strings.Builder
	nodes := make(map[string]models.Node)
	for _, n := range m.graph.Nodes {
		nodes[n.ID] = n
	}

	adj := make(map[string][]string)
	isTarget := make(map[string]bool)
	for _, e := range m.graph.Edges {
		adj[e.Source] = append(adj[e.Source], e.Target)
		isTarget[e.Target] = true
	}

	// Find root nodes for rendering the tree structure.
	// A root is any node that is not a target of an edge.
	var roots []string
	for _, n := range m.graph.Nodes {
		if !isTarget[n.ID] {
			roots = append(roots, n.ID)
		}
	}

	for _, rootID := range roots {
		m.dfsRender(&b, rootID, "", true, nodes, adj)
	}

	return b.String()
}

func (m detailModel) dfsRender(b *strings.Builder, nodeID, prefix string, isLast bool, nodes map[string]models.Node, adj map[string][]string) {
	node, ok := nodes[nodeID]
	if !ok {
		return
	}

	// Style the node type with a color
	kindColor := getColorForKind(node.Type)
	styledType := lipgloss.NewStyle().Foreground(kindColor).Render(node.Type)
	label := fmt.Sprintf("[%s: %s]", styledType, node.Label)

	// Highlight the resource this detail view is for
	if node.ID == string(m.instance.GetUID()) {
		label = lipgloss.NewStyle().Bold(true).Render(label + " [*]")
	}

	b.WriteString(prefix)
	if isLast {
		b.WriteString("└── ")
		prefix += "    "
	} else {
		b.WriteString("├── ")
		prefix += "│   "
	}
	b.WriteString(label)
	b.WriteString("\n")

	children := adj[nodeID]
	for i, childID := range children {
		m.dfsRender(b, childID, prefix, i == len(children)-1, nodes, adj)
	}
}

// getColorForKind returns a specific color for each Kubernetes resource type
// to make the graph more readable, based on the provided color scheme.
func getColorForKind(kind string) lipgloss.Color {
	switch kind {
	// Workload Resources
	case "Pod":
		return lipgloss.Color("#0EA5E9") // sky
	case "Deployment":
		return lipgloss.Color("#10B981") // emerald
	case "StatefulSet":
		return lipgloss.Color("#F59E0B") // amber
	case "DaemonSet":
		return lipgloss.Color("#14B8A6") // teal
	case "Job":
		return lipgloss.Color("#8B5CF6") // violet
	case "CronJob":
		return lipgloss.Color("#D946EF") // fuchsia
	case "ReplicaSet":
		return lipgloss.Color("#06B6D4") // cyan
	case "ReplicationController":
		return lipgloss.Color("#3B82F6") // blue

	// Service Discovery & Load Balancing
	case "Service":
		return lipgloss.Color("#F97316") // orange
	case "Ingress":
		return lipgloss.Color("#6366F1") // indigo
	case "Endpoint", "EndpointSlice":
		return lipgloss.Color("#EC4899") // pink

	// Configuration & Storage
	case "ConfigMap":
		return lipgloss.Color("#84CC16") // lime
	case "Secret":
		return lipgloss.Color("#EF4444") // red
	case "PersistentVolume":
		return lipgloss.Color("#EAB308") // yellow
	case "PersistentVolumeClaim":
		return lipgloss.Color("#22C55E") // green
	case "StorageClass":
		return lipgloss.Color("#A855F7") // purple

	// Security & RBAC
	case "ServiceAccount":
		return lipgloss.Color("#71717A") // zinc
	case "Role", "ClusterRole":
		return lipgloss.Color("#38BDF8") // sky
	case "RoleBinding", "ClusterRoleBinding":
		return lipgloss.Color("#FB923C") // orange

	// Policy Resources
	case "NetworkPolicy":
		return lipgloss.Color("#22D3EE") // cyan
	case "PodDisruptionBudget":
		return lipgloss.Color("#34D399") // emerald

	// Custom Resources
	case "CustomResourceDefinition":
		return lipgloss.Color("#818CF8") // indigo

	default:
		return lipgloss.Color("#FFFFFF") // Default to white
	}
}

func (m detailModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n   %s %s\n\n", ErrStyle.Render("Error:"), m.err)
	}
	if m.loading {
		return fmt.Sprintf("\n   %s Loading details for %s...\n\n", m.spinner.View(), m.instance.GetName())
	}

	title := fmt.Sprintf("%s: %s/%s", m.crd.Kind, m.instance.GetNamespace(), m.instance.GetName())

	tabNames := []string{"Graph", "Definition", "Events"}
	tabs := make([]string, len(tabNames))
	for i, name := range tabNames {
		if detailViewTab(i) == m.activeTab {
			tabs[i] = ActiveTabStyle.Render(name)
		} else {
			tabs[i] = InactiveTabStyle.Render(name)
		}
	}
	tabHeader := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	help := "[↑/↓] Scroll | [Tab] Switch Pane | [b] Back | [q] Quit"

	titleStyle := TitleStyle.Margin(0, 0, 1)

	view := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(title),
		tabHeader,
		m.viewport.View(),
	) + "\n" + HelpStyle.Render(help)

	return AppStyle.Render(view)
}

func highlightYAML(content string) (string, error) {
	l := lexers.Get("yaml")
	s := styles.Get("dracula")
	f := formatters.Get("terminal256")
	it, err := l.Tokenise(nil, content)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = f.Format(&buf, s, it)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RFC3339ToTime Small helper in k8s package for time parsing
func RFC3339ToTime(ts string) (time.Time, error) {
	return time.Parse(time.RFC3339, ts)
}
