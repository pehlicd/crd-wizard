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
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// keyRune builds the tea.KeyMsg a real keypress of r would produce.
func keyRune(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

// Regression test for #149: "a" and "c" are global hotkeys (AI analysis,
// cluster selector) everywhere except while crdListModel's filter text
// input has focus, where every key belongs to the filter box instead.
// clusterManager and aiClient are deliberately left nil: the "c" and "a"
// handlers dereference them as soon as either branch is entered, so a
// regression that lets the hotkey through while filtering panics here
// instead of silently passing.
func TestGlobalHotkeysDoNotFireWhileFiltering(t *testing.T) {
	t.Run("c does not open the cluster selector", func(t *testing.T) {
		m := mainModel{
			view:         crdListView,
			crdListModel: crdListModel{filtering: true},
		}

		updated, _ := m.Update(keyRune('c'))

		got := updated.(mainModel)
		if got.view != crdListView {
			t.Fatalf("view = %v, want unchanged crdListView; the hotkey fired while filtering", got.view)
		}
	})

	t.Run("a does not trigger AI analysis", func(t *testing.T) {
		m := mainModel{
			view:         crdListView,
			crdListModel: crdListModel{filtering: true},
		}

		updated, _ := m.Update(keyRune('a'))

		got := updated.(mainModel)
		if got.analyzing {
			t.Fatalf("analyzing = true, want false; the hotkey fired while filtering")
		}
	})
}

// crdListFiltering must not panic before crdListModel is set (e.g. the
// startup WindowSizeMsg race), and must not mistake some other tea.Model
// for it.
func TestCrdListFiltering(t *testing.T) {
	t.Run("false when crdListModel is nil", func(t *testing.T) {
		m := mainModel{}
		if m.crdListFiltering() {
			t.Fatal("expected false with no crdListModel set")
		}
	})

	t.Run("reflects the underlying model's state", func(t *testing.T) {
		m := mainModel{crdListModel: crdListModel{filtering: true}}
		if !m.crdListFiltering() {
			t.Fatal("expected true when the underlying crdListModel is filtering")
		}
	})
}
