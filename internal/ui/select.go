package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/flushwhy/fe/internal/tty"
)

// Select displays a keyboard-driven choice menu in interactive terminals.
func Select(title string, options []string) (string, error) {
	if !tty.IsInteractive() {
		return "", fmt.Errorf("interactive selection requires a terminal")
	}
	if len(options) == 0 {
		return "", fmt.Errorf("no options available")
	}

	result, err := tea.NewProgram(selectModel{title: title, options: options}).Run()
	if err != nil {
		return "", fmt.Errorf("showing selection: %w", err)
	}
	model, ok := result.(selectModel)
	if !ok || model.selected == "" {
		return "", ErrCancelled
	}
	return model.selected, nil
}

type selectModel struct {
	title    string
	options  []string
	cursor   int
	selected string
}

func (m selectModel) Init() tea.Cmd { return nil }

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch key.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.options)-1 {
			m.cursor++
		}
	case "enter":
		m.selected = m.options[m.cursor]
		return m, tea.Quit
	case "esc", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m selectModel) View() string {
	var b strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(ColorPurple)
	selected := lipgloss.NewStyle().Bold(true).Foreground(ColorPurple)
	muted := lipgloss.NewStyle().Foreground(ColorGray)

	b.WriteString(title.Render(m.title))
	b.WriteString("\n\n")
	for i, option := range m.options {
		if i == m.cursor {
			b.WriteString(selected.Render("  › " + option))
		} else {
			b.WriteString(muted.Render("    " + option))
		}
		b.WriteByte('\n')
	}
	b.WriteString("\n")
	b.WriteString(muted.Render("↑/↓ or j/k to move · Enter to select · Esc to cancel"))
	return b.String()
}
