package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/flushwhy/fe/internal/o3de"
	"github.com/flushwhy/fe/internal/tty"
	"github.com/flushwhy/fe/internal/ui"
	"github.com/spf13/cobra"
)

var o3deBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build an already-configured O3DE CMake build directory.",
	Long: `Runs cmake --build for an existing build directory and adds concise hints when compilation fails.
It does not configure projects or replace O3DE's project/build setup workflow.

Example:
  fe o3de build --build-dir build`,
	RunE: func(cmd *cobra.Command, args []string) error {
		buildDir, _ := cmd.Flags().GetString("build-dir")
		info, err := os.Stat(buildDir)
		if err != nil {
			return fmt.Errorf("build directory %q is unavailable: %w; configure the project using O3DE's documented workflow first", buildDir, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("build path %q is not a directory", buildDir)
		}
		if _, err := os.Stat(strings.TrimSpace(buildDir) + string(os.PathSeparator) + "CMakeCache.txt"); err != nil {
			return fmt.Errorf("%q does not contain CMakeCache.txt; this directory may not be configured yet. Use O3DE's documented project configure workflow first", buildDir)
		}

		ui.Title("Building O3DE project")
		ui.Dim("CMake build directory: %s", buildDir)
		var output bytes.Buffer
		var writer io.Writer = &output
		if !tty.IsInteractive() {
			writer = io.MultiWriter(os.Stdout, &output)
		}
		config, _ := cmd.Flags().GetString("config")
		targets, _ := cmd.Flags().GetStringArray("target")
		for _, target := range targets {
			if strings.TrimSpace(target) == "" {
				return fmt.Errorf("--target values cannot be empty")
			}
		}
		build := exec.CommandContext(cmd.Context(), "cmake", o3deBuildArgs(buildDir, config, targets)...)
		build.Stdout = writer
		build.Stderr = writer

		var buildErr error
		if tty.IsInteractive() {
			buildErr = ui.RunSpinner("Building O3DE project", build.Run)
		} else {
			buildErr = build.Run()
		}
		if buildErr != nil {
			showO3DEBuildFailure(buildDir, buildErr, output.String())
			return fmt.Errorf("cmake --build %q: %w", buildDir, buildErr)
		}
		ui.OK("O3DE build completed")
		return nil
	},
}

func init() {
	o3deCmd.AddCommand(o3deBuildCmd)
	o3deBuildCmd.Flags().String("build-dir", "build", "existing CMake build directory")
	o3deBuildCmd.Flags().String("config", "", "configuration for multi-config generators (for example Debug, Profile, or Release)")
	o3deBuildCmd.Flags().StringArray("target", nil, "CMake target to build (repeatable)")
}

func o3deBuildArgs(buildDir, config string, targets []string) []string {
	args := []string{"--build", buildDir, "--parallel"}
	if strings.TrimSpace(config) != "" {
		args = append(args, "--config", config)
	}
	if len(targets) > 0 {
		args = append(args, "--target")
		args = append(args, targets...)
	}
	return args
}
