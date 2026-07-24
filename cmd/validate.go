package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/flushwhy/fe/internal/config"
	"github.com/flushwhy/fe/internal/ui"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Check config and paths before a build or release.",
	Long: `Validates your .fe.yaml config and checks that referenced paths exist.
Run this in CI before 'fe release' to catch issues early.

Exit code 0 = all good. Non-zero = problems found.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Title("Validating config")
		start := time.Now()

		var cfg *config.Config
		err := ui.RunSpinner("Loading config", func() error {
			var e error
			cfg, e = config.Load()
			return e
		})
		if err != nil {
			return err
		}

		var problems []string
		err = ui.RunSpinner("Checking paths and credentials", func() error {
			if _, err := os.Stat(cfg.Butler.Directory); os.IsNotExist(err) {
				problems = append(problems, fmt.Sprintf("builds directory %q does not exist", cfg.Butler.Directory))
			}
			if cfg.Pack.Input != "" {
				if _, err := os.Stat(cfg.Pack.Input); os.IsNotExist(err) {
					problems = append(problems, fmt.Sprintf("pack input %q does not exist", cfg.Pack.Input))
				}
			}
			if cfg.Itchio.Username == "" {
				problems = append(problems, "itchio.username not set (use .fe.yaml or ITCHIO_USERNAME)")
			}
			if cfg.Itchio.Game == "" {
				problems = append(problems, "itchio.game not set (use .fe.yaml or ITCHIO_GAME)")
			}
			for i, t := range cfg.Build.Targets {
				if t.Cmd == "" {
					problems = append(problems, fmt.Sprintf("build.targets[%d] (%s) has no cmd", i, t.Platform))
				}
			}
			return nil
		})
		if err != nil {
			return err
		}

		fmt.Println()
		fmt.Println(ui.StyleBold.Render("Active config:"))
		for _, line := range cfg.SummaryLines() {
			fmt.Println(line)
		}
		fmt.Println()

		if len(problems) == 0 {
			ui.OK("Everything looks good")
			ui.Elapsed(start)
			return nil
		}

		for _, p := range problems {
			ui.Fail("%s", p)
		}
		return fmt.Errorf("validation failed with %d problem(s)", len(problems))
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
