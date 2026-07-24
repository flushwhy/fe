package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/flushwhy/fe/internal/config"
	"github.com/flushwhy/fe/internal/tty"
	"github.com/flushwhy/fe/internal/ui"
	"github.com/spf13/cobra"
)

// ── Palette ───────────────────────────────────────────────────────────────────

var (
	purple    = lipgloss.Color("#7C5FD6")
	green     = lipgloss.Color("#3DDC84")
	red       = lipgloss.Color("#FF5F57")
	yellow    = lipgloss.Color("#FFBE2E")
	gray      = lipgloss.Color("#6B7280")
	lightGray = lipgloss.Color("#9CA3AF")
	white     = lipgloss.Color("#F9FAFB")
	dark      = lipgloss.Color("#1F2937")

	styleBanner = lipgloss.NewStyle().
			Bold(true).
			Foreground(white).
			Background(purple).
			Padding(0, 2).
			MarginBottom(1)

	styleTab = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(gray)

	styleTabActive = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(purple).
			Bold(true).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(purple)

	stylePane = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(purple).
			Padding(1, 2)

	styleSelected = lipgloss.NewStyle().
			Foreground(purple).
			Bold(true)

	styleCmd = lipgloss.NewStyle().
			Foreground(lightGray).
			Italic(true)

	styleDim = lipgloss.NewStyle().Foreground(gray)

	styleSuccess = lipgloss.NewStyle().Foreground(green).Bold(true)
	styleError   = lipgloss.NewStyle().Foreground(red).Bold(true)
	styleWarn    = lipgloss.NewStyle().Foreground(yellow)

	styleInput = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(purple).
			Padding(0, 1)

	styleLog = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(gray).
			Padding(0, 1).
			Foreground(lightGray)

	styleKey = lipgloss.NewStyle().
			Foreground(purple).
			Bold(true)
)

// ── Tabs ──────────────────────────────────────────────────────────────────────

type tabID int

const (
	tabProject tabID = iota
	tabShip
	tabDoctor
	tabConfig
)

var tabs = []struct {
	id    tabID
	label string
}{
	{tabProject, "  Project  "},
	{tabShip, "  Ship  "},
	{tabDoctor, "  Doctor  "},
	{tabConfig, "  Config  "},
}

// ── Menu items ────────────────────────────────────────────────────────────────

type menuItem struct {
	label   string
	desc    string
	command string
	args    []string
}

var projectItems = []menuItem{
	{"init odin", "Scaffold a new Odin project", "init", []string{"odin"}},
	{"init c", "Scaffold a new C project", "init", []string{"c"}},
	{"init cpp", "Scaffold a new C++ project", "init", []string{"cpp"}},
	{"init zig", "Scaffold a new Zig project", "init", []string{"zig"}},
	{"init go", "Scaffold a new Go project", "init", []string{"go"}},
	{"add clangd", "Wire up clangd for C/C++", "add", []string{"clangd"}},
	{"add ols", "Wire up OLS for Odin", "add", []string{"ols"}},
	{"add zls", "Wire up ZLS for Zig", "add", []string{"zls"}},
	{"add gopls", "Wire up gopls for Go", "add", []string{"gopls"}},
}

var shipItems = []menuItem{
	{"build", "Compile all platform targets", "build", nil},
	{"pack", "Pack sprites into spritesheet", "pack", nil},
	{"transcode", "Convert audio/video via ffmpeg", "transcode", nil},
	{"bmp", "Push builds to itch.io", "bmp", nil},
	{"release", "Full pipeline: build → pack → push", "release", nil},
	{"validate", "Pre-flight config and path check", "validate", nil},
}

// ── State machine ─────────────────────────────────────────────────────────────

type viewState int

const (
	stateMenu    viewState = iota // browsing a menu
	stateInput                    // filling in a text prompt
	stateRunning                  // command is running (spinner)
	stateResult                   // showing output after a run
)

// ── Messages ──────────────────────────────────────────────────────────────────

type cmdDoneMsg struct {
	output string
	err    error
	dur    time.Duration
}

type tickMsg time.Time

// ── Model ─────────────────────────────────────────────────────────────────────

type tuiModel struct {
	// layout
	width  int
	height int

	// tabs
	activeTab tabID

	// menu
	cursor int
	items  []menuItem

	// state machine
	state      viewState
	spinner    spinner.Model
	input      textinput.Model
	inputLabel string
	pending    menuItem // item waiting for input before running

	// result
	lastOutput string
	lastErr    error
	lastDur    time.Duration

	// config
	cfg    *config.Config
	cfgErr error

	// log lines shown in result pane
	logs []string
}

func newTuiModel(cfg *config.Config, cfgErr error) tuiModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(purple)

	ti := textinput.New()
	ti.CharLimit = 64
	ti.Width = 30

	m := tuiModel{
		activeTab: tabProject,
		items:     projectItems,
		state:     stateMenu,
		spinner:   sp,
		input:     ti,
		cfg:       cfg,
		cfgErr:    cfgErr,
	}
	return m
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m tuiModel) Init() tea.Cmd {
	return nil
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case spinner.TickMsg:
		if m.state == stateRunning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case cmdDoneMsg:
		m.state = stateResult
		m.lastOutput = msg.output
		m.lastErr = msg.err
		m.lastDur = msg.dur
		if msg.output != "" {
			m.logs = append(m.logs, msg.output)
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	if m.state == stateInput {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m tuiModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys
	switch msg.String() {
	case "ctrl+c", "q":
		if m.state != stateInput {
			return m, tea.Quit
		}
	case "esc":
		if m.state == stateResult || m.state == stateInput {
			m.state = stateMenu
			m.input.Blur()
			return m, nil
		}
	}

	switch m.state {
	case stateMenu:
		return m.handleMenuKey(msg)
	case stateInput:
		return m.handleInputKey(msg)
	case stateResult:
		if msg.String() == "enter" || msg.String() == " " {
			m.state = stateMenu
		}
	}

	return m, nil
}

func (m tuiModel) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}

	case "tab", "l":
		m.activeTab = (m.activeTab + 1) % tabID(len(tabs))
		m.cursor = 0
		m.items = itemsForTab(m.activeTab)

	case "shift+tab", "h":
		m.activeTab = (m.activeTab + tabID(len(tabs)) - 1) % tabID(len(tabs))
		m.cursor = 0
		m.items = itemsForTab(m.activeTab)

	case "1":
		m.activeTab, m.cursor = tabProject, 0
		m.items = projectItems
	case "2":
		m.activeTab, m.cursor = tabShip, 0
		m.items = shipItems
	case "3":
		m.activeTab, m.cursor = tabDoctor, 0
		m.items = nil
	case "4":
		m.activeTab, m.cursor = tabConfig, 0
		m.items = nil

	case "enter", " ":
		return m.activateItem()
	}

	return m, nil
}

func (m tuiModel) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		val := strings.TrimSpace(m.input.Value())
		if val == "" {
			return m, nil
		}
		m.input.Blur()
		return m.runItem(m.pending, val)
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m tuiModel) activateItem() (tuiModel, tea.Cmd) {
	if m.activeTab == tabDoctor {
		return m.runDoctor()
	}
	if len(m.items) == 0 {
		return m, nil
	}

	item := m.items[m.cursor]

	// Items that need a project name prompt
	needsName := item.command == "init"
	if needsName {
		m.pending = item
		m.inputLabel = fmt.Sprintf("Project name for '%s':", item.label)
		m.input.SetValue("")
		m.input.Placeholder = "my-project"
		m.input.Focus()
		m.state = stateInput
		return m, textinput.Blink
	}

	return m.runItem(item, "")
}

func (m tuiModel) runItem(item menuItem, extraArg string) (tuiModel, tea.Cmd) {
	m.state = stateRunning
	m.lastOutput = ""
	m.lastErr = nil

	feArgs := append([]string{item.command}, item.args...)
	if extraArg != "" {
		feArgs = append(feArgs, "--name", extraArg)
	}

	start := time.Now()

	return m, tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			// Re-run fe itself as a subprocess so the full command logic runs.
			self, _ := os.Executable()
			cmd := exec.Command(self, feArgs...)
			out, err := cmd.CombinedOutput()
			return cmdDoneMsg{
				output: string(out),
				err:    err,
				dur:    time.Since(start),
			}
		},
	)
}

func (m tuiModel) runDoctor() (tuiModel, tea.Cmd) {
	m.state = stateRunning
	start := time.Now()

	return m, tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			self, _ := os.Executable()
			cmd := exec.Command(self, "doctor")
			out, err := cmd.CombinedOutput()
			return cmdDoneMsg{output: string(out), err: err, dur: time.Since(start)}
		},
	)
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m tuiModel) View() string {
	if m.width == 0 {
		return ""
	}

	var b strings.Builder

	// Banner
	b.WriteString(styleBanner.Width(m.width - 2).Render("  fe  —  dev middleware"))
	b.WriteString("\n")

	// Tab bar
	b.WriteString(m.renderTabs())
	b.WriteString("\n\n")

	// Main content area split: left menu + right detail
	menuWidth := 34
	detailWidth := m.width - menuWidth - 6
	if detailWidth < 20 {
		detailWidth = 20
	}

	switch m.state {
	case stateRunning:
		b.WriteString(m.renderRunning(menuWidth, detailWidth))
	case stateResult:
		b.WriteString(m.renderResult(menuWidth, detailWidth))
	case stateInput:
		b.WriteString(m.renderInput(menuWidth, detailWidth))
	default:
		b.WriteString(m.renderMenu(menuWidth, detailWidth))
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	return b.String()
}

func (m tuiModel) renderTabs() string {
	var parts []string
	for _, t := range tabs {
		if t.id == m.activeTab {
			parts = append(parts, styleTabActive.Render(t.label))
		} else {
			parts = append(parts, styleTab.Render(t.label))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func (m tuiModel) renderMenu(menuW, detailW int) string {
	// Left: menu list
	var menu strings.Builder
	for i, item := range m.items {
		prefix := "  "
		style := lipgloss.NewStyle().Foreground(white)
		descStyle := styleDim

		if i == m.cursor {
			prefix = styleKey.Render("▶ ")
			style = styleSelected
			descStyle = lipgloss.NewStyle().Foreground(lightGray)
		}

		menu.WriteString(fmt.Sprintf("%s%s\n",
			prefix,
			style.Render(item.label),
		))
		menu.WriteString(fmt.Sprintf("   %s\n", descStyle.Render(item.desc)))
		menu.WriteString("\n")
	}

	// Special tab content
	if m.activeTab == tabDoctor {
		menu.WriteString(fmt.Sprintf("  %s\n\n", styleSelected.Render("Run doctor")))
		menu.WriteString(fmt.Sprintf("   %s\n", styleDim.Render("Check tools + LSP configs")))
		menu.WriteString("\n")
		menu.WriteString(styleDim.Render("  Press enter to run"))
	}

	if m.activeTab == tabConfig {
		menu.WriteString(m.renderConfigPane())
	}

	leftPane := stylePane.Width(menuW).Height(m.height - 10).Render(menu.String())

	// Right: detail / hints
	detail := m.renderDetail(detailW)
	rightPane := stylePane.Width(detailW).Height(m.height - 10).Render(detail)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, "  ", rightPane)
}

func (m tuiModel) renderDetail(w int) string {
	if m.activeTab == tabConfig || m.activeTab == tabDoctor {
		return ""
	}

	if len(m.items) == 0 || m.cursor >= len(m.items) {
		return styleDim.Render("No item selected")
	}

	item := m.items[m.cursor]
	var b strings.Builder

	b.WriteString(styleSelected.Render(item.label))
	b.WriteString("\n\n")
	b.WriteString(item.desc)
	b.WriteString("\n\n")
	b.WriteString(styleDim.Render("Command:"))
	b.WriteString("\n")

	cmdStr := "fe " + item.command
	for _, a := range item.args {
		cmdStr += " " + a
	}
	b.WriteString(styleCmd.Render("  " + cmdStr))
	b.WriteString("\n\n")

	// Show config summary if available
	if m.cfg != nil {
		b.WriteString(styleDim.Render("Active config:"))
		b.WriteString("\n")
		for _, line := range m.cfg.SummaryLines() {
			b.WriteString(styleDim.Render(line))
			b.WriteString("\n")
		}
	}

	if len(m.logs) > 0 {
		b.WriteString("\n")
		b.WriteString(styleDim.Render("Last run:"))
		b.WriteString("\n")
		last := m.logs[len(m.logs)-1]
		lines := strings.Split(last, "\n")
		if len(lines) > 8 {
			lines = lines[len(lines)-8:]
		}
		b.WriteString(styleLog.Width(w - 4).Render(strings.Join(lines, "\n")))
	}

	return b.String()
}

func (m tuiModel) renderRunning(menuW, detailW int) string {
	content := fmt.Sprintf("\n\n  %s  Running...\n\n",
		m.spinner.View(),
	)
	if len(m.items) > 0 && m.cursor < len(m.items) {
		content += fmt.Sprintf("  %s\n", styleDim.Render(m.items[m.cursor].desc))
	}
	content += fmt.Sprintf("\n  %s", styleDim.Render("Press ctrl+c to cancel"))

	box := stylePane.Width(menuW + detailW + 2).Height(m.height - 10).Render(content)
	return box
}

func (m tuiModel) renderResult(menuW, detailW int) string {
	var header string
	if m.lastErr != nil {
		header = styleError.Render("✗  Command failed") +
			styleDim.Render(fmt.Sprintf("  (%s)", m.lastDur.Round(time.Millisecond)))
	} else {
		header = styleSuccess.Render("✓  Done") +
			styleDim.Render(fmt.Sprintf("  (%s)", m.lastDur.Round(time.Millisecond)))
	}

	output := m.lastOutput
	if output == "" {
		output = "(no output)"
	}

	// Trim to last N lines so it fits
	lines := strings.Split(strings.TrimSpace(output), "\n")
	maxLines := m.height - 16
	if maxLines < 4 {
		maxLines = 4
	}
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}

	content := header + "\n\n" +
		styleLog.Width(menuW+detailW-2).Render(strings.Join(lines, "\n")) +
		"\n\n" + styleDim.Render("  Press enter or esc to go back")

	return stylePane.Width(menuW + detailW + 2).Height(m.height - 10).Render(content)
}

func (m tuiModel) renderInput(menuW, detailW int) string {
	content := fmt.Sprintf("\n  %s\n\n  %s\n\n  %s",
		styleSelected.Render(m.inputLabel),
		styleInput.Render(m.input.View()),
		styleDim.Render("enter to confirm · esc to cancel"),
	)
	return stylePane.Width(menuW + detailW + 2).Height(m.height - 10).Render(content)
}

func (m tuiModel) renderConfigPane() string {
	if m.cfgErr != nil {
		return styleError.Render("Config error: "+m.cfgErr.Error()) +
			"\n\n" + styleDim.Render("Create a .fe.yaml in your project root.")
	}
	if m.cfg == nil {
		return styleDim.Render("No .fe.yaml found — using conventions.")
	}

	var b strings.Builder
	b.WriteString(styleSelected.Render("Active config") + "\n\n")
	for _, line := range m.cfg.SummaryLines() {
		b.WriteString(line + "\n")
	}
	return b.String()
}

func (m tuiModel) renderFooter() string {
	keys := []string{
		styleKey.Render("↑↓") + " navigate",
		styleKey.Render("enter") + " run",
		styleKey.Render("tab") + " switch tab",
		styleKey.Render("1-4") + " jump tab",
		styleKey.Render("q") + " quit",
	}
	return styleDim.Render("  " + strings.Join(keys, "  ·  "))
}

func itemsForTab(t tabID) []menuItem {
	switch t {
	case tabProject:
		return projectItems
	case tabShip:
		return shipItems
	default:
		return nil
	}
}

// ── Command ───────────────────────────────────────────────────────────────────

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Interactive terminal UI (local only).",
	Long:  `Opens the fe interactive TUI. Requires a real terminal — will not run in CI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tty.IsInteractive() {
			return fmt.Errorf("'fe tui' requires an interactive terminal — use individual commands in CI")
		}

		cfg, cfgErr := config.Load()

		m := newTuiModel(cfg, cfgErr)
		p := tea.NewProgram(m,
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)
		_, err := p.Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

// Ensure ui import is used (it provides shared styles referenced from doctor/scaffold)
var _ = ui.StyleSuccess
