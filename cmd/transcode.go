/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
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
