package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/flushwhy/fe/internal/build"
	"github.com/flushwhy/fe/internal/butler"
	"github.com/flushwhy/fe/internal/config"
	"github.com/flushwhy/fe/internal/godot"
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
  2. fe pack    — optionally pack sprites into a spritesheet
  3. fe bmp     — push builds to itch.io via butler

Sprite packing is skipped with a warning when the configured sprite input is missing.
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
		channelSuffix, _ := cmd.Flags().GetString("channel-suffix")

		ui.Title("Release Pipeline")
		start := time.Now()

		// ── Step 1: Build/export ───────────────────────────────────────────
		ui.Step(1, 3, "Build / export")
		if !skipBuild {
			if _, statErr := os.Stat("project.godot"); statErr == nil {
				if err := godot.Export(".", cfg.Godot.Executable, cfg.Godot.OutputDir, "", exec.Command); err != nil {
					return fmt.Errorf("Godot export step failed: %w", err)
				}
			} else if len(cfg.Build.Targets) == 0 {
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

		// A missing build directory is often caused by an empty build.targets
		// configuration. Catch it before packing assets or attempting an upload.
		// When --skip-bmp is set, the caller explicitly requested no upload.
		if !skipBmp {
			if err := validateReleaseBuildDir(cfg.Butler.Directory); err != nil {
				return err
			}
		}

		// ── Step 2: Pack ──────────────────────────────────────────────────
		fmt.Println()
		ui.Step(2, 3, "Pack sprites")
		if skipPack {
			ui.Dim("skipped")
		} else if !releaseSpriteInputsAvailable(cfg.Pack) {
			ui.Warn("No configured sprite input directory found — skipping sprite packing (software-only release). Add sprites or configure pack.input / pack.sheets in .fe.yaml if you want a spritesheet.")
		} else {
			err := ui.RunSpinner(fmt.Sprintf("Packing %s → %s", cfg.Pack.Input, cfg.Pack.Output), func() error {
				return pack.Run(cfg.Pack)
			})
			if err != nil {
				ui.Warn("Pack step: %v (continuing)", err)
			}
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
						channelSuffix,
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

// validateReleaseBuildDir prevents a release from reaching the itch.io push
// step when no build output was produced or supplied.
func validateReleaseBuildDir(directory string) error {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("release cannot push to itch.io: build output directory %q is unavailable (%w); configure build.targets in .fe.yaml or provide prebuilt platform folders and set butler.directory to their parent", directory, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return nil
		}
	}
	return fmt.Errorf("release cannot push to itch.io: build output directory %q contains no platform folders; configure build.targets in .fe.yaml or provide prebuilt platform folders", directory)
}

// releaseSpriteInputsAvailable reports whether any configured sprite input
// directory exists. Sprite assets are optional for software-only releases.
func releaseSpriteInputsAvailable(cfg config.PackConfig) bool {
	if len(cfg.Sheets) == 0 {
		input := cfg.Input
		if input == "" {
			input = "assets/sprites"
		}
		info, err := os.Stat(input)
		return err == nil && info.IsDir()
	}

	root := cfg.Input
	if root == "" {
		root = "assets/sprites"
	}
	for name, sheet := range cfg.Sheets {
		input := sheet.Input
		if input == "" {
			input = filepath.Join(root, name)
		}
		info, err := os.Stat(input)
		if err == nil && info.IsDir() {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(releaseCmd)
	releaseCmd.Flags().Bool("skip-build", false, "skip the build step")
	releaseCmd.Flags().Bool("skip-pack", false, "skip the sprite pack step")
	releaseCmd.Flags().Bool("skip-bmp", false, "skip the itch.io push step")
	releaseCmd.Flags().String("userversion", "", "version string for itch.io channel")
	releaseCmd.Flags().String("channel-suffix", "", "append a suffix to itch.io platform channels (for example rolling)")
	viper.BindPFlag("butler.userversion", releaseCmd.Flags().Lookup("userversion"))
}
