package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "fe",
	Short: "Fe — dev middleware for the awkward parts of shipping software",
	Long: `Fe is a local CLI and CI-friendly middleware layer.

  Project scaffolding:
    fe init <lang> --name <project>   scaffold a new project (odin, c, cpp, zig, go)
    fe add <tool>                     wire up LSP/editor tooling (clangd, ols, zls, gopls)
    fe doctor                         check tools are installed and configs are correct

  Shipping pipeline:
    fe build                          compile for all configured platforms
    fe pack                           pack sprites into a spritesheet
    fe transcode                      convert audio/video via ffmpeg
    fe bmp                            push builds to itch.io via butler
    fe release                        build → pack → push in one shot
    fe validate                       pre-flight config and path check

  Interactive:
    fe tui                            Bubble Tea UI (local terminals only)

Config priority: env vars > .fe.yaml > conventions > defaults`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: .fe.yaml in cwd or $HOME)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		cwd, _ := os.Getwd()
		viper.AddConfigPath(cwd)
		home, _ := os.UserHomeDir()
		if home != "" {
			viper.AddConfigPath(home)
		}
		viper.SetConfigType("yaml")
		viper.SetConfigName(".fe")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "config:", viper.ConfigFileUsed())
	}
}
