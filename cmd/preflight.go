package cmd

import (
	"fmt"

	"github.com/flushwhy/fe/internal/config"
	"github.com/flushwhy/fe/internal/preflight"
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
		cfg, err := config.Load()
		if err != nil {
			return err
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
		return fmt.Errorf("preflight found %d problem(s)", failed)
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
