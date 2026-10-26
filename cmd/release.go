package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			return printReleasePlan(cfg, skipBuild, skipPack, skipBmp, channelSuffix)
		}

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
				return fmt.Errorf("pack step failed; refusing to publish an incomplete release: %w", err)
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
			commit := os.Getenv("GITHUB_SHA")
			if commit == "" {
				if result, err := exec.Command("git", "rev-parse", "HEAD").Output(); err == nil {
					commit = strings.TrimSpace(string(result))
				}
			}
			manifestPath := filepath.Join(cfg.Butler.Directory, "artifact-manifest.json")
			if err := writeArtifactManifest(cfg.Butler.Directory, manifestPath, cfg.Butler.Userversion, commit); err != nil {
				return fmt.Errorf("create artifact manifest: %w", err)
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

func printReleasePlan(cfg *config.Config, skipBuild, skipPack, skipBmp bool, channelSuffix string) error {
	fmt.Println("Release dry run — no builds, exports, packing, or uploads will be performed.")
	fmt.Println()
	problems := 0
	if skipBuild {
		fmt.Println("1. Build/export: skipped by --skip-build")
	} else if _, err := os.Stat("project.godot"); err == nil {
		presets, err := godot.ReadPresets(filepath.Join(".", "export_presets.cfg"))
		if err != nil {
			fmt.Printf("1. Build/export: cannot inspect Godot presets: %v\n", err)
			problems++
		} else {
			fmt.Printf("1. Build/export: %d Godot preset(s) from export_presets.cfg\n", len(presets))
			for _, preset := range presets {
				projectDir, _ := filepath.Abs(".")
				output, err := godot.ExpectedOutputPath(projectDir, cfg.Godot.OutputDir, preset)
				if err != nil {
					fmt.Printf("   - %s (%s): unsupported platform\n", preset.Name, preset.Platform)
					problems++
					continue
				}
				fmt.Printf("   - %s -> %s\n", preset.Name, output)
			}
		}
	} else if len(cfg.Build.Targets) == 0 {
		fmt.Println("1. Build: no targets configured (will be skipped)")
	} else {
		fmt.Printf("1. Build: %d configured target(s)\n", len(cfg.Build.Targets))
		for _, target := range cfg.Build.Targets {
			if target.Cmd == "" {
				fmt.Printf("   - %s: missing command\n", target.Platform)
				problems++
			} else {
				fmt.Printf("   - %s: %s\n", target.Platform, target.Cmd)
			}
		}
	}

	if skipPack {
		fmt.Println("2. Sprite packing: skipped by --skip-pack")
	} else if releaseSpriteInputsAvailable(cfg.Pack) {
		fmt.Printf("2. Sprite packing: %s -> %s\n", cfg.Pack.Input, cfg.Pack.Output)
	} else {
		fmt.Println("2. Sprite packing: skipped (no configured sprite input directory)")
	}

	if skipBmp {
		fmt.Println("3. itch.io upload: skipped by --skip-bmp")
	} else {
		fmt.Printf("3. Artifact manifest: %s\n", filepath.Join(cfg.Butler.Directory, "artifact-manifest.json"))
		fmt.Printf("4. itch.io upload: %s/%s from %s\n", cfg.Itchio.Username, cfg.Itchio.Game, cfg.Butler.Directory)
		if cfg.Itchio.Username == "" {
			fmt.Println("   - missing ITCHIO_USERNAME / itchio.username")
			problems++
		}
		if cfg.Itchio.Game == "" {
			fmt.Println("   - missing ITCHIO_GAME / itchio.game")
			problems++
		}
		if cfg.Butler.Userversion == "" {
			fmt.Println("   - version not set; Butler will use its default version behavior")
		} else {
			fmt.Printf("   - version: %s\n", cfg.Butler.Userversion)
		}
		if channelSuffix != "" {
			fmt.Printf("   - channel suffix: %s\n", channelSuffix)
		}
		if _, err := os.Stat(cfg.Butler.Directory); err != nil {
			fmt.Printf("   - build directory not currently available: %s\n", cfg.Butler.Directory)
		}
	}
	if problems > 0 {
		return fmt.Errorf("dry run found %d issue(s) to review", problems)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(releaseCmd)
	releaseCmd.Flags().Bool("dry-run", false, "show the release plan without running commands or publishing")
	releaseCmd.Flags().Bool("skip-build", false, "skip the build step")
	releaseCmd.Flags().Bool("skip-pack", false, "skip the sprite pack step")
	releaseCmd.Flags().Bool("skip-bmp", false, "skip the itch.io push step")
	releaseCmd.Flags().String("userversion", "", "version string for itch.io channel")
	releaseCmd.Flags().String("channel-suffix", "", "append a suffix to itch.io platform channels (for example rolling)")
	viper.BindPFlag("butler.userversion", releaseCmd.Flags().Lookup("userversion"))
}
