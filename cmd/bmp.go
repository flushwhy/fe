/*
Copyright © 2024 Ryan Flush <roflush@pm.me>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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

		if directory == "" {
			directory, _ = os.Getwd()
		}

		bulter_pusher(username, game, directory)
	},
}

func init() {
	rootCmd.AddCommand(bmpCmd)

	bmpCmd.Flags().String("username", "", "itch.io username")
	bmpCmd.Flags().String("game", "", "itch.io game")
	bmpCmd.Flags().String("directory", "", "Directory to export folder")
}

func bulter_pusher(username, game, directory string) {
	fmt.Println("Pushing " + directory + " to " + username + "/" + game)

	dirs, err := filepath.Glob(filepath.Join(directory + "*"))
	if err != nil {
		fmt.Println("Error reading export directory: ", err)
		return
	}

	for _, dir := range dirs {
		fileInfo, err := os.Stat(dir)
		if err != nil {
			fmt.Println("Error reading export directory: ", err)
			continue
		}

		if fileInfo.IsDir() {
			fmt.Println("Pushing to " + username + "/" + game + "/" + fileInfo.Name())
		}
	}

}
