package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/flushwhy/fe/internal/config"
	"github.com/flushwhy/fe/internal/godot"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export all configured Godot presets.",
	Long: `Exports every preset in export_presets.cfg to Butler-compatible folders under builds/.

Godot must be installed and its export templates/platform requirements must be ready.
Use --preset to export one named preset.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		projectDir, err := os.Getwd()
		if err != nil {
			return err
		}
		preset, _ := cmd.Flags().GetString("preset")
		return godot.Export(projectDir, cfg.Godot.Executable, cfg.Godot.OutputDir, preset, exec.Command)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().String("preset", "", "export only this exact Godot preset name")
}

var _ = fmt.Sprintf
