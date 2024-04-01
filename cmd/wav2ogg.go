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

// wav2oggCmd represents the wav2ogg command
var wav2oggCmd = &cobra.Command{
	Use:   "wav2ogg",
	Short: "takes a wav file and converts it to an ogg file ",
	Long: `Converts a wav file to an ogg file. In the future, this will also be able to convert a folder full of wav files to ogg.
	--input <input file>.wav --output <output file>.ogg`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("wav2ogg called")
	},
}

func init() {
	audioCmd.AddCommand(wav2oggCmd)

	wav2oggCmd.Flags().String("input", "", "input file")
	wav2oggCmd.Flags().String("output", "You_forgot_to_specify_an_output_file.ogg", "output file")
}
