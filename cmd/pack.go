package cmd

import (
	"fmt"
	"time"

	"github.com/flushwhy/fe/internal/config"
	"github.com/flushwhy/fe/internal/pack"
	"github.com/flushwhy/fe/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var packCmd = &cobra.Command{
	Use:   "pack",
	Short: "Pack PNG sprites into a single spritesheet.",
	Long: `Reads all PNG files from the input directory and stitches them
horizontally into a single spritesheet PNG.

Defaults (from convention):
  input:  assets/sprites/
  output: assets/spritesheet.png`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if errs := cfg.Validate("pack"); len(errs) > 0 {
			for _, e := range errs {
				ui.Fail("config: %s", e)
			}
			return fmt.Errorf("missing required config")
		}

		ui.Title("Packing sprites")
		start := time.Now()

		err = ui.RunSpinner(fmt.Sprintf("Packing %s → %s", cfg.Pack.Input, cfg.Pack.Output), func() error {
			return pack.Run(cfg.Pack)
		})
		if err != nil {
			return err
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
