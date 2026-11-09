package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/flushwhy/fe/internal/o3de"
	"github.com/flushwhy/fe/internal/tty"
	"github.com/flushwhy/fe/internal/ui"
)

var (
	o3deBuildErrorTitle = lipgloss.NewStyle().Bold(true).Foreground(ui.ColorRed)
	o3deBuildErrorHint  = lipgloss.NewStyle().Foreground(ui.ColorYellow)
	o3deBuildErrorHelp  = lipgloss.NewStyle().Foreground(ui.ColorGray)
)

type o3deBuildFailureModel struct {
	buildDir string
	buildErr error
	hints    []string
	logs     viewport.Model
}

func newO3DEBuildFailureModel(buildDir string, buildErr error, output string) o3deBuildFailureModel {
	logs := viewport.New(80, 12)
	if strings.TrimSpace(output) == "" {
		output = "CMake did not produce any output."
	}
	logs.SetContent(strings.TrimRight(output, "\n"))
	logs.GotoBottom()

	return o3deBuildFailureModel{
		buildDir: buildDir,
		buildErr: buildErr,
		hints:    o3de.SummarizeBuildFailure(output),
		logs:     logs,
	}
}

func (m o3deBuildFailureModel) Init() tea.Cmd { return nil }

func (m o3deBuildFailureModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.logs.Width = msg.Width - 4
		if m.logs.Width < 20 {
			m.logs.Width = 20
		}
		m.logs.Height = msg.Height - 9
		if m.logs.Height < 3 {
			m.logs.Height = 3
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.logs, cmd = m.logs.Update(msg)
	return m, cmd
}

func (m o3deBuildFailureModel) View() string {
	var b strings.Builder
	b.WriteString(o3deBuildErrorTitle.Render("O3DE build failed"))
	b.WriteString("\n")
	b.WriteString(ui.StyleDim.Render("Build directory: " + m.buildDir))
	b.WriteString("\n")
	b.WriteString(ui.StyleDim.Render(m.buildErr.Error()))
	b.WriteString("\n\n")
	b.WriteString(ui.StyleBold.Render("What to check"))
	b.WriteString("\n")
	for _, hint := range m.hints {
		b.WriteString(o3deBuildErrorHint.Render("  → " + hint))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(ui.StyleBold.Render("Build output"))
	b.WriteString("\n")
	b.WriteString(m.logs.View())
	b.WriteString("\n")
	b.WriteString(o3deBuildErrorHelp.Render("↑/↓ scroll • PgUp/PgDn page • Home/End jump • q/Esc close"))
	return b.String()
}

func showO3DEBuildFailure(buildDir string, buildErr error, output string) {
	if !tty.IsInteractive() {
		printO3DEBuildFailure(buildErr, output)
		return
	}

	model := newO3DEBuildFailureModel(buildDir, buildErr, output)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		printO3DEBuildFailure(buildErr, output)
	}
}

func printO3DEBuildFailure(buildErr error, output string) {
	ui.Fail("CMake build failed: %v", buildErr)
	if strings.TrimSpace(output) != "" {
		fmt.Println()
		fmt.Println(ui.StyleBold.Render("Build output"))
		fmt.Print(output)
		if !strings.HasSuffix(output, "\n") {
			fmt.Println()
		}
	}
	fmt.Println()
	fmt.Println(ui.StyleBold.Render("What to check"))
	for _, hint := range o3de.SummarizeBuildFailure(output) {
		fmt.Printf("  %s %s\n", ui.StyleWarn.Render(ui.IconArrow), hint)
	}
}
