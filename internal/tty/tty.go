package tty

import "golang.org/x/term"

// IsInteractive returns true when stdout is a real terminal (not a pipe or
// CI runner). Commands should only render the Bubble Tea TUI when this is true.
func IsInteractive() bool {
	return term.IsTerminal(1) // fd 1 = stdout
}
