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
Skip individual steps with --skip-build, --skip-pack, and --skip-bmp.

CI example:
  export ITCHIO_USERNAME=myuser ITCHIO_GAME=mygame
  fe release --userversion v1.2.3 --package-format zip`,
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
		allowDirty, _ := cmd.Flags().GetBool("allow-dirty")
		allowUnpushed, _ := cmd.Flags().GetBool("allow-unpushed")
		retryMode, _ := cmd.Flags().GetString("retry")
		if retryMode != "" && retryMode != "last" {
			return fmt.Errorf("unsupported retry target %q (use --retry last)", retryMode)
		}
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
			planErr := printReleasePlan(cfg, skipBuild, skipPack, skipBmp, channelSuffix, changelogOutput, packageFormat, packageOutput)
			if gitErr := validateReleaseGit(gitState, allowDirty, allowUnpushed); gitErr != nil {
				fmt.Printf("Git preflight: %v\n", gitErr)
				return gitErr
			}
			return planErr
		}
		if err := validateReleaseGit(gitState, allowDirty, allowUnpushed); err != nil {
			return err
		}
		if gitState.IsRepo {
			fmt.Printf("Release commit: %s (%s)\n", gitState.Commit, gitState.Branch)
		}
		if err := runConfiguredPreflight(cfg, "release"); err != nil {
			return err
		}

		settingsHash, err := releaseSettingsHash(cfg, skipBuild, skipPack, skipBmp, channelSuffix, changelogOutput, packageFormat, packageOutput)
		if err != nil { return err }
		var journal *releaseJournal
		if retryMode == "last" {
			if cfg.Butler.Userversion == "" {
				return fmt.Errorf("--retry last requires the original --userversion")
			}
			journal, err = loadLastReleaseJournal()
			if err != nil { return err }
			if journal.Complete {
				return fmt.Errorf("the last release already completed; start a new release instead")
			}
			if err := validateRetryJournal(journal, cfg.Butler.Userversion, gitState, channelSuffix, cfg.Butler.Directory, packageFormat, packageOutput); err != nil {
				return err
			}
			if journal.SettingsHash != settingsHash {
				return fmt.Errorf("release configuration changed since the previous attempt; restore the original build, pack, publisher, and output settings before retrying")
			}
			if journal.BuildTargets == nil { journal.BuildTargets = map[string]bool{} }
			if journal.Stages == nil { journal.Stages = map[string]bool{} }
			if journal.ButlerChannels == nil { journal.ButlerChannels = map[string]bool{} }
			fmt.Printf("Resuming release %s from commit %s\n", journal.Version, journal.Commit)
		} else if cfg.Butler.Userversion != "" {
			if err := ensureNoExistingReleaseJournal(cfg.Butler.Userversion); err != nil { return err }
			journal = newReleaseJournal(cfg.Butler.Userversion, gitState, channelSuffix, cfg.Butler.Directory, packageFormat, packageOutput)
			journal.SettingsHash = settingsHash
			if err := journal.save(); err != nil { return err }
		}
		stageDone := func(name string) bool { return journal != nil && journal.Stages[name] }
		markStageDone := func(name string) error {
			if journal == nil { return nil }
			journal.Stages[name] = true
			return journal.save()
		}

		ui.Title("Release Pipeline")
		start := time.Now()
		totalSteps := 3
		if changelogOutput != "" { totalSteps++ }
		if packageFormat != "" { totalSteps++ }
		stepNumber := 1

		// ── Step 1: Build/export ───────────────────────────────────────────
		ui.Step(stepNumber, totalSteps, "Build / export")
		stepNumber++
		if !skipBuild {
			if _, statErr := os.Stat("project.godot"); statErr == nil {
				if !stageDone("build:godot") {
					if err := godot.Export(".", cfg.Godot.Executable, cfg.Godot.OutputDir, "", exec.Command); err != nil {
						return fmt.Errorf("Godot export step failed: %w", err)
					}
					if err := markStageDone("build:godot"); err != nil { return err }
				} else { ui.Dim("Godot export already complete") }
			} else if len(cfg.Build.Targets) == 0 {
				ui.Warn("No build targets defined — skipping build step")
			} else {
				for _, target := range cfg.Build.Targets {
					key := releaseBuildTargetKey(target.Name, target.Platform, target.Cmd)
					if journal != nil && journal.BuildTargets[key] {
						ui.Dim("already built: %s", target.Platform)
						continue
					}
					if err := ui.RunSpinner(fmt.Sprintf("Building %s", target.Platform), func() error {
						return build.RunTarget(target, build.DefaultRunner)
					}); err != nil {
						return fmt.Errorf("build step failed: %w", err)
					}
					if journal != nil {
						journal.BuildTargets[key] = true
						if err := journal.save(); err != nil { return err }
					}
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
		ui.Step(stepNumber, totalSteps, "Pack sprites")
		stepNumber++
		if skipPack {
			ui.Dim("skipped")
		} else if !releaseSpriteInputsAvailable(cfg.Pack) {
			ui.Warn("No configured sprite input directory found — skipping sprite packing (software-only release). Add sprites or configure pack.input / pack.sheets in .fe.yaml if you want a spritesheet.")
		} else if stageDone("pack") {
			ui.Dim("sprite packing already complete")
		} else {
			err := ui.RunSpinner(fmt.Sprintf("Packing %s → %s", cfg.Pack.Input, cfg.Pack.Output), func() error {
				return pack.Run(cfg.Pack)
			})
			if err != nil {
				return fmt.Errorf("pack step failed; refusing to publish an incomplete release: %w", err)
			}
			if err := markStageDone("pack"); err != nil { return err }
		}

		if !skipBmp || packageFormat != "" {
			commit := gitState.Commit
			if commit == "" {
				commit = strings.TrimSpace(os.Getenv("GITHUB_SHA"))
			}
			if stageDone("manifest") {
				manifestPath := filepath.Join(cfg.Butler.Directory, "artifact-manifest.json")
				if err := verifyArtifactManifest(cfg.Butler.Directory, manifestPath, cfg.Butler.Userversion, commit); err != nil {
					return fmt.Errorf("completed artifacts are no longer valid; refusing to resume publication: %w", err)
				}
				ui.Dim("artifact manifest and completed artifacts verified")
			} else {
				manifestPath := filepath.Join(cfg.Butler.Directory, "artifact-manifest.json")
				if err := writeArtifactManifest(cfg.Butler.Directory, manifestPath, cfg.Butler.Userversion, commit); err != nil {
					return fmt.Errorf("create artifact manifest: %w", err)
				}
				if err := markStageDone("manifest"); err != nil { return err }
			}
		}

		if changelogOutput != "" {
			fmt.Println()
			ui.Step(stepNumber, totalSteps, "Generate changelog")
			stepNumber++
			if stageDone("changelog") {
				ui.Dim("changelog already complete")
			} else {
				if err := writeReleaseChangelog(changelogOutput); err != nil {
					return err
				}
				if err := markStageDone("changelog"); err != nil { return err }
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
			if stageDone("package") {
				if err := verifyReleasePackage(journal, packageOutput); err != nil {
					return fmt.Errorf("completed package archive is no longer valid: %w", err)
				}
				ui.Dim("package archive verified; skipping completed stage")
			} else {
				if err := archivepkg.Create(cfg.Butler.Directory, packageOutput, packageFormat); err != nil {
					return fmt.Errorf("package build artifacts: %w", err)
				}
				if journal != nil {
					journal.PackageSize, journal.PackageSHA256, err = releasePackageIntegrity(packageOutput)
					if err != nil {
						return err
					}
				}
				if err := markStageDone("package"); err != nil { return err }
				ui.OK("Created %s", packageOutput)
			}
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
					if journal == nil {
						return butler.Push(cfg.Itchio.Username, cfg.Itchio.Game, cfg.Butler.Directory, cfg.Butler.Userversion, exec.Command, logger, channelSuffix)
					}
					return butler.PushWithProgress(
						cfg.Itchio.Username, cfg.Itchio.Game, cfg.Butler.Directory, cfg.Butler.Userversion,
						exec.Command, logger, journal.ButlerChannels,
						func(channel string) error {
							journal.ButlerChannels[channel] = true
							return journal.save()
						}, channelSuffix,
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
		if journal != nil {
			journal.Complete = true
			if err := journal.save(); err != nil { return err }
		}
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

func printReleasePlan(cfg *config.Config, skipBuild, skipPack, skipBmp bool, channelSuffix, changelogOutput, packageFormat, packageOutput string) error {
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

	if !skipBmp || packageFormat != "" {
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
	releaseCmd.Flags().String("retry", "", "resume the last incomplete release (use: last)")
	releaseCmd.Flags().Bool("skip-build", false, "skip the build step")
	releaseCmd.Flags().Bool("skip-pack", false, "skip the sprite pack step")
	releaseCmd.Flags().Bool("skip-bmp", false, "skip the itch.io push step")
	releaseCmd.Flags().String("changelog-output", "", "write Git-based release notes to this Markdown file")
	releaseCmd.Flags().String("package-format", "", "also package build artifacts as zip or tar.gz")
	releaseCmd.Flags().String("package-output", "", "package output path (defaults to dist/<project>-<version>.<format>)")
	releaseCmd.Flags().String("userversion", "", "version string for itch.io channel")
	releaseCmd.Flags().String("channel-suffix", "", "append a suffix to itch.io platform channels (for example rolling)")
	viper.BindPFlag("butler.userversion", releaseCmd.Flags().Lookup("userversion"))
}
