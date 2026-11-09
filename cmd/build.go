package cmd

import (
	"fmt"
	"time"

	"github.com/flushwhy/fe/internal/build"
	"github.com/flushwhy/fe/internal/config"
	"github.com/flushwhy/fe/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build all configured targets or select one by name/platform.",
	Long: `Runs the build commands defined in .fe.yaml build.targets.

Example .fe.yaml:
  build:
    targets:
      - name: linux-release
        platform: linux-x64
        cmd: "gcc -o builds/linux-x64/game src/main.c"
      - name: windows-release
        platform: windows-x64
        cmd: "x86_64-w64-mingw32-gcc -o builds/windows-x64/game.exe src/main.c"

Use --target to select a named target, --platform to build every target for
one platform, or --list-targets to inspect the configured targets. Existing
platform-only configurations remain supported.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		target, _ := cmd.Flags().GetString("target")
		platform, _ := cmd.Flags().GetString("platform")
		listTargets, _ := cmd.Flags().GetBool("list-targets")
		if listTargets {
			if len(cfg.Build.Targets) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No build targets configured.")
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%-24s %-20s %s\n", "NAME", "PLATFORM", "OUTPUT")
			for _, t := range cfg.Build.Targets {
				name := t.Name
				if name == "" {
					name = "(unnamed)"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-24s %-20s %s\n", name, t.Platform, t.Output)
			}
			return nil
		}

		containerBuild, _ := cmd.Flags().GetBool("container")
		if containerBuild {
			if target != "" || platform != "" {
				return fmt.Errorf("--target and --platform cannot be used with --container")
			}
			if err := runConfiguredPreflight(cfg, "build"); err != nil {
				return err
			}
			runtimeName, _ := cmd.Flags().GetString("runtime")
			return build.RunContainer(cfg.Build.Container, runtimeName)
		}

		if project, projectErr := findGodotProject("."); projectErr == nil {
			restorePath, err := activateProjectGodotPath(project)
			if err != nil {
				return err
			}
			defer restorePath()
		}

		if errs := cfg.Validate("build"); len(errs) > 0 {
			for _, e := range errs {
				ui.Fail("config: %s", e)
			}
			return fmt.Errorf("fix config errors above and try again")
		}

		targets, err := selectBuildTargets(cfg.Build.Targets, target, platform)
		if err != nil {
			return err
		}
		// Validate configuration and selection before running user-configured checks.
		if err := runConfiguredPreflight(cfg, "build"); err != nil {
			return err
		}

		ui.Title("Building targets")
		start := time.Now()
		steps := make([]ui.StepTask, len(targets))
		for i, t := range targets {
			t := t
			label := t.Name
			if label == "" {
				label = t.Platform
			} else if t.Platform != "" {
				label += " (" + t.Platform + ")"
			}
			steps[i] = ui.StepTask{
				Label: fmt.Sprintf("Building %s", label),
				Run: func() error {
					return build.RunTarget(t, build.DefaultRunner)
				},
			}
		}

		if err := ui.RunSteps(steps); err != nil {
			return err
		}

		ui.Elapsed(start)
		return nil
	},
}

func selectBuildTargets(configured []config.BuildTarget, target, platform string) ([]config.BuildTarget, error) {
	if target != "" && platform != "" {
		return nil, fmt.Errorf("--target and --platform cannot be used together")
	}
	if target != "" {
		selected := build.FilterByName(configured, target)
		if len(selected) == 0 {
			return nil, fmt.Errorf("no build target named %q; use --list-targets to see named targets", target)
		}
		if len(selected) > 1 {
			return nil, fmt.Errorf("build target name %q is not unique", target)
		}
		return selected, nil
	}
	if platform != "" {
		selected := build.FilterByPlatform(configured, platform)
		if len(selected) == 0 {
			return nil, fmt.Errorf("no target found for platform %q", platform)
		}
		return selected, nil
	}
	return configured, nil
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().String("target", "", "build one named target from build.targets")
	buildCmd.Flags().String("platform", "", "build all targets for this platform (e.g. linux-x64)")
	buildCmd.Flags().Bool("list-targets", false, "list configured build targets without running them")
	buildCmd.Flags().Bool("container", false, "build in a temporary container")
	buildCmd.Flags().String("runtime", "docker", "container runtime to use: docker or podman")
	viper.BindPFlag("build.target", buildCmd.Flags().Lookup("target"))
	viper.BindPFlag("build.platform", buildCmd.Flags().Lookup("platform"))
	viper.BindPFlag("build.container.runtime", buildCmd.Flags().Lookup("runtime"))
}
