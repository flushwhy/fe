package ui

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/flushwhy/fe/internal/tty"
)

var ErrCancelled = errors.New("prompt cancelled")

// PrintCancelled reports a user-cancelled prompt without treating it as a failure.
func PrintCancelled() {
	fmt.Fprintln(os.Stderr, StyleDim.Render("Cancelled."))
}

// PrintError renders a consistent command failure with a useful help path.
func PrintError(command string, err error) {
	if command == "" {
		command = "fe"
	}
	help := command + " --help"
	if !tty.IsInteractive() {
		fmt.Fprintf(os.Stderr, "Error: %v\nHelp:  %s\n", err, help)
		return
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(ColorRed)
	body := lipgloss.NewStyle().PaddingLeft(2)
	hint := lipgloss.NewStyle().Foreground(ColorGray)
	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorRed).
		Padding(1, 2)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title.Render("✗ Command failed"),
		body.Render(err.Error()),
		hint.Render("Try: "+help),
	)
	fmt.Fprintln(os.Stderr, panel.Render(content))
}
