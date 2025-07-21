/*
Copyright © 2025 Ryan Flush <roflush@pm.me>

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
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const defaultConfigContent = `
# Settings for the 'bmp' (butler) command
itchio:
  username: "your-itch-username"
  game: "your-itch-game-name"

# Settings for the 'pack' command
pack:
  input: "./assets/sprites"
  output: "./assets/spritesheet.png"

# Default settings for the 'transcode' command
transcode:
  codec: "libvorbis"
  bitrate: "128k"
`

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "This builds/inits a game project with a standard structure.",
	Long:  `This creates a file structure for a game project. That follows the standard fill structure.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		if _, err := os.Stat(projectName); !os.IsNotExist(err) {
			return fmt.Errorf("directory %s already exists", projectName)
		}

		fmt.Printf("🚀 Initializing new project: %s\n", projectName)

		dirsToCreate := []string{
			"assets/audio",
			"assets/fonts",
			"assets/sprites",
			"builds",
			"src",
		}

		for _, dir := range dirsToCreate {
			fullPath := filepath.Join(projectName, dir)

			if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
			}
			fmt.Printf("  ✓ Created directory: %s\n", fullPath)
		}

		configPath := filepath.Join(projectName, ".fe.yaml")

		if err := os.WriteFile(configPath, []byte(defaultConfigContent), 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}
		fmt.Printf("  ✓ created config file: %s\n", configPath)

		fmt.Printf("\n🎉 project '%s' initialization complete!", projectName)
		return nil

	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
