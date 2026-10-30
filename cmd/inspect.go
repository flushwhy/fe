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

		out := cmd.OutOrStdout()
		fmt.Fprintln(out, "Project inspection")
		fmt.Fprintf(out, "  Directory: %s\n\n", report.Root)
		if len(report.Projects) == 0 {
			fmt.Fprintln(out, "No known project type detected.")
			fmt.Fprintln(out, "You can still use Fe with custom build targets in .fe.yaml.")
			fmt.Fprintln(out, "Tip: add a supported project manifest or configure the project explicitly.")
			return nil
		}

		fmt.Fprintf(out, "Detected %d project type(s):\n", len(report.Projects))
		for _, project := range report.Projects {
			fmt.Fprintf(out, "\n  %s\n", project.Type)
			for _, marker := range project.Markers {
				fmt.Fprintf(out, "    - %s\n", marker)
			}
		}

		fmt.Fprintln(out, "\nBuild suggestions")
		if len(report.Suggestions) == 0 {
			fmt.Fprintln(out, "  No safe default build command is available for these detected project types.")
			fmt.Fprintln(out, "  Configure build.targets in .fe.yaml or use the project’s native build instructions.")
		} else {
			for _, suggestion := range report.Suggestions {
				fmt.Fprintf(out, "  %s: %s %s\n", suggestion.ProjectType, suggestion.Command, strings.Join(suggestion.Args, " "))
				fmt.Fprintf(out, "    %s\n", suggestion.Reason)
			}
		}
		fmt.Fprintln(out, "  Suggestions are informational; Fe does not run them or change .fe.yaml.")
		fmt.Fprintln(out, "\nInspection is read-only. Detection identifies project conventions; it does not guarantee that every target can be built on this machine.")
		fmt.Fprintln(out, "\nNext steps:")
		fmt.Fprintln(out, "  fe doctor       Check project and tool prerequisites")
		fmt.Fprintln(out, "  fe build        Run configured build targets")
		fmt.Fprintln(out, "  fe release      Run the configured release workflow")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}
