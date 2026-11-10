package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/flushwhy/fe/internal/tty"
)

// Input asks for a single line of text, prefilled with the supplied value.
func Input(title, initial string) (string, error) {
	if !tty.IsInteractive() {
		return "", fmt.Errorf("interactive input requires a terminal")
	}

	field := textinput.New()
	field.CharLimit = 128
	field.Width = 40
	field.SetValue(initial)
	field.Focus()

	result, err := tea.NewProgram(inputModel{title: title, field: field}).Run()
	if err != nil {
		return "", fmt.Errorf("showing input: %w", err)
	}
	model, ok := result.(inputModel)
	if !ok || model.cancelled {
		return "", ErrCancelled
	}
	return strings.TrimSpace(model.field.Value()), nil
}

type inputModel struct {
	title     string
	field     textinput.Model
	cancelled bool
}

func (m inputModel) Init() tea.Cmd { return textinput.Blink }

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter":
			return m, tea.Quit
		case "esc", "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.field, cmd = m.field.Update(msg)
	return m, cmd
}

func (m inputModel) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(ColorPurple)
	field := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPurple).
		Padding(0, 1)
	help := lipgloss.NewStyle().Foreground(ColorGray)

	return title.Render(m.title) + "\n\n" +
		field.Render(m.field.View()) + "\n\n" +
		help.Render("Enter to confirm · Esc to cancel")
}
