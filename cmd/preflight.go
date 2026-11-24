package cmd

import (
	"fmt"

	"github.com/flushwhy/fe/internal/config"
	"github.com/flushwhy/fe/internal/preflight"
	"github.com/flushwhy/fe/internal/ui"
	"github.com/spf13/cobra"
)

var preflightCmd = &cobra.Command{
	Use:   "preflight",
	Short: "Inspect project toolchains and run configured pre-build checks.",
	Long: `Detects Python and Lua projects, checks for their runtimes, and runs
commands configured under preflight.checks in .fe.yaml.

Checks are read-only by default. Missing tools are reported with installation
guidance; Fe never installs software automatically.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadProjectConfig()
		if err != nil {
			return err
		}
		if project, projectErr := findGodotProject("."); projectErr == nil {
			restorePath, err := activateProjectGodotPath(project)
			if err != nil {
				return err
			}
			defer restorePath()
		}
		return runPreflight(cfg, "all")
	},
}

func runPreflight(cfg *config.Config, stage string) error {
	results, err := preflight.Run(cfg.Preflight, stage, ".")
	if err != nil {
		return err
	}
	if len(results) == 0 {
		fmt.Println("No applicable preflight checks.")
		return nil
	}
	failed := 0
	for _, result := range results {
		status := "PASS"
		if !result.OK {
			status = "FAIL"
			failed++
		}
		fmt.Printf("%-5s %s: %s\n", status, result.Name, result.Message)
		if result.Fix != "" {
			fmt.Printf("      Fix: %s\n", result.Fix)
		}
	}
	if failed > 0 {
		details := make([]string, 0, failed)
		for _, result := range results {
			if !result.OK {
				detail := result.Name + ": " + result.Message
				if result.Fix != "" {
					detail += " (fix: " + result.Fix + ")"
				}
				details = append(details, detail)
			}
		}
		return &ui.CommandError{
			Message: fmt.Sprintf("preflight found %d problem(s)", failed),
			Details: details,
			Hint:    "resolve the failed checks above, then rerun fe preflight",
		}
	}
	return nil
}

func runConfiguredPreflight(cfg *config.Config, stage string) error {
	if !cfg.Preflight.Enabled {
		return nil
	}
	fmt.Printf("Running %s preflight checks...\n", stage)
	return runPreflight(cfg, stage)
}

func init() {
	rootCmd.AddCommand(preflightCmd)
}
