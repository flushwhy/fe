package cmd

import (
	"fmt"
	"strings"

	"github.com/flushwhy/fe/internal/scaffold"
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
Refresh a cached template with: fe template update <template-id>`, supportedLangs()),
	Args: func(cmd *cobra.Command, args []string) error {
		templateID, _ := cmd.Flags().GetString("template")
		if templateID != "" && len(args) == 0 {
			return nil
		}
		if len(args) != 1 {
			return fmt.Errorf("provide a language or use --template <id>")
		}
		if templateID != "" {
			return fmt.Errorf("provide either a language or --template, not both")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		templateID, _ := cmd.Flags().GetString("template")
		lang := scaffold.Lang("")
		if len(args) > 0 {
			lang = scaffold.Lang(strings.ToLower(args[0]))
		}
		name, _ := cmd.Flags().GetString("name")
		module, _ := cmd.Flags().GetString("module")

		if name == "" {
			if templateID != "" {
				name = templateID + "-project"
			} else {
				name = string(lang) + "-project"
			}
		}

		// Remote URL from flag or config file.
		remoteURL := viper.GetString("templates.source")

		return scaffold.Init(scaffold.Options{
			Lang:      lang,
			Name:      name,
			Module:    module,
			RemoteURL: remoteURL,
			TemplateID: templateID,
		})
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
