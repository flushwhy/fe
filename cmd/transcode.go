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
	Specify the input file, output file, codec, and more.`,
	Run: func(cmd *cobra.Command, args []string) {
		inputFilename := cmd.Flag("inputFile").Value.String()
		outputFilename := cmd.Flag("outputFile").Value.String()
		ffmpegArgs := make([]string, 0)

		addFfmpegArg := func(flagName, flagValue string) {
			if flagValue != "" {
				ffmpegArgs = append(ffmpegArgs, "-"+flagName, flagValue)
			}
		}

		addFfmpegArg("c:v", cmd.Flag("codec").Value.String())
		addFfmpegArg("b:v", cmd.Flag("bitrate").Value.String())
		addFfmpegArg("ac", cmd.Flag("audioChannels").Value.String())
		addFfmpegArg("r", cmd.Flag("videoFrameRate").Value.String())
		addFfmpegArg("s", cmd.Flag("videoResolution").Value.String())
		addFfmpegArg("ss", cmd.Flag("startTime").Value.String())
		addFfmpegArg("t", cmd.Flag("endTime").Value.String())

		argsMap := make(map[string]interface{})
		for i := 0; i < len(ffmpegArgs); i += 2 {
			argsMap[ffmpegArgs[i]] = ffmpegArgs[i+1]
		}

		err := ffmpeg.Input(inputFilename).
			Output(outputFilename, argsMap).
			OverWriteOutput().ErrorToStdOut().Run()

		if err != nil {
			fmt.Printf("Error converting %s to %s: %v", inputFilename, outputFilename, err)
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
}
