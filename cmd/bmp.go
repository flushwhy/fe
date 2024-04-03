/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// bmpCmd represents the bmp command
var bmpCmd = &cobra.Command{
	Use:   "bmp",
	Short: "BMP is a wrapper for itchio's bulter tool.",
	Long: `BMP is a wrapper for itchio's bulter tool. It allows you to push all of your exports to itch.io. With one Command
    Just call BMP in the root directory of your export folder and it will push all folders to your itchio project.

    You do have to be signed into the CLI inorder for this to work.`,
	Run: func(cmd *cobra.Command, args []string) {

        username := cmd.Flag("username").Value.String()
        game := cmd.Flag("game").Value.String()
        directory := cmd.Flag("directory").Value.String()
        fmt.Print(username, game, directory)

	},
}

func init() {
	rootCmd.AddCommand(bmpCmd)

	bmpCmd.Flags().String("username", "", "itch.io username")
    bmpCmd.Flags().String("game", "", "itch.io game")
    bmpCmd.Flags().String("directory", "", "Directory to export folder")
}
