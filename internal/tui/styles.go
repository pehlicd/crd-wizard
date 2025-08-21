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

import "github.com/charmbracelet/lipgloss"

var (
	AppStyle         = lipgloss.NewStyle().Margin(1, 2)
	TitleStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true).Margin(0, 1)
	HelpStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0)
	ErrStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87")).Bold(true)
	HeaderStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true).Padding(0, 1).Border(lipgloss.NormalBorder(), false, false, true, false).BorderForeground(lipgloss.Color("#7D56F4"))
	CellStyle        = lipgloss.NewStyle().Padding(0, 1)
	SelectedStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F8F8F2")).Background(lipgloss.Color("#7D56F4"))
	TabStyle         = lipgloss.NewStyle().Padding(0, 1).MarginRight(2)
	ActiveTabStyle   = TabStyle.Foreground(lipgloss.Color("#F8F8F2")).Background(lipgloss.Color("#7D56F4")).Bold(true)
	InactiveTabStyle = TabStyle.Foreground(lipgloss.Color("241"))
)
