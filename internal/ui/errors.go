package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

// PrintError renders a consistent command failure with a useful help path.
func PrintError(command string, err error) {
	title := lipgloss.NewStyle().Bold(true).Foreground(ColorRed)
	body := lipgloss.NewStyle().PaddingLeft(2)
	hint := lipgloss.NewStyle().Foreground(ColorGray)

	if command == "" {
		command = "fe"
	}
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, title.Render("✗ Fe could not complete the command"))
	fmt.Fprintln(os.Stderr, body.Render(err.Error()))
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "%s %s\n", hint.Render("Help:"), hint.Render(command+" --help"))
}
