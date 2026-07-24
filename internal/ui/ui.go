// Package ui provides shared terminal output helpers used across all fe commands.
// When stdout is not a TTY (CI), spinners degrade to plain log lines automatically.
package ui

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/flushwhy/fe/internal/tty"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	ColorPurple = lipgloss.Color("#7C5FD6")
	ColorGreen  = lipgloss.Color("#3DDC84")
	ColorRed    = lipgloss.Color("#FF5F57")
	ColorYellow = lipgloss.Color("#FFBE2E")
	ColorGray   = lipgloss.Color("#6B7280")
	ColorWhite  = lipgloss.Color("#FAFAFA")

	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorPurple).
			Padding(0, 1)

	StyleSuccess = lipgloss.NewStyle().Foreground(ColorGreen).Bold(true)
	StyleError   = lipgloss.NewStyle().Foreground(ColorRed).Bold(true)
	StyleWarn    = lipgloss.NewStyle().Foreground(ColorYellow).Bold(true)
	StyleDim     = lipgloss.NewStyle().Foreground(ColorGray)
	StyleBold    = lipgloss.NewStyle().Bold(true)

	StyleStep = lipgloss.NewStyle().
			Foreground(ColorPurple).
			Bold(true)

	StyleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPurple).
			Padding(0, 1)
)

// ── Icons ─────────────────────────────────────────────────────────────────────

const (
	IconOK   = "✓"
	IconFail = "✗"
	IconWarn = "⚠"
	IconArrow = "→"
	IconRocket = "🚀"
	IconSpark  = "✨"
)

// OK prints a green success line. Always prints regardless of TTY.
func OK(format string, args ...interface{}) {
	fmt.Printf("  %s %s\n", StyleSuccess.Render(IconOK), fmt.Sprintf(format, args...))
}

// Fail prints a red error line.
func Fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "  %s %s\n", StyleError.Render(IconFail), fmt.Sprintf(format, args...))
}

// Warn prints a yellow warning line.
func Warn(format string, args ...interface{}) {
	fmt.Printf("  %s %s\n", StyleWarn.Render(IconWarn), fmt.Sprintf(format, args...))
}

// Dim prints a muted info line.
func Dim(format string, args ...interface{}) {
	fmt.Printf("  %s\n", StyleDim.Render(fmt.Sprintf(format, args...)))
}

// Title prints a bold header.
func Title(format string, args ...interface{}) {
	fmt.Printf("\n%s\n\n", StyleTitle.Render(fmt.Sprintf(format, args...)))
}

// Step prints a numbered step header.
func Step(n, total int, label string) {
	fmt.Printf("%s\n", StyleStep.Render(fmt.Sprintf("── Step %d/%d: %s", n, total, label)))
}

// Hint prints a styled next-steps block.
func Hint(lines ...string) {
	fmt.Println()
	for _, l := range lines {
		fmt.Printf("  %s %s\n", StyleDim.Render(IconArrow), l)
	}
	fmt.Println()
}

// ── Spinner ───────────────────────────────────────────────────────────────────

// RunSpinner runs fn in the background while showing a spinner in the terminal.
// In non-TTY environments (CI) it just prints the label and runs fn directly.
// Returns the error from fn.
func RunSpinner(label string, fn func() error) error {
	if !tty.IsInteractive() {
		fmt.Printf("  %s ...\n", label)
		return fn()
	}

	m := spinnerModel{
		spinner: newSpinner(),
		label:   label,
	}

	var runErr error

	// Run fn in a goroutine; signal completion via a channel.
	done := make(chan error, 1)
	go func() { done <- fn() }()

	p := tea.NewProgram(m)

	// Poll for completion and quit the spinner program when done.
	go func() {
		runErr = <-done
		p.Send(spinnerDoneMsg{err: runErr})
	}()

	if _, err := p.Run(); err != nil {
		// Program itself failed (not fn) — still wait for fn.
		return err
	}

	return runErr
}

// ── Spinner Bubble Tea model ───────────────────────────────────────────────────

type spinnerDoneMsg struct{ err error }

type spinnerModel struct {
	spinner spinner.Model
	label   string
	done    bool
	err     error
}

func newSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPurple)
	return s
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinnerDoneMsg:
		m.done = true
		m.err = msg.err
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		if m.err != nil {
			return fmt.Sprintf("  %s %s\n", StyleError.Render(IconFail), m.label)
		}
		return fmt.Sprintf("  %s %s\n", StyleSuccess.Render(IconOK), m.label)
	}
	return fmt.Sprintf("  %s %s\n", m.spinner.View(), m.label)
}

// ── Multi-step spinner ────────────────────────────────────────────────────────

// Step is a single unit of work for RunSteps.
type StepTask struct {
	Label string
	Run   func() error
}

// RunSteps executes a list of StepTasks sequentially, showing a spinner per step.
// Stops at the first error.
func RunSteps(steps []StepTask) error {
	for i, step := range steps {
		label := fmt.Sprintf("[%d/%d] %s", i+1, len(steps), step.Label)
		if err := RunSpinner(label, step.Run); err != nil {
			Fail("%s failed: %v", step.Label, err)
			return err
		}
	}
	return nil
}

// ── Confirm prompt ────────────────────────────────────────────────────────────

// Confirm shows a y/N prompt and returns true if the user typed y/yes.
// In non-TTY mode it always returns true (CI-safe — assume yes).
func Confirm(prompt string) bool {
	if !tty.IsInteractive() {
		return true
	}

	fmt.Printf("  %s [y/N] ", StyleBold.Render(prompt))

	var input string
	fmt.Scanln(&input)
	return input == "y" || input == "Y" || input == "yes"
}

// ── Progress bar (simple, no bubbletea) ──────────────────────────────────────

// Progress renders a simple ASCII progress bar to stdout.
func Progress(current, total int, label string) {
	width := 30
	filled := 0
	if total > 0 {
		filled = (current * width) / total
	}
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	pct := 0
	if total > 0 {
		pct = (current * 100) / total
	}
	fmt.Printf("\r  %s %s %d%%", lipgloss.NewStyle().Foreground(ColorPurple).Render(bar), StyleDim.Render(label), pct)
	if current >= total {
		fmt.Println()
	}
}

// Elapsed prints a dim elapsed time string. Call with time.Now() before the op.
func Elapsed(start time.Time) {
	Dim("done in %s", time.Since(start).Round(time.Millisecond))
}
