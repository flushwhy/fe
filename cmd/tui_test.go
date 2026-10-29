package cmd

import (
	"strings"
	"testing"

	"github.com/flushwhy/fe/internal/config"

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

func TestCompactTUIKeepsMenuWithinSmallTerminal(t *testing.T) {
	m := newTuiModel(&config.Config{}, nil)
	m.width = 60
	m.height = 12
	m.items = []menuItem{
		{label: "First", desc: "First action", command: "first"},
		{label: "Second", desc: "Second action", command: "second"},
		{label: "Third", desc: "Third action", command: "third"},
		{label: "Fourth", desc: "Fourth action", command: "fourth"},
	}
	m.cursor = 3

	view := m.View()
	if !strings.Contains(view, "fe — dev middleware") {
		t.Fatal("compact view should render the banner")
	}
	if !strings.Contains(view, "Fourth") || !strings.Contains(view, "Fourth action") {
		t.Fatal("compact view should show the selected item and its description")
	}
	if !strings.Contains(view, "items 3–4 of 4") {
		t.Fatalf("compact view should indicate the visible menu range; got:\\n%s", view)
	}
	if strings.Contains(view, "First action") {
		t.Fatal("compact view should not render descriptions for unselected items")
	}
}

func TestCompactTUITruncatesResultToAvailableHeight(t *testing.T) {
	m := newTuiModel(&config.Config{}, nil)
	m.width = 60
	m.height = 10
	m.state = stateResult
	m.lastOutput = "line 1\\nline 2\\nline 3\\nline 4\\nline 5\\nline 6"

	view := m.View()
	if !strings.Contains(view, "line 6") {
		t.Fatal("compact result should preserve the latest output")
	}
	if strings.Contains(view, "line 1") {
		t.Fatal("compact result should drop older output when it exceeds the available height")
	}
}
