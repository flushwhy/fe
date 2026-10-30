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
	Short: "Compile for all configured platforms.",
	Long: `Runs the build commands defined in .fe.yaml build.targets.

Example .fe.yaml:
  build:
    targets:
      - platform: linux-x64
        cmd: "gcc -o builds/linux-x64/game src/main.c"
      - platform: windows-x64
        cmd: "x86_64-w64-mingw32-gcc -o builds/windows-x64/game.exe src/main.c"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if err := runConfiguredPreflight(cfg, "build"); err != nil {
			return err
		}
		containerBuild, _ := cmd.Flags().GetBool("container")
		if containerBuild {
			runtimeName, _ := cmd.Flags().GetString("runtime")
			if err := build.RunContainer(cfg.Build.Container, runtimeName); err != nil {
				return err
			}
			return nil
		}
		if errs := cfg.Validate("build"); len(errs) > 0 {
			for _, e := range errs {
				ui.Fail("config: %s", e)
			}
			return fmt.Errorf("fix config errors above and try again")
		}

		platform, _ := cmd.Flags().GetString("platform")
		ui.Title("Building targets")
		start := time.Now()

		targets := cfg.Build.Targets
		if platform != "" {
			targets = build.FilterByPlatform(cfg.Build.Targets, platform)
			if len(targets) == 0 {
				return fmt.Errorf("no target found for platform %q", platform)
			}
		}

		steps := make([]ui.StepTask, len(targets))
		for i, t := range targets {
			t := t // capture
			steps[i] = ui.StepTask{
				Label: fmt.Sprintf("Building %s", t.Platform),
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

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().String("platform", "", "build only this platform (e.g. linux-x64)")
	buildCmd.Flags().Bool("container", false, "build in a temporary container")
	buildCmd.Flags().String("runtime", "docker", "container runtime to use: docker or podman")
	viper.BindPFlag("build.platform", buildCmd.Flags().Lookup("platform"))
	viper.BindPFlag("build.container.runtime", buildCmd.Flags().Lookup("runtime"))
}
