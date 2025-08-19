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
	var initialModel tea.Model

	// Only search for a specific CRD if at least one flag is provided.
	if crdName != "" || kind != "" {
		allCRDs, err := client.GetCRDs(context.Background())
		if err != nil {
			return fmt.Errorf("failed to get CRDs: %w", err)
		}

		var targetCRD models.CRD
		var found bool

		for _, crd := range allCRDs {
			matchesName := crdName != "" && crd.Name == crdName
			matchesKind := kind != "" && crd.Kind == kind

			// If both flags are set, both must match.
			if crdName != "" && kind != "" {
				if matchesName && matchesKind {
					targetCRD = crd
					found = true
					break
				}
				// If only one flag is set, either can match.
			} else {
				if matchesName || matchesKind {
					targetCRD = crd
					found = true
					break
				}
			}
		}

		// Check if crd is found
		if !found {
			return fmt.Errorf("could not find a matching CRD for name=%q and kind=%q", crdName, kind)
		}

		initialModel = newInstanceListModelWithActiveTab(client, targetCRD, 0, 0, schemaTab)
	}

	// If no specific CRD was requested, use the main model.
	if initialModel == nil {
		initialModel = newMainModel(client)
	}

	p := tea.NewProgram(initialModel, tea.WithAltScreen(), tea.WithMouseCellMotion())

	_, err := p.Run()
	return err
}
