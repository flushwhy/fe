package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/flushwhy/fe/internal/butler"
	"github.com/flushwhy/fe/internal/config"
	"github.com/flushwhy/fe/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var bmpCmd = &cobra.Command{
	Use:   "bmp",
	Short: "Push builds to itch.io via butler.",
	Long: `Pushes all platform folders inside your builds directory to itch.io
using itchio's butler tool.

Convention layout (produced by 'fe build'):
  builds/windows-x64/  →  channel windows-x64
  builds/linux-arm64/  →  channel linux-arm64

CI usage:
  export ITCHIO_USERNAME=myuser ITCHIO_GAME=mygame
  fe bmp`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if errs := cfg.Validate("bmp"); len(errs) > 0 {
			for _, e := range errs {
				ui.Fail("config: %s", e)
			}
			return fmt.Errorf("fix config errors above and try again")
		}

		ui.Title("Pushing to itch.io  %s/%s", cfg.Itchio.Username, cfg.Itchio.Game)
		start := time.Now()

		logger := log.New(os.Stdout, "", 0)
		err = ui.RunSpinner("Pushing builds via butler", func() error {
			return butler.Push(
				cfg.Itchio.Username,
				cfg.Itchio.Game,
				cfg.Butler.Directory,
				cfg.Butler.Userversion,
				exec.Command,
				logger,
			)
		})
		if err != nil {
			return err
		}

		ui.Elapsed(start)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(bmpCmd)
	bmpCmd.Flags().String("username", "", "itch.io username")
	bmpCmd.Flags().String("game", "", "itch.io game slug")
	bmpCmd.Flags().String("directory", "", "builds directory (default: builds/)")
	bmpCmd.Flags().String("userversion", "", "optional version string")
	viper.BindPFlag("itchio.username", bmpCmd.Flags().Lookup("username"))
	viper.BindPFlag("itchio.game", bmpCmd.Flags().Lookup("game"))
	viper.BindPFlag("butler.directory", bmpCmd.Flags().Lookup("directory"))
	viper.BindPFlag("butler.userversion", bmpCmd.Flags().Lookup("userversion"))
}
