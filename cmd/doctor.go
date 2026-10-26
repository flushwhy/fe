package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flushwhy/fe/internal/config"

	"github.com/flushwhy/fe/internal/doctor"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check that required tools are installed and configs are correct.",
	Long: `Inspects the current project directory, detects the language, and checks:

  • Required binaries are in PATH (clangd, bear, ols, zls, gopls, etc.)
  • LSP config files exist and are valid (.clangd, ols.json, zls.json)
  • compile_commands.json exists and is not an empty stub (C/C++)
  • cmake has EXPORT_COMPILE_COMMANDS set (C++ cmake projects)
  • go.mod exists (Go projects)

Exit code 0 if all checks pass. Non-zero if anything fails (CI-safe).

Run 'fe add <tool>' to fix most issues.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		results := doctor.Run()
		results = append(results, feProjectChecks()...)
		allOK := doctor.Print(results)
		if !allOK {
			return fmt.Errorf("doctor found one or more problems")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func feProjectChecks() []doctor.Result {
	var results []doctor.Result
	cfg, err := config.Load()
	if err != nil {
		return []doctor.Result{{Name: "Fe config", OK: false, Message: err.Error(), Fix: "check .fe.yaml syntax and field types"}}
	}
	for i, target := range cfg.Build.Targets {
		if target.Platform == "" || target.Cmd == "" {
			results = append(results, doctor.Result{
				Name: fmt.Sprintf("build.targets[%d]", i),
				OK: false,
				Message: "platform and cmd must both be set",
				Fix: "set platform and cmd for each build target in .fe.yaml",
			})
		}
	}
	if cfg.Pack.Input != "" && cfg.Pack.Input != "assets/sprites" {
		if info, err := os.Stat(cfg.Pack.Input); err != nil || !info.IsDir() {
			results = append(results, doctor.Result{
				Name: "pack input",
				OK: false,
				Message: fmt.Sprintf("configured path %q is not an accessible directory", cfg.Pack.Input),
				Fix: "correct pack.input in .fe.yaml or create the directory",
			})
		}
	}
	if _, err := os.Stat("project.godot"); err == nil {
		path, err := exec.LookPath(cfg.Godot.Executable)
		if err != nil {
			results = append(results, doctor.Result{
				Name: "Godot executable",
				OK: false,
				Message: fmt.Sprintf("%q not found in PATH", cfg.Godot.Executable),
				Fix: "install Godot or set godot.executable in .fe.yaml",
			})
		} else {
			results = append(results, doctor.Result{Name: "Godot executable", OK: true, Message: path})
		}
		if _, err := os.Stat(filepath.Join(".", "export_presets.cfg")); err != nil {
			results = append(results, doctor.Result{
				Name: "Godot export presets",
				OK: true,
				Message: "not configured (only required for exports)",
				Fix: "create export presets in the Godot editor when you are ready to export",
			})
		} else {
			results = append(results, doctor.Result{Name: "Godot export presets", OK: true, Message: "export_presets.cfg found"})
		}
	}
	return results
}
