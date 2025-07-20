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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// transcodeCmd represents the transcode command
var transcodeCmd = &cobra.Command{
	Use:   "transcode",
	Short: "ffmpeg wrapper for transcoding audio and video.",
	Long: `You can transcode audio and video using any anything supported by ffmpeg.
	Specify the input file, output file, codec, and more.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the main input/output files
		inputFilename := viper.GetString("transcode.inputFile")
		outputFilename := viper.GetString("transcode.outputFile")

		if inputFilename == "" {
			log.Fatal("Error: --inputFile must be set via flag or config file.")
		}

		// Define the ffmpeg arguments by their proper names
		ffmpegArgs := map[string]string{
			"-c:v": viper.GetString("transcode.codec"),
			"-b:v": viper.GetString("transcode.bitrate"),
			"-ac":  viper.GetString("transcode.audioChannels"),
			"-r":   viper.GetString("transcode.videoFrameRate"),
			"-s":   viper.GetString("transcode.videoResolution"),
			"-ss":  viper.GetString("transcode.startTime"),
			"-t":   viper.GetString("transcode.endTime"),
		}

		// Build the final map for the ffmpeg-go library, skipping empty values
		outputArgs := make(map[string]interface{})
		for key, val := range ffmpegArgs {
			if val != "" {
				outputArgs[key] = val
			}
		}

		// Run the ffmpeg command
		err := ffmpeg.Input(inputFilename).
			Output(outputFilename, outputArgs).
			OverWriteOutput().ErrorToStdOut().Run()

		if err != nil {
			fmt.Printf("Error converting %s to %s: %v\n", inputFilename, outputFilename, err)
		} else {
			fmt.Printf("Successfully transcoded %s to %s\n", inputFilename, outputFilename)
		}
	},
}

func init() {
	rootCmd.AddCommand(transcodeCmd)

	transcodeCmd.Flags().String("inputFile", "", "input file")
	transcodeCmd.Flags().String("outputFile", "You_forgot_to_specify_an_output_file.ogg", "output file")
	transcodeCmd.Flags().String("codec", "", "codec")
	transcodeCmd.Flags().String("bitrate", "", "bitrate")
	transcodeCmd.Flags().String("audioChannels", "", "audio channels")
	transcodeCmd.Flags().String("videoFrameRate", "", "video frame rate")
	transcodeCmd.Flags().String("videoResolution", "", "video resolution")
	transcodeCmd.Flags().String("startTime", "", "start time")
	transcodeCmd.Flags().String("endTime", "", "end time")

	viper.BindPFlag("transcode.inputFile", transcodeCmd.Flags().Lookup("inputFile"))
	viper.BindPFlag("transcode.outputFile", transcodeCmd.Flags().Lookup("outputFile"))
	viper.BindPFlag("transcode.codec", transcodeCmd.Flags().Lookup("codec"))
	viper.BindPFlag("transcode.bitrate", transcodeCmd.Flags().Lookup("bitrate"))
	viper.BindPFlag("transcode.audioChannels", transcodeCmd.Flags().Lookup("audioChannels"))
	viper.BindPFlag("transcode.videoFrameRate", transcodeCmd.Flags().Lookup("videoFrameRate"))
	viper.BindPFlag("transcode.videoResolution", transcodeCmd.Flags().Lookup("videoResolution"))
	viper.BindPFlag("transcode.startTime", transcodeCmd.Flags().Lookup("startTime"))
	viper.BindPFlag("transcode.endTime", transcodeCmd.Flags().Lookup("endTime"))
}
