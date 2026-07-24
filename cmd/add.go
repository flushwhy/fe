package cmd

import (
	"fmt"
	"strings"

	"github.com/flushwhy/fe/internal/tooling"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <tool>",
	Short: "Wire up an LSP or editor tool for the current project.",
	Long: fmt.Sprintf(`Adds editor/LSP configuration files to the current project directory
so your tooling works correctly from day one.

Supported tools: %s

clangd  — writes .clangd with correct CompilationDatabase path,
           creates compile_commands.json stub if missing so clangd
           starts without errors. Detects cmake vs bear/make layout.

ols     — writes or patches ols.json for the Odin Language Server.

zls     — writes zls.json for the Zig Language Server.

gopls   — writes .gopls.yaml workspace settings for Go.

Examples:
  cd myproject && fe add clangd
  cd myodingame && fe add ols
  cd myzigapp  && fe add zls
  cd mygoapp   && fe add gopls`, supportedTools()),
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
