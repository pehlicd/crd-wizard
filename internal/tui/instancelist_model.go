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
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/pehlicd/crd-wizard/internal/k8s"
	"github.com/pehlicd/crd-wizard/internal/models"
)

type tab int

type instancesLoadedMsg struct{ instances []unstructured.Unstructured }
type fullCRDLoadedMsg struct {
	def *apiextensionsv1.CustomResourceDefinition
}

const (
	instancesTab tab = iota
	schemaTab
)

var (
	tabStyle         = lipgloss.NewStyle().Padding(0, 2).Margin(0, 1)
	activeTabStyle   = tabStyle.Background(lipgloss.Color("#7D56F4")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	inactiveTabStyle = tabStyle.Background(lipgloss.Color("#242424"))
	tabRowStyle      = lipgloss.NewStyle().Margin(1, 0)

	schemaKeyStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true)
	schemaTypeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("228")).
			Background(lipgloss.Color("63")).
			Padding(0, 1).
			MarginLeft(1)
	schemaDescStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Italic(true)
	focusedNodeStyle = lipgloss.NewStyle().Background(lipgloss.Color("236"))

	expandIcon   = "▾ "
	collapseIcon = "▸ "
)

type schemaNode struct {
	name        string
	propType    string
	description string
	children    []*schemaNode
	parent      *schemaNode
	expanded    bool
}

type instanceListModel struct {
	client          *k8s.Client
	crd             models.CRD
	fullDefinition  *apiextensionsv1.CustomResourceDefinition
	table           table.Model
	spinner         spinner.Model
	viewport        viewport.Model
	instances       []unstructured.Unstructured
	loading         bool
	err             error
	width, height   int
	activeTab       tab
	schemaRoot      []*schemaNode // The full tree
	flattenedSchema []*schemaNode // The visible nodes for rendering and navigation
	schemaCursor    int           // The cursor position in the flattenedSchema
}

func newInstanceListModel(client *k8s.Client, crd models.CRD, width, height int) instanceListModel {
	return newInstanceListModelWithActiveTab(client, crd, width, height, instancesTab)
}
func newInstanceListModelWithActiveTab(client *k8s.Client, crd models.CRD, width, height int, activeTab tab) instanceListModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	tableHeight := height - 8
	tbl := table.New(
		table.WithColumns([]table.Column{
			{Title: "NAME", Width: 40},
			{Title: "NAMESPACE", Width: 30},
			{Title: "STATUS", Width: 20},
			{Title: "AGE", Width: 10},
		}),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)
	tbl.SetStyles(table.Styles{Header: HeaderStyle, Selected: SelectedStyle})

	vp := viewport.New(width, tableHeight)
	vp.Style = lipgloss.NewStyle().Padding(0, 1)
	vp.SetContent("Loading schema...")

	return instanceListModel{
		client:    client,
		crd:       crd,
		table:     tbl,
		spinner:   s,
		viewport:  vp,
		loading:   true,
		width:     width,
		height:    height,
		activeTab: activeTab,
	}
}

func (m instanceListModel) Init() tea.Cmd {
	fetchInstancesCmd := func() tea.Msg {
		instances, err := m.client.GetCRsForCRD(context.Background(), m.crd.Name)
		if err != nil {
			return errMsg{err}
		}
		return instancesLoadedMsg{instances: instances}
	}
	fetchFullCRDCmd := func() tea.Msg {
		def, err := m.client.GetFullCRD(context.Background(), m.crd.Name)
		if err != nil {
			return errMsg{fmt.Errorf("failed to get full CRD definition: %w", err)}
		}
		return fullCRDLoadedMsg{def: def}
	}
	return tea.Batch(m.spinner.Tick, fetchInstancesCmd, fetchFullCRDCmd)
}

func (m instanceListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	var viewportNeedsUpdate bool

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		tableHeight := m.height - 8
		m.table.SetHeight(tableHeight)
		m.viewport.Width = msg.Width
		m.viewport.Height = tableHeight
		viewportNeedsUpdate = true

	case instancesLoadedMsg:
		m.loading = false
		m.instances = msg.instances
		m.updateTableRows()

	case fullCRDLoadedMsg:
		m.fullDefinition = msg.def
		m.schemaRoot = m.buildSchemaTree()
		m.flattenSchema()
		viewportNeedsUpdate = true

	case errMsg:
		m.err = msg.err
		m.loading = false

	case tea.KeyMsg:
		if m.activeTab == schemaTab {
			if m.handleSchemaKeys(msg) {
				viewportNeedsUpdate = true
			}
		} else if m.activeTab == instancesTab && !m.loading {
			if msg.String() == "enter" {
				if m.table.Cursor() < len(m.instances) {
					selected := m.instances[m.table.Cursor()]
					return m, func() tea.Msg { return showDetailsMsg{crd: m.crd, instance: selected} }
				}
			}
		}

		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "b", "esc":
			return m, func() tea.Msg { return goBackMsg{} }
		case "tab", "right", "left", "shift+tab":
			m.activeTab = (m.activeTab + 1) % 2
			if m.activeTab == instancesTab {
				m.table.Focus()
			} else {
				m.table.Blur()
			}
			viewportNeedsUpdate = true
		}
	}

	if m.loading {
		m.spinner, cmd = m.spinner.Update(msg)
	} else {
		var tableCmd, viewportCmd tea.Cmd
		m.table, tableCmd = m.table.Update(msg)
		m.viewport, viewportCmd = m.viewport.Update(msg) // Allow mouse scrolling
		cmd = tea.Batch(tableCmd, viewportCmd)
	}
	cmds = append(cmds, cmd)

	// If any state change occurred that affects the schema view, update the viewport now.
	if viewportNeedsUpdate {
		m.updateViewportContent()
	}

	return m, tea.Batch(cmds...)
}

// handleSchemaKeys returns true if the view needs to be updated.
func (m *instanceListModel) handleSchemaKeys(msg tea.KeyMsg) bool {
	var changed bool
	switch msg.String() {
	case "up", "k":
		if m.schemaCursor > 0 {
			m.schemaCursor--
			changed = true
		}
	case "down", "j":
		if m.schemaCursor < len(m.flattenedSchema)-1 {
			m.schemaCursor++
			changed = true
		}
	case "enter", "l", " ":
		if m.schemaCursor >= 0 && m.schemaCursor < len(m.flattenedSchema) {
			node := m.flattenedSchema[m.schemaCursor]
			if len(node.children) > 0 {
				node.expanded = !node.expanded
				m.flattenSchema()
				changed = true
			}
		}
	}
	return changed
}

func (m *instanceListModel) updateTableRows() {
	if len(m.instances) == 0 {
		m.table.SetRows([]table.Row{{"No instances found for this CRD.", "", "", ""}})
		return
	}
	rows := make([]table.Row, len(m.instances))
	for i, inst := range m.instances {
		status, _, _ := unstructured.NestedString(inst.Object, "status", "phase")
		if status == "" {
			if conditions, found, _ := unstructured.NestedSlice(inst.Object, "status", "conditions"); found && len(conditions) > 0 {
				if firstCond, ok := conditions[0].(map[string]interface{}); ok {
					status, _, _ = unstructured.NestedString(firstCond, "reason")
				}
			}
		}
		if status == "" {
			status = "Unknown"
		}
		ts, _, _ := unstructured.NestedString(inst.Object, "metadata", "creationTimestamp")
		t, _ := RFC3339ToTime(ts)
		rows[i] = table.Row{inst.GetName(), inst.GetNamespace(), status, k8s.HumanReadableAge(t)}
	}
	m.table.SetRows(rows)
}

func (m instanceListModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n   %s %s\n\n", ErrStyle.Render("Error:"), m.err)
	}
	title := TitleStyle.Render(fmt.Sprintf("CRD: %s", m.crd.Name))
	tabHeaders := []string{"Instances", "Schema"}
	renderedTabs := make([]string, len(tabHeaders))
	for i, t := range tabHeaders {
		style := inactiveTabStyle
		if tab(i) == m.activeTab {
			style = activeTabStyle
		}
		renderedTabs[i] = style.Render(t)
	}
	tabs := tabRowStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...))
	var tabContent string
	if m.loading {
		tabContent = fmt.Sprintf("\n   %s Fetching instances for %s...\n\n", m.spinner.View(), m.crd.Kind)
	} else {
		switch m.activeTab {
		case instancesTab:
			tabContent = m.table.View()
		case schemaTab:
			// The viewport's content is set in Update, so we just View it here.
			tabContent = m.viewport.View()
		}
	}
	help := "[←/→] Switch Tab | [↑/↓] Navigate | [Enter] Expand/Collapse | [b] Back | [q] Quit"
	return lipgloss.JoinVertical(lipgloss.Left, title, tabs, tabContent) + "\n" + HelpStyle.Render(help)
}

func (m *instanceListModel) buildSchemaTree() []*schemaNode {
	if m.fullDefinition == nil {
		return nil
	}
	var openAPISchema *apiextensionsv1.JSONSchemaProps
	for _, v := range m.fullDefinition.Spec.Versions {
		if v.Served {
			openAPISchema = v.Schema.OpenAPIV3Schema
			break
		}
	}
	if openAPISchema == nil {
		return nil
	}
	props, ok := openAPISchema.Properties["spec"]
	if !ok || props.Properties == nil {
		return nil
	}
	return m.parseProperties(nil, props.Properties)
}

func (m *instanceListModel) parseProperties(parent *schemaNode, properties map[string]apiextensionsv1.JSONSchemaProps) []*schemaNode {
	keys := make([]string, 0, len(properties))
	for k := range properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var nodes []*schemaNode
	for _, key := range keys {
		prop := properties[key]
		node := &schemaNode{
			name:        key,
			propType:    prop.Type,
			description: prop.Description,
			parent:      parent,
		}
		if prop.Type == "object" && prop.Properties != nil {
			node.children = m.parseProperties(node, prop.Properties)
		}
		if prop.Type == "array" && prop.Items != nil && prop.Items.Schema != nil {
			itemSchema := prop.Items.Schema
			itemNode := &schemaNode{
				name:        "[items]",
				propType:    itemSchema.Type,
				description: "Defines the structure of items in the array.",
				parent:      node,
			}
			if itemSchema.Type == "object" && itemSchema.Properties != nil {
				itemNode.children = m.parseProperties(itemNode, itemSchema.Properties)
			}
			node.children = []*schemaNode{itemNode}
		}
		nodes = append(nodes, node)
	}
	return nodes
}

func (m *instanceListModel) flattenSchema() {
	m.flattenedSchema = []*schemaNode{}
	var flatten func([]*schemaNode)
	flatten = func(nodes []*schemaNode) {
		for _, node := range nodes {
			m.flattenedSchema = append(m.flattenedSchema, node)
			if node.expanded {
				flatten(node.children)
			}
		}
	}
	flatten(m.schemaRoot)
	if m.schemaCursor >= len(m.flattenedSchema) {
		m.schemaCursor = len(m.flattenedSchema) - 1
	}
	if m.schemaCursor < 0 {
		m.schemaCursor = 0
	}
}

func (m *instanceListModel) getDepth(node *schemaNode) int {
	depth := 0
	for p := node.parent; p != nil; p = p.parent {
		depth++
	}
	return depth
}

// updateViewportContent is called from Update() and handles all rendering and viewport updates.
func (m *instanceListModel) updateViewportContent() {
	if m.activeTab != schemaTab {
		m.viewport.SetContent("") // Clear content if not visible
		return
	}
	if len(m.flattenedSchema) == 0 {
		m.viewport.SetContent("Schema not available or empty.")
		return
	}

	// This struct helps track the layout of each node.
	type nodeLayout struct {
		startLine int
		endLine   int
	}
	layouts := make(map[*schemaNode]nodeLayout)
	currentLine := 0
	contentWidth := m.viewport.Width - 2 // Approximate content width
	if contentWidth < 1 {
		contentWidth = 1
	}

	// First pass: calculate the layout of each node, accounting for word wrap.
	for _, node := range m.flattenedSchema {
		start := currentLine
		// The main property line is always at least 1 line.
		currentLine++
		// Add lines for the description if it's visible and wraps.
		if node.description != "" && (len(node.children) == 0 || node.expanded) {
			descIndentStr := strings.Repeat("    ", m.getDepth(node)+1)
			descContentWidth := contentWidth - lipgloss.Width(descIndentStr)
			if descContentWidth < 1 {
				descContentWidth = 1
			}
			// Use lipgloss to calculate height of the wrapped description.
			wrappedDesc := lipgloss.NewStyle().Width(descContentWidth).Render(node.description)
			currentLine += lipgloss.Height(wrappedDesc)
		}
		layouts[node] = nodeLayout{startLine: start, endLine: currentLine - 1}
	}

	// Second pass: build the string content for the viewport.
	var b strings.Builder
	for i, node := range m.flattenedSchema {
		depth := m.getDepth(node)
		indent := strings.Repeat("    ", depth)

		icon := "  "
		if len(node.children) > 0 {
			if node.expanded {
				icon = expandIcon
			} else {
				icon = collapseIcon
			}
		}

		line := fmt.Sprintf("%s%s%s %s",
			indent,
			icon,
			schemaKeyStyle.Render(node.name),
			schemaTypeStyle.Render(node.propType),
		)

		// Only highlight the main property line
		if i == m.schemaCursor {
			line = focusedNodeStyle.Render(line)
		}
		b.WriteString(line + "\n")

		if node.description != "" && (len(node.children) == 0 || node.expanded) {
			descIndentStr := strings.Repeat("    ", depth+1)
			descContentWidth := contentWidth - lipgloss.Width(descIndentStr)
			if descContentWidth < 1 {
				descContentWidth = 1
			}
			// Render with the same wrapping to match height calculation
			wrappedDesc := lipgloss.NewStyle().Width(descContentWidth).Render(schemaDescStyle.Render(node.description))
			descLine := fmt.Sprintf("%s%s", descIndentStr, wrappedDesc)

			// Do NOT highlight the description line
			b.WriteString(descLine + "\n")
		}
	}

	// Finally, update the viewport content and adjust its scroll position.
	m.viewport.SetContent(b.String())
	if m.viewport.Height > 0 && m.schemaCursor >= 0 && m.schemaCursor < len(m.flattenedSchema) {
		selectedLayout := layouts[m.flattenedSchema[m.schemaCursor]]
		if selectedLayout.startLine < m.viewport.YOffset {
			m.viewport.YOffset = selectedLayout.startLine
		} else if selectedLayout.endLine >= m.viewport.YOffset+m.viewport.Height {
			m.viewport.YOffset = selectedLayout.endLine - m.viewport.Height + 1
		}
	}
}
