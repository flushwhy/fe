package cmd

import (
	"testing"

	bubbletea "github.com/charmbracelet/bubbletea"
	"github.com/flushwhy/fe/internal/o3de"
)

func TestO3DEFinderMovesAndSelectsEntry(t *testing.T) {
	entries := []o3de.Entry{
		{Kind: "engine", Name: "Engine A", Path: "/engines/a"},
		{Kind: "engine", Name: "Engine B", Path: "/engines/b"},
	}
	model := newO3DEFinderModel("engines", entries)

	updated, _ := model.Update(bubbletea.KeyMsg{Type: bubbletea.KeyDown})
	model = updated.(o3deFinderModel)
	if model.cursor != 1 {
		t.Fatalf("expected cursor at second entry, got %d", model.cursor)
	}

	updated, cmd := model.Update(bubbletea.KeyMsg{Type: bubbletea.KeyEnter})
	model = updated.(o3deFinderModel)
	if cmd == nil {
		t.Fatal("expected enter to quit the picker")
	}
	if model.selected == nil || model.selected.Path != "/engines/b" {
		t.Fatalf("expected second entry to be selected, got %#v", model.selected)
	}
}

func TestO3DEFinderClampsCursorAtBounds(t *testing.T) {
	model := newO3DEFinderModel("engines", []o3de.Entry{
		{Kind: "engine", Name: "Only engine", Path: "/engine"},
	})
	updated, _ := model.Update(bubbletea.KeyMsg{Type: bubbletea.KeyUp})
	model = updated.(o3deFinderModel)
	if model.cursor != 0 {
		t.Fatalf("cursor moved before first entry: %d", model.cursor)
	}
	updated, _ = model.Update(bubbletea.KeyMsg{Type: bubbletea.KeyDown})
	model = updated.(o3deFinderModel)
	if model.cursor != 0 {
		t.Fatalf("cursor moved after last entry: %d", model.cursor)
	}
}
