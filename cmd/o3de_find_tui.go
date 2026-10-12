package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/flushwhy/fe/internal/o3de"
)

var (
	o3deFinderSelected = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	o3deFinderMuted    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	o3deFinderTitle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
)

type o3deFinderModel struct {
	kind     string
	entries  []o3de.Entry
	cursor   int
	offset   int
	height   int
	selected *o3de.Entry
	quitting bool
}

func newO3DEFinderModel(kind string, entries []o3de.Entry) o3deFinderModel {
	return o3deFinderModel{kind: kind, entries: entries, height: 12}
}

func (m o3deFinderModel) Init() bubbletea.Cmd { return nil }

func (m o3deFinderModel) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	switch msg := msg.(type) {
	case bubbletea.WindowSizeMsg:
		m.height = msg.Height - 10
		if m.height < 3 {
			m.height = 3
		}
	case bubbletea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.quitting = true
			return m, bubbletea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
		case "home", "g":
			m.cursor = 0
		case "end", "G":
			m.cursor = len(m.entries) - 1
		case "pgup":
			m.cursor -= m.height
			if m.cursor < 0 {
				m.cursor = 0
			}
		case "pgdown":
			m.cursor += m.height
			if m.cursor >= len(m.entries) {
				m.cursor = len(m.entries) - 1
			}
		case "enter":
			if len(m.entries) > 0 {
				entry := m.entries[m.cursor]
				m.selected = &entry
			}
			return m, bubbletea.Quit
		}
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.height {
		m.offset = m.cursor - m.height + 1
	}
	return m, nil
}

func (m o3deFinderModel) View() string {
	var b strings.Builder
	b.WriteString(o3deFinderTitle.Render(fmt.Sprintf("O3DE %s finder", m.kind)))
	b.WriteString("\nSelect an item to inspect its details; press Enter to choose.\n\n")
	end := m.offset + m.height
	if end > len(m.entries) {
		end = len(m.entries)
	}
	for i := m.offset; i < end; i++ {
		entry := m.entries[i]
		name := entry.Name
		if entry.Version != "" {
			name += "  (" + entry.Version + ")"
		}
		prefix := "  "
		if i == m.cursor {
			prefix = "› "
			name = o3deFinderSelected.Render(name)
		}
		b.WriteString(prefix + name + "\n")
	}
	if len(m.entries) > m.height {
		b.WriteString(o3deFinderMuted.Render(fmt.Sprintf("\nShowing %d-%d of %d", m.offset+1, end, len(m.entries))) + "\n")
	}
	if len(m.entries) > 0 {
		entry := m.entries[m.cursor]
		b.WriteString("\n" + o3deFinderTitle.Render("Details") + "\n")
		b.WriteString("Path:   " + entry.Path + "\n")
		b.WriteString("Source: " + entry.Source + "\n")
	}
	b.WriteString("\n" + o3deFinderMuted.Render("↑/↓ or j/k move • PgUp/PgDn page • Home/End jump • Enter select • q/Esc cancel"))
	return b.String()
}

func runO3DEFinder(kind string, entries []o3de.Entry) (*o3de.Entry, error) {
	model := newO3DEFinderModel(kind, entries)
	program := bubbletea.NewProgram(model)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("run O3DE finder: %w", err)
	}
	result, ok := finalModel.(o3deFinderModel)
	if !ok {
		return nil, fmt.Errorf("O3DE finder returned an unexpected model")
	}
	if result.quitting {
		return nil, nil
	}
	return result.selected, nil
}
