package cmd

import (
	"fmt"
	"strings"

	projectinspect "github.com/flushwhy/fe/internal/inspect"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect [path]",
	Short: "Detect project types and explain what Fe recognizes.",
	Long: `Inspects a directory for common project and ecosystem markers.

This command is read-only: it does not execute project code, install tools,
or modify files. If multiple project types are present, Fe reports all of them.

Examples:
  fe inspect
  fe inspect ./addons/my-addon`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) == 1 {
			root = args[0]
		}
		report, err := projectinspect.Inspect(root)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Project inspection\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  Directory: %s\n\n", report.Root)
		if len(report.Projects) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No known project type detected.")
			fmt.Fprintln(cmd.OutOrStdout(), "You can still use Fe with custom build targets in .fe.yaml.")
			fmt.Fprintln(cmd.OutOrStdout(), "Tip: add a supported project manifest or configure the project explicitly.")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Detected %d project type(s):\n", len(report.Projects))
		for _, project := range report.Projects {
			fmt.Fprintf(cmd.OutOrStdout(), "\n  %s\n", project.Type)
			for _, marker := range project.Markers {
				fmt.Fprintf(cmd.OutOrStdout(), "    - %s\n", marker)
			}
		}
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "Inspection is read-only. Detection identifies project conventions; it does not guarantee that every target can be built on this machine.")
		fmt.Fprintf(cmd.OutOrStdout(), "\nNext steps:\n  fe doctor       Check project and tool prerequisites\n  fe build        Run configured build targets\n  fe release      Run the configured release workflow\n")
		_ = strings.Builder{} // keep command output independent of terminal styling
		return nil
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}
