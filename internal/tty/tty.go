package tty

import "golang.org/x/term"

// IsInteractive returns true only when both stdin and stdout are terminals.
// Bubble Tea needs keyboard input as well as a terminal to render into.
func IsInteractive() bool {
	return term.IsTerminal(0) && term.IsTerminal(1)
}
