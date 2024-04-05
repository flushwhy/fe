/*
Copyright © 2024 Ryan Flush <roflush@pm.me>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// bmpCmd represents the bmp command
var bmpCmd = &cobra.Command{
	Use:   "bmp",
	Short: "BMP is a wrapper for itchio's bulter tool.",
	Long: `BMP is a wrapper for itchio's bulter tool. It allows you to push all of your exports to itch.io. With one Command
    Just call BMP in the root directory of your export folder and it will push all folders to your itchio project.
    
    You have to itchio butler installed(You should use the itch.io desktop client) and make sure you add your path.
    You do have to be signed into the CLI in order for this to work.
	
	This tool is very experimental. It only supports the base butler push command. "butler push <dir> <username>/<game>:<platform>".
	`,
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
	bmpCmd.Flags().String("directory", "export", "Directory to export folder")
}

func bulter_pusher(username, game, directory string) {

	files, err := os.ReadDir(directory)
	if err != nil {
		log.Fatal("Could not read directory: ", err)
	}

	for _, f := range files {

		architecture := ""
		log.Printf("Checking for %s\n", f.Name())
		fnameLower := strings.ToLower(f.Name())
		switch fnameLower {
		case "linux", "windows", "macos", "win", "mac", "osx":
			subFiles, err := os.ReadDir(filepath.Join(directory, f.Name()))
			if err != nil {
				log.Println("Could not read subdirectory:", err)
				continue
			}

			for _, subF := range subFiles {
				log.Printf("Checking for %s\n", subF.Name())
				switch subF.Name() {
				case "x32", "x64", "arm64", "arm32", "32", "64":
					//fmt.Printf("Pushing to %s/%s:%s\n", username, game, f.Name(), subF.Name())
					architecture = subF.Name()

				case "win-x32", "win-x64", "win-arm64", "win-arm32":
					//fmt.Printf("Pushing to %s/%s:%s\n", username, game, subF.Name())
					architecture = subF.Name()
				}
				cmd := exec.Command("butler", "push", directory+"/"+f.Name()+"/"+subF.Name(), username+"/"+game+":"+f.Name()+architecture)
				fmt.Println(cmd)
				/*		err := cmd.Run()
						if err != nil {
							fmt.Println("Could not push: ", err)
							return
						} */
			}
		}
	}
}
