package ui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func TestInputModelAcceptsCurrentValue(t *testing.T) {
	field := textinput.New()
	field.SetValue("my-project")
	model := inputModel{title: "Project name", field: field}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(inputModel)
	if model.cancelled {
		t.Fatal("accepted input should not be marked cancelled")
	}
	if model.field.Value() != "my-project" {
		t.Fatalf("value = %q, want my-project", model.field.Value())
	}
	if cmd == nil {
		t.Fatal("accepting input should quit the prompt")
	}
}

func TestInputModelCanBeCancelled(t *testing.T) {
	model := inputModel{field: textinput.New()}
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(inputModel)
	if !model.cancelled {
		t.Fatal("escape should cancel input")
	}
	if cmd == nil {
		t.Fatal("cancelling input should quit the prompt")
	}
}
