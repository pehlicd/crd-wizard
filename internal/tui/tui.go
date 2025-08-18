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
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pehlicd/crd-wizard/internal/k8s"
	"github.com/pehlicd/crd-wizard/internal/models"
)

// Start initializes and runs the Bubble Tea TUI.
func Start(client *k8s.Client, crdName string, kind string) error {
	mainModel := newMainModel(client)
	p := tea.NewProgram(mainModel, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if kind != "" || crdName != "" {
		var targetCRD models.CRD
		var found bool

		allCRDs, err := client.GetCRDs(context.Background())
		if err != nil {
			return err
		}

		for _, crd := range allCRDs {
			if (crdName != "" && crd.Name == crdName) || (kind != "" && crd.Kind == kind) {
				targetCRD = crd
				found = true
				break
			}
		}
		if found {
			// Set the view to instanceListView and initialize the corresponding model.
			instanceModel := newInstanceListModelWithActiveTab(client, targetCRD, mainModel.height, mainModel.width, schemaTab)
			p = tea.NewProgram(instanceModel, tea.WithAltScreen(), tea.WithMouseCellMotion())
		}
	}

	_, err := p.Run()
	return err
}
