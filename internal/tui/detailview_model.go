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

	"github.com/pehlicd/crd-explorer/internal/k8s"
	"github.com/pehlicd/crd-explorer/internal/models"
)

type detailViewTab uint

const (
	definitionTab detailViewTab = iota
	eventsTab
)

type detailModel struct {
	client        *k8s.Client
	crd           models.CRD
	instance      unstructured.Unstructured
	events        []corev1.Event
	yamlContent   string
	eventsContent string
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
}

func newDetailModel(client *k8s.Client, crd models.CRD, instance unstructured.Unstructured, width, height int) detailModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	vp := viewport.New(width-4, height-8)
	vp.Style = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).BorderForeground(lipgloss.Color("#7D56F4"))

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
		var wg sync.WaitGroup
		var err1, err2 error

		wg.Add(2)
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
		wg.Wait()

		if err1 != nil {
			return errMsg{err1}
		}
		if err2 != nil {
			return errMsg{err2}
		}

		return contentLoadedMsg{yamlStr: yamlStr, events: events}
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
		m.eventsContent = m.formatEvents()
		m.viewport.SetContent(m.yamlContent)
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
			m.activeTab = (m.activeTab + 1) % 2
			m.switchTabContent()
		case "left", "h":
			m.activeTab--
			if m.activeTab < definitionTab {
				m.activeTab = eventsTab
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
	if m.activeTab == definitionTab {
		m.viewport.SetContent(m.yamlContent)
	} else {
		m.viewport.SetContent(m.eventsContent)
	}
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

func (m detailModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n   %s %s\n\n", ErrStyle.Render("Error:"), m.err)
	}
	if m.loading {
		return fmt.Sprintf("\n   %s Loading details for %s...\n\n", m.spinner.View(), m.instance.GetName())
	}

	title := fmt.Sprintf("Details for %s: %s/%s", m.crd.Kind, m.instance.GetNamespace(), m.instance.GetName())

	var tabs []string
	if m.activeTab == definitionTab {
		tabs = []string{ActiveTabStyle.Render("Definition"), InactiveTabStyle.Render("Events")}
	} else {
		tabs = []string{InactiveTabStyle.Render("Definition"), ActiveTabStyle.Render("Events")}
	}
	tabHeader := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	help := "[↑/↓] Scroll | [Tab] Switch Pane | [b] Back | [q] Quit"

	return lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render(title),
		tabHeader,
		m.viewport.View(),
	) + "\n" + HelpStyle.Render(help)
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
