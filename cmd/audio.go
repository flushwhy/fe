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

	"github.com/spf13/cobra"
)

// audioCmd represents the audio command
var audioCmd = &cobra.Command{
	Use:   "audio",
	Short: "This is the top level command for audio work, it does a cool read of audio files",
	Long: `todo: add a description here.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("audio called")
	},
}

func init() {
	rootCmd.AddCommand(audioCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// audioCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// audioCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}