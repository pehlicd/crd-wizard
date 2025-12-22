package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type modalModel struct {
	content  string
	rendered string
	title    string
	viewport viewport.Model
	width    int
	height   int
}

func newModalModel(title, content string, width, height int) modalModel {
	// Calculate modal dimensions (e.g., 80% of screen)
	modalWidth := int(float64(width) * 0.8)
	modalHeight := int(float64(height) * 0.8)

	// Render Markdown
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(modalWidth-4),
	)
	rendered, err := r.Render(content)
	if err != nil {
		rendered = content // Fallback
	}

	vp := viewport.New(modalWidth, modalHeight-4) // -4 for headers/borders
	vp.SetContent(rendered)

	return modalModel{
		content:  content,
		rendered: rendered,
		title:    title,
		viewport: vp,
		width:    width,
		height:   height,
	}
}

func (m modalModel) Init() tea.Cmd {
	return nil
}

func (m modalModel) Update(msg tea.Msg) (modalModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m modalModel) View() string {
	// Styling
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(m.viewport.Width + 2).
		Height(m.viewport.Height + 2)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("62")).
		Bold(true).
		Padding(0, 1)

	// Combine components
	header := titleStyle.Render(m.title)
	body := m.viewport.View()
	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")).Render("Esc to Close")

	modalContent := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
	modal := borderStyle.Render(modalContent)

	return modal
}
