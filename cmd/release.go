package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/flushwhy/fe/internal/archivepkg"
	"github.com/flushwhy/fe/internal/build"
	"github.com/flushwhy/fe/internal/butler"
	"github.com/flushwhy/fe/internal/changelog"
	"github.com/flushwhy/fe/internal/config"
	"github.com/flushwhy/fe/internal/godot"
	"github.com/flushwhy/fe/internal/gumroad"
	"github.com/flushwhy/fe/internal/pack"
	"github.com/flushwhy/fe/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Build, package, and publish a project release.",
	Long: `Runs the release pipeline in order:
  1. fe build   — compile all configured platform targets
  2. fe pack    — optionally pack sprites into a spritesheet
  3. changelog  — optionally write Git-based release notes
  4. package    — optionally archive build artifacts as ZIP or tar.gz
  5. fe bmp     — push builds to itch.io via butler

Sprite packing is skipped with a warning when the configured sprite input is missing.
Use --changelog-output and --package-format to enable the optional stages.
Skip individual steps with --skip-build, --skip-pack, and --skip-bmp.\nUse --gumroad to upload a stable version to its explicitly configured Gumroad product.

CI example:
  export ITCHIO_USERNAME=myuser ITCHIO_GAME=mygame
  fe release --userversion v1.2.3 --package-format zip`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadProjectConfig()
		if err != nil {
			return err
		}

		skipBuild, _ := cmd.Flags().GetBool("skip-build")
		skipPack, _ := cmd.Flags().GetBool("skip-pack")
		skipBmp, _ := cmd.Flags().GetBool("skip-bmp")
		gumroadRequested, _ := cmd.Flags().GetBool("gumroad")
		channelSuffix, _ := cmd.Flags().GetString("channel-suffix")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		allowDirty, _ := cmd.Flags().GetBool("allow-dirty")
		allowUnpushed, _ := cmd.Flags().GetBool("allow-unpushed")
		gitState, gitErr := inspectReleaseGit()
		if gitErr != nil {
			return gitErr
		}
		changelogOutput, _ := cmd.Flags().GetString("changelog-output")
		packageFormat, _ := cmd.Flags().GetString("package-format")
		packageOutput, _ := cmd.Flags().GetString("package-output")
		if packageFormat != "" && packageFormat != "zip" && packageFormat != "tar.gz" {
			return fmt.Errorf("unsupported package format %q (use zip or tar.gz)", packageFormat)
		}
		if packageFormat != "" {
			packageOutput, err = releasePackagePath(packageOutput, packageFormat, cfg.Butler.Userversion)
			if err != nil {
				return err
			}
		}
		if dryRun {
			printReleaseGitState(gitState)
			planErr := printReleasePlan(cfg, skipBuild, skipPack, skipBmp, gumroadRequested, channelSuffix, changelogOutput, packageFormat, packageOutput)
			if gitErr := validateReleaseGit(gitState, allowDirty, allowUnpushed); gitErr != nil {
				fmt.Printf("Git preflight: %v\n", gitErr)
				return gitErr
			}
			return planErr
		}
		if err := validateReleaseGit(gitState, allowDirty, allowUnpushed); err != nil {
			return err
		}
		if gumroadRequested {
			if cfg.Butler.Userversion == "" {
				return fmt.Errorf("Gumroad stable release requires --userversion")
			}
			if strings.TrimSpace(cfg.Gumroad.Stable[cfg.Butler.Userversion]) == "" {
				return fmt.Errorf("no Gumroad stable product configured for version %q (set gumroad.stable.%s in .fe.yaml)", cfg.Butler.Userversion, cfg.Butler.Userversion)
			}
			if _, err := exec.LookPath("gumroad"); err != nil {
				return fmt.Errorf("Gumroad CLI not found; install it and authenticate with gumroad auth login")
			}
		}
		if gitState.IsRepo {
			fmt.Printf("Release commit: %s (%s)\n", gitState.Commit, gitState.Branch)
		}
		if project, projectErr := findGodotProject("."); projectErr == nil {
			restorePath, err := activateProjectGodotPath(project)
			if err != nil {
				return err
			}
			defer restorePath()
		}
		if err := runConfiguredPreflight(cfg, "release"); err != nil {
			return err
		}

		ui.Title("Release Pipeline")
		start := time.Now()
		totalSteps := 3
		if gumroadRequested {
			totalSteps++
		}
		if changelogOutput != "" { totalSteps++ }
		if packageFormat != "" { totalSteps++ }
		stepNumber := 1

		// ── Step 1: Build/export ───────────────────────────────────────────
		ui.Step(stepNumber, totalSteps, "Build / export")
		stepNumber++
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
		// When both destinations are skipped, no upload directory is required.
		if !skipBmp || gumroadRequested {
			if err := validateReleaseBuildDir(cfg.Butler.Directory); err != nil {
				return err
			}
		}

		// ── Step 2: Pack ──────────────────────────────────────────────────
		fmt.Println()
		ui.Step(stepNumber, totalSteps, "Pack sprites")
		stepNumber++
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

		if !skipBmp || packageFormat != "" || gumroadRequested {
			commit := gitState.Commit
			if commit == "" {
				commit = strings.TrimSpace(os.Getenv("GITHUB_SHA"))
			}
			manifestPath := filepath.Join(cfg.Butler.Directory, "artifact-manifest.json")
			if err := writeArtifactManifest(cfg.Butler.Directory, manifestPath, cfg.Butler.Userversion, commit); err != nil {
				return fmt.Errorf("create artifact manifest: %w", err)
			}
		}
		if gumroadRequested {
			if err := gumroad.CheckProduct(cfg.Butler.Userversion, cfg.Butler.Directory, cfg.Gumroad.Stable, exec.Command); err != nil {
				return err
			}
		}

		if changelogOutput != "" {
			fmt.Println()
			ui.Step(stepNumber, totalSteps, "Generate changelog")
			stepNumber++
			if err := writeReleaseChangelog(changelogOutput); err != nil {
				return err
			}
			ui.OK("Wrote changelog to %s", changelogOutput)
		}

		if packageFormat != "" {
			fmt.Println()
			ui.Step(stepNumber, totalSteps, "Package build artifacts")
			stepNumber++
			if err := os.MkdirAll(filepath.Dir(packageOutput), 0755); err != nil {
				return fmt.Errorf("create package output directory: %w", err)
			}
			if err := archivepkg.Create(cfg.Butler.Directory, packageOutput, packageFormat); err != nil {
				return fmt.Errorf("package build artifacts: %w", err)
			}
			ui.OK("Created %s", packageOutput)
		}

		// ── Final step: Push ──────────────────────────────────────────────
		fmt.Println()
		ui.Step(stepNumber, totalSteps, "Push to itch.io")
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

		if gumroadRequested {
			fmt.Println()
			ui.Step(stepNumber+1, totalSteps, "Publish stable release to Gumroad")
			if _, err := gumroad.ValidateRelease(cfg.Butler.Userversion, cfg.Butler.Directory, cfg.Gumroad.Stable); err != nil {
				return err
			}
			logger := log.New(os.Stdout, "", 0)
			if err := ui.RunSpinner("Uploading stable release to Gumroad", func() error {
				return gumroad.Publish(cfg.Butler.Userversion, cfg.Butler.Directory, cfg.Gumroad.Stable, exec.Command, logger)
			}); err != nil {
				return fmt.Errorf("Gumroad publish step failed: %w", err)
			}
		}

		fmt.Println()
		ui.OK("Release complete")
		ui.Elapsed(start)
		return nil
	},
}

func releasePackagePath(output, format, version string) (string, error) {
	if output != "" {
		return output, nil
	}
	if version == "" {
		version = "dev"
	}
	directory, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join("dist", filepath.Base(directory)+"-"+version+archivepkg.FormatExtension(format)), nil
}

func writeReleaseChangelog(output string) error {
	from := ""
	if tag, err := exec.Command("git", "describe", "--tags", "--abbrev=0").Output(); err == nil {
		from = strings.TrimSpace(string(tag))
	}
	content, err := changelog.Generate(".", from, "HEAD")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return fmt.Errorf("create changelog directory: %w", err)
	}
	if err := os.WriteFile(output, []byte(content), 0644); err != nil {
		return fmt.Errorf("write changelog %q: %w", output, err)
	}
	return nil
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

func printReleasePlan(cfg *config.Config, skipBuild, skipPack, skipBmp, gumroadRequested bool, channelSuffix, changelogOutput, packageFormat, packageOutput string) error {
	fmt.Println("Release dry run — no builds, exports, packing, changelog generation, packaging, or uploads will be performed.")
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

	if changelogOutput != "" {
		fmt.Printf("Changelog: %s\n", changelogOutput)
	}
	if packageFormat != "" {
		fmt.Printf("Package: %s -> %s\n", packageFormat, packageOutput)
	}

	if !skipBmp || packageFormat != "" || gumroadRequested {
		fmt.Printf("Artifact manifest: %s\n", filepath.Join(cfg.Butler.Directory, "artifact-manifest.json"))
	}
	if skipBmp {
		fmt.Println("itch.io upload: skipped by --skip-bmp")
	} else {
		fmt.Printf("itch.io upload: %s/%s from %s\n", cfg.Itchio.Username, cfg.Itchio.Game, cfg.Butler.Directory)
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
	if gumroadRequested {
		productID := cfg.Gumroad.Stable[cfg.Butler.Userversion]
		if cfg.Butler.Userversion == "" {
			fmt.Println("Gumroad stable upload: missing release version (set --userversion)")
			problems++
		} else if productID == "" {
			fmt.Printf("Gumroad stable upload: no product configured for version %s (set gumroad.stable.%s)\n", cfg.Butler.Userversion, cfg.Butler.Userversion)
			problems++
		} else {
			fmt.Printf("Gumroad stable upload: product %s for version %s\\n", productID, cfg.Butler.Userversion)
			fmt.Println("   - files will be read from the artifact manifest after build/package steps")
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
	releaseCmd.Flags().Bool("allow-dirty", false, "allow a release from a dirty Git working tree")
	releaseCmd.Flags().Bool("allow-unpushed", false, "allow a release when the commit is not verified on a remote")
	releaseCmd.Flags().Bool("skip-build", false, "skip the build step")
	releaseCmd.Flags().Bool("skip-pack", false, "skip the sprite pack step")
	releaseCmd.Flags().Bool("skip-bmp", false, "skip the itch.io push step")
	releaseCmd.Flags().Bool("gumroad", false, "also upload this stable version to its configured Gumroad product")
	releaseCmd.Flags().String("changelog-output", "", "write Git-based release notes to this Markdown file")
	releaseCmd.Flags().String("package-format", "", "also package build artifacts as zip or tar.gz")
	releaseCmd.Flags().String("package-output", "", "package output path (defaults to dist/<project>-<version>.<format>)")
	releaseCmd.Flags().String("userversion", "", "version string for itch.io channel")
	releaseCmd.Flags().String("channel-suffix", "", "append a suffix to itch.io platform channels (for example rolling)")
	viper.BindPFlag("butler.userversion", releaseCmd.Flags().Lookup("userversion"))
}
