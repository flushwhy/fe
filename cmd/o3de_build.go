package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/flushwhy/fe/internal/o3de"
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
		writer := io.MultiWriter(os.Stdout, &output)
		build := exec.CommandContext(cmd.Context(), "cmake", "--build", buildDir, "--parallel")
		build.Stdout = writer
		build.Stderr = writer
		if err := build.Run(); err != nil {
			ui.Fail("CMake build failed")
			fmt.Println()
			fmt.Println(ui.StyleBold.Render("What to check"))
			for _, hint := range o3de.SummarizeBuildFailure(output.String()) {
				fmt.Printf("  %s %s\n", ui.StyleWarn.Render(ui.IconArrow), hint)
			}
			return fmt.Errorf("cmake --build %q: %w", buildDir, err)
		}
		ui.OK("O3DE build completed")
		return nil
	},
}

func init() {
	o3deCmd.AddCommand(o3deBuildCmd)
	o3deBuildCmd.Flags().String("build-dir", "build", "existing CMake build directory")
}
