package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flushwhy/fe/internal/projectstate"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync [path]",
	Short: "Sync Fe-managed project files and register local cleanup state.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if isCI() {
			return fmt.Errorf("fe sync is for developer machines and is disabled in CI")
		}
		root := "."
		if len(args) == 1 {
			root = args[0]
		}
		if err := projectstate.Sync(root); err != nil {
			return err
		}
		abs, err := filepath.Abs(root)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Synced Fe project files for %s\n", abs)
		return nil
	},
}

var cleanCmd = &cobra.Command{
	Use:   "clean [path]",
	Short: "Remove files and registrations managed by Fe.",
	Long: "Removes only files recorded as Fe-managed. Modified files are preserved.\n" +
		"Use --all to clean every registered project, including records for deleted repositories.\n" +
		"Use --dry-run to preview changes without modifying anything.",
	Args: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		if all && len(args) > 0 {
			return fmt.Errorf("provide a project path or --all, not both")
		}
		return cobra.MaximumNArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		all, _ := cmd.Flags().GetBool("all")
		keepUninstallers, _ := cmd.Flags().GetBool("keep-uninstallers")
		var results []projectstate.Cleanup
		if all {
			var err error
			results, err = projectstate.CleanAll(dryRun)
			if err != nil {
				return err
			}
		} else {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}
			result, err := projectstate.Clean(root, dryRun, keepUninstallers)
			if err != nil {
				return err
			}
			results = append(results, result)
		}
		for _, result := range results {
			for _, path := range result.Removed {
				action := "Removed"
				if dryRun {
					action = "Would remove"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", action, path)
			}
			for _, path := range result.Preserved {
				fmt.Fprintf(cmd.OutOrStdout(), "Preserved: %s\n", path)
			}
			for _, path := range result.Retained {
				fmt.Fprintf(cmd.OutOrStdout(), "Retained: %s\n", path)
			}
		}
		if len(results) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No Fe-managed project records found.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd, cleanCmd)
	cleanCmd.Flags().Bool("all", false, "clean every registered project and remove stale records")
	cleanCmd.Flags().Bool("dry-run", false, "show what would be removed without changing files")
	cleanCmd.Flags().Bool("keep-uninstallers", false, "leave uninstall scripts for the script to remove itself")
}

func isCI() bool {
	return strings.EqualFold(os.Getenv("CI"), "true") ||
		strings.EqualFold(os.Getenv("GITHUB_ACTIONS"), "true") ||
		os.Getenv("JENKINS_URL") != "" ||
		os.Getenv("BUILDKITE") != "" ||
		os.Getenv("BUILD_ID") != "" ||
		os.Getenv("TF_BUILD") != ""
}
