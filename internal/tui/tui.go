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

	tea "github.com/charmbracelet/bubbletea"

	"github.com/pehlicd/crd-wizard/internal/clustermanager"
)

// Start initializes and runs the Bubble Tea TUI.
func Start(clusterMgr *clustermanager.ClusterManager, crdName string, kind string) error {
	// Get the default client to start with
	client := clusterMgr.GetDefaultClient()
	if client == nil {
		return fmt.Errorf("no clusters available")
	}

	// Pass the cluster manager and current client to the main model constructor.
	mainModel := newMainModel(clusterMgr, client, crdName, kind)
	p := tea.NewProgram(mainModel, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}
