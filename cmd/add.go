package cmd

import (
	"fmt"
	"strings"

	"github.com/flushwhy/fe/internal/tooling"
	"github.com/flushwhy/fe/internal/tty"
	"github.com/flushwhy/fe/internal/ui"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <tool>",
	Short: "Wire up an LSP or editor tool for the current project.",
	Long: fmt.Sprintf(`Adds editor/LSP configuration files to the current project directory
so your tooling works correctly from day one.

Run 'fe add' without a tool in an interactive terminal to choose one.

Supported tools: %s

clangd  — writes .clangd with the detected CompilationDatabase path.
           For CMake projects, run 'fe fix compile-commands' to generate
           a real database; Fe does not create an empty placeholder.

ols     — writes or patches ols.json for the Odin Language Server.

zls     — writes zls.json for the Zig Language Server.

gopls   — writes .gopls.yaml workspace settings for Go.

Examples:
  cd myproject && fe add clangd
  cd myproject && fe fix compile-commands
  cd myodingame && fe add ols
  cd myzigapp  && fe add zls
  cd mygoapp   && fe add gopls`, supportedTools()),
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			if !tty.IsInteractive() {
				return fmt.Errorf("provide a tool, for example: fe add clangd")
			}
			choices := make([]string, len(tooling.SupportedTools))
			for i, supported := range tooling.SupportedTools {
				choices[i] = string(supported)
			}
			selected, err := ui.Select("Choose a tool to configure", choices)
			if err != nil {
				return err
			}
			args = []string{selected}
		}
		tool := tooling.Tool(strings.ToLower(args[0]))
		return tooling.Add(tool)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func supportedTools() string {
	tools := make([]string, len(tooling.SupportedTools))
	for i, t := range tooling.SupportedTools {
		tools[i] = string(t)
	}
	return strings.Join(tools, ", ")
}
