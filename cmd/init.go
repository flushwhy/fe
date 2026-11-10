package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/flushwhy/fe/internal/scaffold"
	"github.com/flushwhy/fe/internal/tty"
	"github.com/flushwhy/fe/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var initCmd = &cobra.Command{
	Use:   "init [lang] --name <project>",
	Short: "Scaffold a new project for a given language.",
	Long: fmt.Sprintf(`Creates a new project directory with language-appropriate structure,
build files, and editor/LSP config stubs.

Supported languages: %s

Examples:
  fe init odin --name mygame
  fe init c --name mylib
  fe init cpp --name myengine
  fe init zig --name mytool
  fe init go --name myapp --module github.com/you/myapp
  fe init --template cpp-raylib-xmake --name mygame

Named templates are downloaded from https://github.com/flushwhy/fe-templates and cached locally.
List available templates with: fe template list
Refresh a cached template with: fe template update <template-id>

Run fe init with no arguments in a terminal for an interactive setup.`, supportedLangs()),
	Args: func(cmd *cobra.Command, args []string) error {
		templateID, _ := cmd.Flags().GetString("template")
		if templateID != "" {
			if len(args) > 0 {
				return fmt.Errorf("provide either a language or --template, not both")
			}
			return nil
		}
		if len(args) > 1 || (len(args) == 0 && !tty.IsInteractive()) {
			return fmt.Errorf("provide a language, or run 'fe init' in an interactive terminal")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		templateID, _ := cmd.Flags().GetString("template")
		lang := scaffold.Lang("")
		reader := bufio.NewReader(cmd.InOrStdin())

		if len(args) > 0 {
			lang = scaffold.Lang(strings.ToLower(args[0]))
		} else if templateID == "" {
			var err error
			lang, err = promptInitLanguage(reader, cmd.OutOrStdout())
			if err != nil {
				return err
			}
		}

		name, _ := cmd.Flags().GetString("name")
		module, _ := cmd.Flags().GetString("module")
		defaultName := string(lang) + "-project"
		if templateID != "" {
			defaultName = templateID + "-project"
		}
		if name == "" {
			name = defaultName
			if tty.IsInteractive() {
				var err error
				name, err = promptProjectName(reader, cmd.OutOrStdout(), defaultName)
				if err != nil {
					return err
				}
			}
		}

		remoteURL := viper.GetString("templates.source")
		if err := scaffold.Init(scaffold.Options{
			Lang:       lang,
			Name:       name,
			Module:     module,
			RemoteURL:  remoteURL,
			TemplateID: templateID,
		}); err != nil {
			return err
		}
		if isCI() {
			return nil
		}
		_, err := syncProjectRoot(name)
		return err
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().String("name", "", "project name / directory to create")
	initCmd.Flags().String("module", "", "Go module path (go only, e.g. github.com/you/myapp)")
	initCmd.Flags().String("template", "", "named template ID from flushwhy/fe-templates")
}

func supportedLangs() string {
	langs := make([]string, len(scaffold.SupportedLangs))
	for i, l := range scaffold.SupportedLangs {
		langs[i] = string(l)
	}
	return strings.Join(langs, ", ")
}

func promptInitLanguage(reader *bufio.Reader, out io.Writer) (scaffold.Lang, error) {
	fmt.Fprintln(out, ui.StyleTitle.Render("Create a project"))
	fmt.Fprintln(out, "Choose a language:")
	for i, lang := range scaffold.SupportedLangs {
		fmt.Fprintf(out, "  %s %s\n", ui.StyleStep.Render(fmt.Sprintf("%d.", i+1)), lang)
	}

	for {
		fmt.Fprint(out, ui.StyleBold.Render("Language [1]: "))
		choice, err := readPromptLine(reader)
		if err != nil {
			return "", fmt.Errorf("reading language selection: %w", err)
		}
		if choice == "" {
			choice = "1"
		}
		if index, err := strconv.Atoi(choice); err == nil && index >= 1 && index <= len(scaffold.SupportedLangs) {
			return scaffold.SupportedLangs[index-1], nil
		}
		for _, lang := range scaffold.SupportedLangs {
			if strings.EqualFold(choice, string(lang)) {
				return lang, nil
			}
		}
		fmt.Fprintln(out, ui.StyleWarn.Render("Choose a listed number or language name."))
	}
}

func promptProjectName(reader *bufio.Reader, out io.Writer, fallback string) (string, error) {
	fmt.Fprintf(out, "%s ", ui.StyleBold.Render("Project name ["+fallback+"]:"))
	name, err := readPromptLine(reader)
	if err != nil {
		return "", fmt.Errorf("reading project name: %w", err)
	}
	if name == "" {
		return fallback, nil
	}
	return name, nil
}

func readPromptLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	if err == io.EOF && line == "" {
		return "", io.EOF
	}
	return strings.TrimSpace(line), nil
}
