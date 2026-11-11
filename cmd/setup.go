package cmd

import (
	"fmt"
	"io"

	projectsetup "github.com/flushwhy/fe/internal/setup"
	"github.com/flushwhy/fe/internal/ui"
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
		if project, projectErr := findGodotProject("."); projectErr == nil {
			restorePath, err := activateProjectGodotPath(project)
			if err != nil {
				return err
			}
			defer restorePath()
		}
		root := "."
		if len(args) == 1 {
			root = args[0]
		}
		tools, err := projectsetup.Inspect(root)
		if err != nil {
			return err
		}
		printSetupResults(cmd.OutOrStdout(), root, tools)
		return nil
	},
}

func printSetupResults(out io.Writer, root string, tools []projectsetup.Tool) {
	fmt.Fprintf(out, "%s\n  Directory: %s\n\n", ui.StyleTitle.Render("Project setup"), root)
	if len(tools) == 0 {
		fmt.Fprintln(out, "No tool requirements detected.")
		fmt.Fprintln(out, "Run 'fe inspect' to see the project markers Fe recognizes.")
		return
	}

	missing := 0
	for _, tool := range tools {
		status := ui.StyleSuccess.Render(ui.IconOK + " FOUND")
		if tool.Found == "" {
			status = ui.StyleError.Render(ui.IconFail + " MISSING")
			if tool.Required {
				missing++
			}
		}
		requirement := "required"
		if !tool.Required {
			requirement = "optional"
		}
		fmt.Fprintf(out, "  %-12s %-18s (%s) — %s\n", status, tool.Name, requirement, tool.Reason)
		if tool.Found != "" {
			fmt.Fprintf(out, "               %s\n", ui.StyleDim.Render(tool.Found))
		} else if tool.Required {
			fmt.Fprintf(out, "               Install %s and ensure it is available in PATH.\n", tool.Name)
		} else {
			fmt.Fprintln(out, "               Optional; install only if you use this workflow.")
		}
	}

	fmt.Fprintln(out)
	if missing == 0 {
		fmt.Fprintln(out, ui.StyleSuccess.Render(ui.IconOK+" All required tools are available."))
		fmt.Fprintln(out, "Next: run 'fe doctor' to check project configuration.")
	} else {
		fmt.Fprintf(out, ui.StyleError.Render("%s %d required tool(s) missing.")+" Install them before running the corresponding build or test workflow.\n", ui.IconFail, missing)
		fmt.Fprintln(out, "Next: install the missing tools, then run 'fe setup' again.")
	}
	fmt.Fprintln(out, ui.StyleDim.Render("Read-only check: no software was installed and no files were changed."))
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
