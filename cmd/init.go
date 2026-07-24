package cmd

import (
	"fmt"
	"strings"

	"github.com/flushwhy/fe/internal/scaffold"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var initCmd = &cobra.Command{
	Use:   "init <lang> --name <project>",
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

Templates come from built-in defaults. To use your own templates instead,
set templates.source in .fe.yaml:

  templates:
    source: "https://github.com/you/fe-templates"`, supportedLangs()),
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lang := scaffold.Lang(strings.ToLower(args[0]))
		name, _ := cmd.Flags().GetString("name")
		module, _ := cmd.Flags().GetString("module")

		if name == "" {
			name = string(lang) + "-project"
		}

		// Remote URL from flag or config file.
		remoteURL := viper.GetString("templates.source")

		return scaffold.Init(scaffold.Options{
			Lang:      lang,
			Name:      name,
			Module:    module,
			RemoteURL: remoteURL,
		})
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().String("name", "", "project name / directory to create")
	initCmd.Flags().String("module", "", "Go module path (go only, e.g. github.com/you/myapp)")

	viper.BindPFlag("templates.source", initCmd.Flags().Lookup("remote"))
}

func supportedLangs() string {
	langs := make([]string, len(scaffold.SupportedLangs))
	for i, l := range scaffold.SupportedLangs {
		langs[i] = string(l)
	}
	return strings.Join(langs, ", ")
}
