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

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/pehlicd/crd-wizard/internal/k8s"
	"github.com/pehlicd/crd-wizard/internal/models"
)

type instanceListModel struct {
	client        *k8s.Client
	crd           models.CRD
	table         table.Model
	spinner       spinner.Model
	instances     []unstructured.Unstructured
	loading       bool
	err           error
	width, height int
}

type instancesLoadedMsg struct{ instances []unstructured.Unstructured }

func newInstanceListModel(client *k8s.Client, crd models.CRD, width, height int) instanceListModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	cols := []table.Column{
		{Title: "NAME", Width: 40},
		{Title: "NAMESPACE", Width: 30},
		{Title: "STATUS", Width: 20},
		{Title: "AGE", Width: 10},
	}
	tbl := table.New(
		table.WithColumns(cols),
		table.WithFocused(true),
		table.WithHeight(height-8),
	)
	tbl.SetStyles(table.Styles{
		Header:   HeaderStyle,
		Selected: SelectedStyle,
	})

	return instanceListModel{
		client:  client,
		crd:     crd,
		table:   tbl,
		spinner: s,
		loading: true,
		width:   width,
		height:  height,
	}
}

func (m instanceListModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		instances, err := m.client.GetCRsForCRD(context.Background(), m.crd.Name)
		if err != nil {
			return errMsg{err}
		}
		return instancesLoadedMsg{instances}
	})
}

func (m instanceListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.table.SetHeight(m.height - 8)
	case instancesLoadedMsg:
		m.loading = false
		m.instances = msg.instances
		m.updateTableRows()
	case errMsg:
		m.err = msg.err
		m.loading = false
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "b", "esc":
			return m, func() tea.Msg { return goBackMsg{} }
		case "enter":
			if m.table.Cursor() < len(m.instances) {
				selected := m.instances[m.table.Cursor()]
				return m, func() tea.Msg { return showDetailsMsg{crd: m.crd, instance: selected} }
			}
		}
	}
	if m.loading {
		m.spinner, cmd = m.spinner.Update(msg)
	} else {
		m.table, cmd = m.table.Update(msg)
	}
	return m, cmd
}

func (m *instanceListModel) updateTableRows() {
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
	if m.loading {
		return fmt.Sprintf("\n   %s Fetching instances for %s...\n\n", m.spinner.View(), m.crd.Kind)
	}

	title := fmt.Sprintf("Instances for: %s", m.crd.Name)
	help := "[↑/↓] Navigate | [Enter] Select | [b] Back | [q] Quit"

	return lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render(title),
		m.table.View(),
	) + "\n" + HelpStyle.Render(help)
}
