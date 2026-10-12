package cmd

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTUIControlCStopsRunningCommand(t *testing.T) {
	cancelled := false
	model := tuiModel{
		state: stateRunning,
		runCancel: func() {
			cancelled = true
		},
	}

	updated, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	got := updated.(tuiModel)
	if !cancelled {
		t.Fatal("expected ctrl+c to cancel the running command")
	}
	if cmd != nil {
		t.Fatal("expected ctrl+c to keep the TUI open instead of quitting")
	}
	if got.state != stateRunning {
		t.Fatalf("expected result state to remain running until command exits, got %v", got.state)
	}
}

func TestTUIControlCQuitsWhenIdle(t *testing.T) {
	model := tuiModel{state: stateMenu}
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected ctrl+c to quit when no command is running")
	}
}
