/*
Copyright © 2024 Ryan Flush <roflush@pm.me>
*/
package cmd

import (
	"fmt"
	"os"   
    "log"
	

	"github.com/spf13/cobra"
)

// bmpCmd represents the bmp command
var bmpCmd = &cobra.Command{
	Use:   "bmp",
	Short: "BMP is a wrapper for itchio's bulter tool.",
	Long: `BMP is a wrapper for itchio's bulter tool. It allows you to push all of your exports to itch.io. With one Command
    Just call BMP in the root directory of your export folder and it will push all folders to your itchio project.
    
    You have to itchio buter installed and in your path.
    You do have to be signed into the CLI in order for this to work.`,
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
    
	files, err := os.ReadDir(directory)
	if err != nil {
		log.Fatal("Could not read directory: ", err)
	}

	for _, f := range files {
       fmt.Println("Pushing  to " + username + "/" + game + ":" + directory + "-" + f.Name())
	}

}
