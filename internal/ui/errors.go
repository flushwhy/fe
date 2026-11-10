package ui

import (
	"errors"
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
)

// CommandError adds a concise explanation and optional recovery hint to a cause.
type CommandError struct {
	Message string
	Hint    string
	Cause   error
}

func (e *CommandError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return "command failed"
}

func (e *CommandError) Unwrap() error { return e.Cause }

// WriteError prints a consistent, readable error panel to w.
func WriteError(w io.Writer, err error) {
	if err == nil {
		return
	}

	var commandErr *CommandError
	body := err.Error()
	if errors.As(err, &commandErr) {
		body = commandErr.Message
		if body == "" {
			body = commandErr.Error()
		}
		if commandErr.Cause != nil && commandErr.Cause.Error() != body {
			body += "\n" + StyleDim.Render("Cause: "+commandErr.Cause.Error())
		}
		if commandErr.Hint != "" {
			body += "\n\n" + StyleWarn.Render("Try: "+commandErr.Hint)
		}
	}

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorRed).
		Padding(0, 1).
		Render(StyleError.Render("✗ Fe could not complete that command") + "\n\n" + body)
	fmt.Fprintln(w, panel)
}
