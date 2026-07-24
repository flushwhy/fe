package cmd

import (
	"os"

	"github.com/flushwhy/fe/internal/doctor"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check that required tools are installed and configs are correct.",
	Long: `Inspects the current project directory, detects the language, and checks:

  • Required binaries are in PATH (clangd, bear, ols, zls, gopls, etc.)
  • LSP config files exist and are valid (.clangd, ols.json, zls.json)
  • compile_commands.json exists and is not an empty stub (C/C++)
  • cmake has EXPORT_COMPILE_COMMANDS set (C++ cmake projects)
  • go.mod exists (Go projects)

Exit code 0 if all checks pass. Non-zero if anything fails (CI-safe).

Run 'fe add <tool>' to fix most issues.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		results := doctor.Run()
		allOK := doctor.Print(results)
		if !allOK {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
