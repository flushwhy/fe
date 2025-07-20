/*
Copyright © 2024 Ryan Flush <roflush@pm.me>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
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
	"github.com/spf13/viper"
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

		username := viper.GetString("itchio.username")
		game := viper.GetString("itchio.game")
		directory := viper.GetString("butler.directory")
		userversion := viper.GetString("butler.userversion")

		if directory == "" {
			directory, _ = os.Getwd()
		}

		Butler_pusher(username, game, directory, userversion)
	},
}

func init() {
	rootCmd.AddCommand(bmpCmd)

	// Define Flags
	bmpCmd.Flags().String("username", "", "itch.io username")
	bmpCmd.Flags().String("game", "", "itch.io game")
	bmpCmd.Flags().String("directory", "export", "Directory to export folder")
	bmpCmd.Flags().String("userversion", "", "This is only needed if you want to use your own versioning, default itch versioning still works")

	// Bind Flags to Viper
	viper.BindPFlag("itchio.username", bmpCmd.Flags().Lookup("username"))
	viper.BindPFlag("itchio.game", bmpCmd.Flags().Lookup("game"))
	viper.BindPFlag("butler.directory", bmpCmd.Flags().Lookup("directory"))
	viper.BindPFlag("butler.userversion", bmpCmd.Flags().Lookup("userversion"))
}

func Butler_pusher(username, game, directory string, userversion string) error {
	log.Printf("Starting to push to %s/%s\n", username, game)
	log.Printf("Directory: %s\n", directory)
	log.Printf("Userversion: %s\n", userversion)

	files, err := os.ReadDir(directory)
	if err != nil {
		log.Fatalf("Could not read directory: %s", err)
	}

	for _, f := range files {
		if f == nil {
			log.Println("Skipping nil file")
			continue
		}
		log.Printf("Checking for %s\n", f.Name())

		architecture := ""
		fnameLower := strings.ToLower(f.Name())
		log.Printf("Lowercase: %s\n", fnameLower)
		switch fnameLower {
		case "linux", "windows", "macos", "win", "mac", "osx":
			subFiles, err := os.ReadDir(filepath.Join(directory, f.Name()))
			if err != nil {
				log.Printf("Could not read subdirectory: %s", err)
				continue
			}

			for _, subF := range subFiles {
				if subF == nil {
					log.Println("Skipping nil subfile")
					continue
				}
				log.Printf("Checking for %s\n", subF.Name())
				switch subF.Name() {
				case "x32", "x64", "arm64", "arm32", "32", "64":
					architecture = f.Name() + subF.Name()
				case "win-x32", "win-x64", "win-arm64", "win-arm32", "linux32":
					architecture = subF.Name()
				default:
					log.Printf("Skipping %s as it isn't a valid architecture\n", subF.Name())
					continue
				}

				cmd := exec.Command("butler", "push", filepath.Join(directory, f.Name(), subF.Name()), fmt.Sprintf("%s/%s:%s", username, game, architecture))
				if userversion != "" {
					cmd.Args = append(cmd.Args, "--userversion", userversion)
				}
				log.Printf("Running: %s\n", cmd.String())
				if err := cmd.Run(); err != nil {
					log.Fatalf("Could not push: %s", err)
				}
				log.Println("Pushed successfully")
			}
		default:
			log.Printf("Skipping %s as it isn't a valid platform\n", f.Name())
		}
	}
	return nil
}
