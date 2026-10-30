package cmd

import (
	"fmt"

	projectsetup "github.com/flushwhy/fe/internal/setup"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup [path]",
	Short: "Show the tools needed to work with this project.",
	Long: `Inspects project conventions and checks for the relevant toolchains.

This command is read-only. It does not install software or modify project files.
Optional tools are shown separately from tools required for the detected workflow.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) == 1 {
			root = args[0]
		}
		tools, err := projectsetup.Inspect(root)
		if err != nil {
			return err
		}
		if len(tools) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No tool requirements detected.")
			fmt.Fprintln(cmd.OutOrStdout(), "Run 'fe inspect' to see the project markers Fe recognizes.")
			return nil
		}

		missing := 0
		for _, tool := range tools {
			status := "FOUND"
			if tool.Found == "" {
				status = "MISSING"
				if tool.Required {
					missing++
				}
			}
			requirement := "required"
			if !tool.Required {
				requirement = "optional"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%-7s %-18s (%s) — %s\n", status, tool.Name, requirement, tool.Reason)
			if tool.Found != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "        %s\n", tool.Found)
			} else if tool.Required {
				fmt.Fprintf(cmd.OutOrStdout(), "        Install %s and ensure it is available in PATH.\n", tool.Name)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "        Not needed for basic ZIP packaging.\n")
			}
		}
		if missing == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "\nRequired tools are available.")
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "\n%d required tool(s) missing. Install them before running the corresponding build or test workflow.\n", missing)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "No software was installed and no files were changed.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
