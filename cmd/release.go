package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/flushwhy/fe/internal/build"
	"github.com/flushwhy/fe/internal/butler"
	"github.com/flushwhy/fe/internal/config"
	"github.com/flushwhy/fe/internal/pack"
	"github.com/flushwhy/fe/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Full release pipeline: build → pack → push to itch.io.",
	Long: `Runs the complete release pipeline in order:
  1. fe build   — compile all platform targets
  2. fe pack    — pack sprites into spritesheet
  3. fe bmp     — push builds to itch.io via butler

Skip individual steps with --skip-build, --skip-pack, --skip-bmp.

CI example:
  export ITCHIO_USERNAME=myuser ITCHIO_GAME=mygame
  fe release --userversion v1.2.3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		skipBuild, _ := cmd.Flags().GetBool("skip-build")
		skipPack, _ := cmd.Flags().GetBool("skip-pack")
		skipBmp, _ := cmd.Flags().GetBool("skip-bmp")

		ui.Title("Release Pipeline")
		start := time.Now()

		// ── Step 1: Build ─────────────────────────────────────────────────
		ui.Step(1, 3, "Build")
		if !skipBuild {
			if len(cfg.Build.Targets) == 0 {
				ui.Warn("No build targets defined — skipping build step")
			} else {
				steps := make([]ui.StepTask, len(cfg.Build.Targets))
				for i, t := range cfg.Build.Targets {
					t := t
					steps[i] = ui.StepTask{
						Label: fmt.Sprintf("Building %s", t.Platform),
						Run:   func() error { return build.RunTarget(t, build.DefaultRunner) },
					}
				}
				if err := ui.RunSteps(steps); err != nil {
					return fmt.Errorf("build step failed: %w", err)
				}
			}
		} else {
			ui.Dim("skipped")
		}

		// ── Step 2: Pack ──────────────────────────────────────────────────
		fmt.Println()
		ui.Step(2, 3, "Pack sprites")
		if !skipPack {
			err := ui.RunSpinner(fmt.Sprintf("Packing %s → %s", cfg.Pack.Input, cfg.Pack.Output), func() error {
				return pack.Run(cfg.Pack)
			})
			if err != nil {
				ui.Warn("Pack step: %v (continuing)", err)
			}
		} else {
			ui.Dim("skipped")
		}

		// ── Step 3: Push ──────────────────────────────────────────────────
		fmt.Println()
		ui.Step(3, 3, "Push to itch.io")
		if !skipBmp {
			if errs := cfg.Validate("bmp"); len(errs) > 0 {
				for _, e := range errs {
					ui.Fail("%s", e)
				}
				return fmt.Errorf("cannot push — fix config errors above")
			}
			logger := log.New(os.Stdout, "", 0)
			err := ui.RunSpinner(
				fmt.Sprintf("Pushing to %s/%s", cfg.Itchio.Username, cfg.Itchio.Game),
				func() error {
					return butler.Push(
						cfg.Itchio.Username,
						cfg.Itchio.Game,
						cfg.Butler.Directory,
						cfg.Butler.Userversion,
						exec.Command,
						logger,
					)
				},
			)
			if err != nil {
				return fmt.Errorf("push step failed: %w", err)
			}
		} else {
			ui.Dim("skipped")
		}

		fmt.Println()
		ui.OK("Release complete")
		ui.Elapsed(start)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)
	releaseCmd.Flags().Bool("skip-build", false, "skip the build step")
	releaseCmd.Flags().Bool("skip-pack", false, "skip the sprite pack step")
	releaseCmd.Flags().Bool("skip-bmp", false, "skip the itch.io push step")
	releaseCmd.Flags().String("userversion", "", "version string for itch.io channel")
	viper.BindPFlag("butler.userversion", releaseCmd.Flags().Lookup("userversion"))
}
