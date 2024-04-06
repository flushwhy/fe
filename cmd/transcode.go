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
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// transcodeCmd represents the transcode command
var transcodeCmd = &cobra.Command{
	Use:   "transcode",
	Short: "ffmpeg wrapper for transcoding audio and video.",
	Long: `You can transcode audio and video using any anything supported by ffmpeg.
	Right now it only supports transcoding audio. because most of the vars are hard coded for now.`,
	Run: func(cmd *cobra.Command, args []string) {
		inputFile := cmd.Flag("inputFile").Value.String()
		outputFile := cmd.Flag("outputFile").Value.String()

		err := ffmpeg.Input(inputFile).
			Output(outputFile, ffmpeg.KwArgs{"c:v": "libx265"}).
			OverWriteOutput().ErrorToStdOut().Run()

		if err != nil {
			fmt.Printf("Error converting %s to %s: %v", inputFile, outputFile, err)
		}
	},
}

func init() {
	rootCmd.AddCommand(transcodeCmd)
	transcodeCmd.Flags().String("inputFile", "", "input file")
	transcodeCmd.Flags().String("outputFile", "You_forgot_to_specify_an_output_file.ogg", "output file")
}
