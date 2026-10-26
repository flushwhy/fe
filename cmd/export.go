package cmd

import (
	"os"
	"os/exec"
	"fmt"
	"path/filepath"

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
		listPresets, _ := cmd.Flags().GetBool("list-presets")
		if listPresets {
			presets, err := godot.ReadPresets(filepath.Join(projectDir, "export_presets.cfg"))
			if err != nil {
				return err
			}
			fmt.Printf("%-28s %-20s %s\n", "PRESET", "PLATFORM", "OUTPUT")
			for _, p := range presets {
				output, mapErr := godot.ExpectedOutputPath(projectDir, cfg.Godot.OutputDir, p)
				if mapErr != nil {
					output = p.ExportPath + " (unsupported platform)"
				}
				fmt.Printf("%-28s %-20s %s\n", p.Name, p.Platform, output)
			}
			return nil
		}
		return godot.Export(projectDir, cfg.Godot.Executable, cfg.Godot.OutputDir, preset, exec.Command)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().String("preset", "", "export only this exact Godot preset name")
}

