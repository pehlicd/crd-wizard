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
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pehlicd/crd-wizard/internal/k8s"
)

// Start initializes and runs the Bubble Tea TUI.
func Start(client *k8s.Client, crdName string, kind string) error {
	// Pass the flag values to the main model constructor.
	mainModel := newMainModel(client, crdName, kind)
	p := tea.NewProgram(mainModel, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err

}
