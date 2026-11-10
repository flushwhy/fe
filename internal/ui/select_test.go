package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSelectModelNavigationAndSelection(t *testing.T) {
	model := selectModel{title: "Choose", options: []string{"go", "odin", "zig"}}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(selectModel)
	if model.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", model.cursor)
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(selectModel)
	if model.selected != "odin" {
		t.Fatalf("selected = %q, want odin", model.selected)
	}
	if cmd == nil {
		t.Fatal("selecting an option should quit the prompt")
	}
}

func TestSelectModelViewShowsOptionsAndControls(t *testing.T) {
	view := (selectModel{title: "Choose a language", options: []string{"go", "odin"}}).View()
	for _, want := range []string{"Choose a language", "go", "odin", "Enter to select"} {
		if !strings.Contains(view, want) {
			t.Errorf("view does not contain %q", want)
		}
	}
}
