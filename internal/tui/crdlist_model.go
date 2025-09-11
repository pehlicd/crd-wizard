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
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/pehlicd/crd-wizard/internal/k8s"
	"github.com/pehlicd/crd-wizard/internal/models"
)

type crdListModel struct {
	client        *k8s.Client
	table         table.Model
	spinner       spinner.Model
	textInput     textinput.Model
	crds          []models.CRD
	filteredCRDs  []models.CRD
	loading       bool
	filtering     bool
	err           error
	width, height int
}

func newCRDListModel(client *k8s.Client, targetCRDs []models.CRD) crdListModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	// Define columns with initial (placeholder) widths.
	// These will be dynamically resized on window size changes.
	cols := []table.Column{
		{Title: "KIND", Width: 20},
		{Title: "FULL NAME", Width: 40},
		{Title: "INSTANCES", Width: 15},
	}
	tbl := table.New(
		table.WithColumns(cols),
		table.WithFocused(true),
		// Set an initial height. This will also be resized.
		table.WithHeight(15),
	)

	// Set the styles for all parts of the table for consistent alignment.
	tbl.SetStyles(table.Styles{
		Header:   HeaderStyle,
		Cell:     CellStyle,
		Selected: SelectedStyle,
	})

	ti := textinput.New()
	ti.Placeholder = "Filter by name or kind..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	return crdListModel{
		client:       client,
		table:        tbl,
		spinner:      s,
		textInput:    ti,
		loading:      true,
		filteredCRDs: targetCRDs,
	}
}

func (m crdListModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		if len(m.filteredCRDs) != 0 {
			return crdsLoadedMsg{m.filteredCRDs}
		}
		crds, err := m.client.GetCRDs(context.Background())
		if err != nil {
			return errMsg{err}
		}
		return crdsLoadedMsg{crds}
	})
}

func (m crdListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		appHorizontalMargin, appVerticalMargin := AppStyle.GetHorizontalFrameSize(), AppStyle.GetVerticalFrameSize()

		headerHeight := 1
		footerHeight := 2

		verticalSpaceForTable := m.height - appVerticalMargin - headerHeight - footerHeight
		if m.filtering {
			verticalSpaceForTable--
		}
		m.table.SetHeight(verticalSpaceForTable)

		// Set the width for the table and text input.
		m.table.SetWidth(m.width - appHorizontalMargin)
		m.textInput.Width = m.width - appHorizontalMargin

		instancesColWidth := 15
		remainingWidth := m.table.Width() - instancesColWidth - 4
		kindColWidth := int(float64(remainingWidth) * 0.35)
		fullNameColWidth := remainingWidth - kindColWidth

		newColumns := m.table.Columns()
		newColumns[0].Width = kindColWidth
		newColumns[1].Width = fullNameColWidth
		newColumns[2].Width = instancesColWidth
		m.table.SetColumns(newColumns)

	case crdsLoadedMsg:
		m.loading = false
		m.crds = msg.crds
		m.filteredCRDs = msg.crds
		m.updateTableRows()
	case errMsg:
		m.err = msg.err
		m.loading = false
	case tea.KeyMsg:
		if m.filtering {
			switch msg.String() {
			case "enter", "esc":
				m.filtering = false
			default:
				m.textInput, cmd = m.textInput.Update(msg)
				m.filterTable()
			}
			return m, cmd
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "/":
			m.filtering = true
			return m, nil
		case "enter":
			if m.table.Cursor() < len(m.filteredCRDs) {
				selectedCRD := m.filteredCRDs[m.table.Cursor()]
				return m, func() tea.Msg { return showInstancesMsg{crd: selectedCRD} }
			}
		case "r", "R":
			m.loading = true
			m.err = nil
			return m, func() tea.Msg {
				crds, err := m.client.GetCRDs(context.Background())
				if err != nil {
					return errMsg{err}
				}
				return crdsLoadedMsg{crds}
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

func (m *crdListModel) filterTable() {
	val := strings.ToLower(m.textInput.Value())
	if val == "" {
		m.filteredCRDs = m.crds
	} else {
		filtered := make([]models.CRD, 0)
		for _, crd := range m.crds {
			if strings.Contains(strings.ToLower(crd.Name), val) || strings.Contains(strings.ToLower(crd.Kind), val) {
				filtered = append(filtered, crd)
			}
		}
		m.filteredCRDs = filtered
	}
	m.table.SetCursor(0)
	m.updateTableRows()
}

func (m *crdListModel) updateTableRows() {
	crdsCount := len(m.filteredCRDs)
	if crdsCount < 1 {
		m.table.SetRows([]table.Row{[]string{"No CRD found!", "", ""}})
		return
	}

	rows := make([]table.Row, crdsCount)
	for i, crd := range m.filteredCRDs {
		instanceText := fmt.Sprintf("%d in use", crd.InstanceCount)
		if crd.InstanceCount == 0 {
			instanceText = "Not in use"
		}
		rows[i] = table.Row{crd.Kind, crd.Name, instanceText}
	}
	m.table.SetRows(rows)
}

func (m crdListModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n   %s %s\n\n", ErrStyle.Render("Error:"), m.err)
	}
	if m.loading {
		return fmt.Sprintf("\n   %s Fetching CRDs from cluster...\n\n", m.spinner.View())
	}

	var viewContent string
	var help string
	titlestyle := TitleStyle.PaddingBottom(1)

	if m.filtering {
		help = "[Enter/Esc] Confirm/Cancel Filter"
		viewContent = lipgloss.JoinVertical(lipgloss.Left,
			titlestyle.Render("️🧙 CRD Wizard"),
			m.textInput.View(),
			m.table.View(),
		)
	} else {
		help = "[↑/↓] Navigate | [Enter] Select | [/] Filter | [r] Refresh | [q] Quit"
		viewContent = lipgloss.JoinVertical(lipgloss.Left,
			titlestyle.Render("🧙 CRD Wizard - CRD Selector"),
			m.table.View(),
		)
	}

	// Wrap the entire view in the AppStyle to provide consistent margins.
	return AppStyle.Render(viewContent + "\n" + HelpStyle.Render(help))
}
