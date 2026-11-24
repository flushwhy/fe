package cmd

import (
	"fmt"
	"time"

	"github.com/flushwhy/fe/internal/pack"
	"github.com/flushwhy/fe/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var packCmd = &cobra.Command{
	Use:   "pack",
	Short: "Pack PNG sprites into one or more spritesheets.",
	Long: `Reads PNG files from the input directory and stitches them horizontally.
By default, it creates one spritesheet. Configure pack.sheets in .fe.yaml
for separate sheets from named input directories.

Defaults (from convention):
  input:  assets/sprites/
  output: assets/spritesheet.png

Example:
  pack:
    sheets:
      player:
        input: assets/sprites/player
        output: assets/player.png
      npcs:
        input: assets/sprites/npcs
        output: assets/npcs.png`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadProjectConfig()
		if err != nil {
			return err
		}
		if err := runConfiguredPreflight(cfg, "package"); err != nil {
			return err
		}
		if errs := cfg.Validate("pack"); len(errs) > 0 {
			return &ui.CommandError{
				Message: "sprite packing configuration is invalid",
				Details: errs,
				Hint:    "correct the pack settings in .fe.yaml, then run fe validate",
			}
		}

		ui.Title("Packing sprites")
		start := time.Now()

		label := fmt.Sprintf("Packing %s → %s", cfg.Pack.Input, cfg.Pack.Output)
		if len(cfg.Pack.Sheets) > 0 {
			label = fmt.Sprintf("Packing %d spritesheets", len(cfg.Pack.Sheets))
		}
		err = ui.RunSpinner(label, func() error {
			return pack.Run(cfg.Pack)
		})
		if err != nil {
			return &ui.CommandError{
				Message: "sprite packing failed",
				Cause:   err,
				Hint:    "check that the input directory contains readable PNG files and that the output directory is writable",
			}
		}

		ui.Elapsed(start)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(packCmd)
	packCmd.Flags().String("input", "", "directory containing PNG sprites")
	packCmd.Flags().String("output", "", "output spritesheet path")
	viper.BindPFlag("pack.input", packCmd.Flags().Lookup("input"))
	viper.BindPFlag("pack.output", packCmd.Flags().Lookup("output"))
}
