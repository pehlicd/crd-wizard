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

	tea "github.com/charmbracelet/bubbletea"

	"github.com/pehlicd/crd-wizard/internal/k8s"
	"github.com/pehlicd/crd-wizard/internal/models"
)

type currentView uint

const (
	crdListView currentView = iota
	instanceListView
	detailView
)

type mainModel struct {
	client            *k8s.Client
	view              currentView
	err               error
	width, height     int
	crdListModel      tea.Model
	instanceListModel tea.Model
	detailViewModel   tea.Model
}

func newMainModel(client *k8s.Client, crdName, kind string) mainModel {
	model := mainModel{
		client:       client,
		view:         crdListView,
		crdListModel: newCRDListModel(client, nil),
	}

	// If a CRD name or Kind is provided via flags, fetch it and pre-filter crdList view
	if crdName != "" || kind != "" {
		var targetCRD []models.CRD
		var err error

		allCRDs, err := client.GetCRDs(context.Background())
		if err != nil {
			model.err = fmt.Errorf("failed to list CRDs to find match: %w", err)
			return model
		}

		for _, crd := range allCRDs {
			if (crdName != "" && crd.Name == crdName) || (kind != "" && crd.Kind == kind) {
				targetCRD = append(targetCRD, crd)
			}
		}

		if len(targetCRD) != 0 {
			model.crdListModel = newCRDListModel(client, targetCRD)
			return model
		}
	}

	return model
}

func (m mainModel) Init() tea.Cmd {
	return m.crdListModel.Init()
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		// Propagate the size message to all models so they can resize correctly
		// even when not in focus. This is important for the new tabbed view.
		if m.crdListModel != nil {
			m.crdListModel, _ = m.crdListModel.Update(msg)
		}
		if m.instanceListModel != nil {
			m.instanceListModel, _ = m.instanceListModel.Update(msg)
		}
		if m.detailViewModel != nil {
			m.detailViewModel, _ = m.detailViewModel.Update(msg)
		}

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case showInstancesMsg:
		m.instanceListModel = newInstanceListModel(m.client, msg.crd, m.width, m.height)
		cmds = append(cmds, m.instanceListModel.Init())
		m.view = instanceListView

	case showDetailsMsg:
		m.detailViewModel = newDetailModel(m.client, msg.crd, msg.instance, m.width, m.height)
		cmds = append(cmds, m.detailViewModel.Init())
		m.view = detailView

	case goBackMsg:
		// Improved back navigation logic
		switch m.view {
		case detailView:
			m.view = instanceListView
		case instanceListView:
			m.view = crdListView
			cmds = append(cmds, m.instanceListModel.Init())
		default:
			m.view = instanceListView
		}

	case errMsg:
		m.err = msg.err
	}

	// Route updates to the active view model
	switch m.view {
	case crdListView:
		m.crdListModel, cmd = m.crdListModel.Update(msg)
	case instanceListView:
		m.instanceListModel, cmd = m.instanceListModel.Update(msg)
	case detailView:
		m.detailViewModel, cmd = m.detailViewModel.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	switch m.view {
	case crdListView:
		return m.crdListModel.View()
	case instanceListView:
		return m.instanceListModel.View()
	case detailView:
		return m.detailViewModel.View()
	default:
		return "Unknown view"
	}
}
