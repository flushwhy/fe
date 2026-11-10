package cmd

import (
	"bufio"
	"fmt"
	"strconv"
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
  cd mygoapp   && fe add gopls

Run fe add with no arguments in a terminal to choose a tool interactively.`, supportedTools()),
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 && tty.IsInteractive() {
			return nil
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		toolName := ""
		if len(args) == 1 {
			toolName = strings.ToLower(args[0])
		} else {
			selected, err := promptTool(bufio.NewReader(cmd.InOrStdin()), cmd.OutOrStdout())
			if err != nil {
				return err
			}
			toolName = string(selected)
		}
		return tooling.Add(tooling.Tool(toolName))
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

func promptTool(reader *bufio.Reader, out io.Writer) (tooling.Tool, error) {
	fmt.Fprintln(out, ui.StyleTitle.Render("Configure a project tool"))
	for i, tool := range tooling.SupportedTools {
		fmt.Fprintf(out, "  %s %s\n", ui.StyleStep.Render(fmt.Sprintf("%d.", i+1)), tool)
	}
	for {
		fmt.Fprint(out, ui.StyleBold.Render("Tool [1]: "))
		choice, err := readPromptLine(reader)
		if err != nil {
			return "", fmt.Errorf("reading tool selection: %w", err)
		}
		if choice == "" {
			choice = "1"
		}
		if index, err := strconv.Atoi(choice); err == nil && index >= 1 && index <= len(tooling.SupportedTools) {
			return tooling.SupportedTools[index-1], nil
		}
		for _, tool := range tooling.SupportedTools {
			if strings.EqualFold(choice, string(tool)) {
				return tool, nil
			}
		}
		fmt.Fprintln(out, ui.StyleWarn.Render("Choose a listed number or tool name."))
	}
}
