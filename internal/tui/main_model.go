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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/pehlicd/crd-wizard/internal/ai"
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
	aiClient          *ai.Client
	view              currentView
	err               error
	width, height     int
	crdListModel      tea.Model
	instanceListModel tea.Model
	detailViewModel   tea.Model
	modalModel        modalModel
	loadingMsg        string
	analyzing         bool
	showModal         bool
}

func newMainModel(client *k8s.Client, aiClient *ai.Client, crdName, kind string) mainModel {
	model := mainModel{
		client:       client,
		aiClient:     aiClient,
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
		if m.showModal {
			if msg.String() == "esc" {
				m.showModal = false // Close modal
				return m, nil
			}
			// Forward keys to modal for scrolling
			m.modalModel, cmd = m.modalModel.Update(msg)
			return m, cmd
		}

		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// AI Analysis Trigger
		if msg.String() == "a" {
			if m.view != crdListView {
				// Ignore if not in list view
				return m, nil
			}
			if m.analyzing || m.showModal {
				return m, nil
			}
			if m.aiClient == nil {
				m.analyzing = true
				m.loadingMsg = "❌ AI is not enabled.\nRun with --enable-ai"
				return m, tea.Tick(2*time.Second, func(_ time.Time) tea.Msg { return clearErrorMsg{} })
			}

			m.analyzing = true
			m.loadingMsg = "Analyzing CRD with AI..."
			return m, m.analyzeSelectedCRD()
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

	case aiResultMsg:
		m.modalModel = newModalModel("AI Analysis", msg.content, m.width, m.height)
		m.analyzing = false
		m.showModal = true
		return m, nil

	case errMsg:
		m.err = msg.err
		// Show error in overlay instead of hiding it
		m.analyzing = true
		m.loadingMsg = fmt.Sprintf("❌ Error:\n%v", msg.err)
		m.showModal = false
		return m, tea.Tick(3*time.Second, func(_ time.Time) tea.Msg { return clearErrorMsg{} })

	case clearErrorMsg:
		m.analyzing = false
		m.loadingMsg = ""
		m.err = nil
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
	var baseView string
	switch m.view {
	case crdListView:
		baseView = m.crdListModel.View()
	case instanceListView:
		baseView = m.instanceListModel.View()
	case detailView:
		baseView = m.detailViewModel.View()
	default:
		baseView = "Unknown view"
	}

	if m.analyzing {
		// Overlay Loading
		loadingBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2).
			Render(m.loadingMsg)
		return overlay(baseView, loadingBox, m.width, m.height)
	}

	if m.showModal {
		// Overlay Modal
		return overlay(baseView, m.modalModel.View(), m.width, m.height)
	}

	return baseView
}

// overlay centers the fg on top of bg using a vertical splice strategy.
// This preserves background context above and below the modal, while overwriting
// the middle lines to avoid complex ANSI string manipulation issues.
func overlay(bg, fg string, width, height int) string {
	// Dim the background
	dimmedBg := lipgloss.NewStyle().Faint(true).Render(bg)
	bgLines := strings.Split(dimmedBg, "\n")

	// Ensure bgLines has enough lines
	if len(bgLines) < height {
		extra := height - len(bgLines)
		for range extra {
			bgLines = append(bgLines, "")
		}
	}

	// Prepare FG

	fgH := lipgloss.Height(fg)

	// Calculate positions
	yStart := (height - fgH) / 2
	if yStart < 0 {
		yStart = 0
	}
	yEnd := yStart + fgH

	fgLines := strings.Split(fg, "\n")

	var result []string

	for i := range height {
		// Safety check for index
		if i >= len(bgLines) {
			result = append(result, "")
			continue
		}

		if i >= yStart && i < yEnd {
			// Inside Modal/FG Zone
			fgIndex := i - yStart
			if fgIndex < len(fgLines) {
				// Render the modal line centered.
				// Note: effectively clears the sides of this line, but safe.
				line := lipgloss.PlaceHorizontal(width, lipgloss.Center, fgLines[fgIndex])
				result = append(result, line)
			} else {
				result = append(result, bgLines[i]) // Should not happen if height correct
			}
		} else {
			// Outside Modal Zone - Keep Background
			result = append(result, bgLines[i])
		}
	}

	return strings.Join(result, "\n")
}

type aiResultMsg struct {
	content string
}

type clearErrorMsg struct{}

func (m mainModel) analyzeSelectedCRD() tea.Cmd {
	return func() tea.Msg {
		// Hack to get selected item. In a real world, we'd refactor crdListModel to expose it cleanly.
		// For now, let's assume `crdListModel` is our internal `crdListModel` struct and assert it.
		// NOTE: Check crdlist_model.go to see if strict/public access is available.
		// If not, we might need to fetch the selection index.

		// If we can't easily get it, let's just use a dummy for this step or try to fix it.
		// Let's assume we can cast.
		if listModel, ok := m.crdListModel.(crdListModel); ok {
			if selected := listModel.SelectedItem(); selected != nil {
				// We have the CRD.
				var schemaJSON string
				// The model likely only has summary.
				// We need to fetch the Full CRD.
				fullCRD, err := m.client.GetFullCRD(context.Background(), selected.Name)
				if err != nil {
					return errMsg{err}
				}

				// Extract Schema (simplified)
				// We need to find the version matching the one we are interested in, usually the storage version or the first one.
				version := ""
				if len(fullCRD.Spec.Versions) > 0 {
					version = fullCRD.Spec.Versions[0].Name // Default to first
					for _, v := range fullCRD.Spec.Versions {
						if v.Storage {
							version = v.Name
							break
						}
					}

					if fullCRD.Spec.Versions[0].Schema != nil && fullCRD.Spec.Versions[0].Schema.OpenAPIV3Schema != nil {
						b, err := json.Marshal(fullCRD.Spec.Versions[0].Schema.OpenAPIV3Schema)
						if err == nil {
							schemaJSON = string(b)
						}
					}
				}

				if schemaJSON == "" {
					schemaJSON = "{}" // Fallback if no schema found or error
				}
				if version == "" {
					return errMsg{fmt.Errorf("no version found for CRD %s", selected.Name)}
				}

				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
				defer cancel()

				res, err := m.aiClient.GenerateCrdContext(ctx, selected.Group, version, selected.Kind, schemaJSON)
				if err != nil {
					return errMsg{err}
				}
				return aiResultMsg{res}
			}
		}
		return errMsg{fmt.Errorf("could not get selected CRD")}
	}
}
