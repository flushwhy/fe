package cmd

import (
	"fmt"
	"time"

	"github.com/flushwhy/fe/internal/build"
	"github.com/flushwhy/fe/internal/config"
	projectinspect "github.com/flushwhy/fe/internal/inspect"
	"github.com/flushwhy/fe/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build configured targets or a safe detected host target.",
	Long: `Builds targets from .fe.yaml. If no targets are configured, Fe can
suggest a conventional host build for supported project types.

Example .fe.yaml:
  build:
    targets:
      - platform: linux-x64
        cmd: gcc
        args: [-o, builds/linux-x64/game, src/main.c]
      - platform: windows-x64
        cmd: x86_64-w64-mingw32-gcc
        args: [-o, builds/windows-x64/game.exe, src/main.c]`,
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
			return build.RunContainer(cfg.Build.Container, runtimeName)
		}

		platform, _ := cmd.Flags().GetString("platform")
		var suggestions []config.BuildTarget
		if len(cfg.Build.Targets) == 0 {
			report, err := projectinspect.Inspect(".")
			if err != nil {
				return err
			}
			for _, suggestion := range report.Suggestions {
				suggestions = append(suggestions, config.BuildTarget{
					Cmd:  suggestion.Command,
					Args: suggestion.Args,
				})
				fmt.Fprintf(cmd.OutOrStdout(), "Detected %s project; suggesting %s %v\n",
					suggestion.ProjectType, suggestion.Command, suggestion.Args)
			}
		}

		plan, err := build.PlanWithSuggestions(cfg.Build, build.Options{Platform: platform}, suggestions)
		if err != nil {
			return err
		}
		if err := build.CheckTools(plan); err != nil {
			return err
		}

		ui.Title("Building targets")
		start := time.Now()
		steps := make([]ui.StepTask, len(plan))
		for i, target := range plan {
			target := target
			steps[i] = ui.StepTask{
				Label: fmt.Sprintf("Building %s", target.Platform),
				Run: func() error {
					return build.RunPlannedTarget(target, build.DefaultRunner)
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
