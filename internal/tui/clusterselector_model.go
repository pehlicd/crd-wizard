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
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/pehlicd/crd-wizard/internal/clustermanager"
)

type clusterItem struct {
	name string
}

func (i clusterItem) Title() string       { return i.name }
func (i clusterItem) Description() string { return "" }
func (i clusterItem) FilterValue() string { return i.name }

type clusterSelectorModel struct {
	clusterMgr    *clustermanager.ClusterManager
	list          list.Model
	width, height int
	selectedName  string
}

func newClusterSelectorModel(clusterMgr *clustermanager.ClusterManager, currentCluster string, width, height int) clusterSelectorModel {
	clusters := clusterMgr.ListClusters()
	items := make([]list.Item, len(clusters))
	for i, name := range clusters {
		items[i] = clusterItem{name: name}
	}

	l := list.New(items, list.NewDefaultDelegate(), width, height-10)
	l.Title = "Select Cluster"
	l.SetShowHelp(true)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	return clusterSelectorModel{
		clusterMgr:   clusterMgr,
		list:         l,
		width:        width,
		height:       height,
		selectedName: currentCluster,
	}
}

func (m clusterSelectorModel) Init() tea.Cmd {
	return nil
}

func (m clusterSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.list.SetSize(m.width, m.height-10)

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, goBackCmd
		case "enter":
			if item, ok := m.list.SelectedItem().(clusterItem); ok {
				m.selectedName = item.name
				return m, switchClusterCmd(item.name)
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m clusterSelectorModel) View() string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Padding(1, 0)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(1, 0)

	header := headerStyle.Render(fmt.Sprintf("Current cluster: %s", m.selectedName))
	help := helpStyle.Render("↑/↓: navigate • enter: select • esc/q: back")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		m.list.View(),
		help,
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func switchClusterCmd(clusterName string) tea.Cmd {
	return func() tea.Msg {
		return switchClusterMsg{clusterName: clusterName}
	}
}
