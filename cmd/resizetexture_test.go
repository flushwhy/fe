/*
Copyright 2024 Ryan Flush <roflush@pm.me>
*/
package cmd

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

// resizetextureCmd represents the resizetexture command
var resizetextureCmd = &cobra.Command{
	Use:   "resizetexture",
	Short: "Take your textures from 16k to 2k.",
	Long:  `This command is used to resize textures from everything from 16k to 2k. You can only go down not up.`,
	Run: func(cmd *cobra.Command, args []string) {

		inputFile, _ := cmd.Flags().GetString("input")
		outputFile, _ := cmd.Flags().GetString("output")
		rangeValue, _ := cmd.Flags().GetInt("range")

		if inputFile == "" || outputFile == "" || rangeValue == 0 {
			cmd.Help()
			return
		}

		if rangeValue > 16 || rangeValue < 2 || rangeValue%2 != 0 {
			panic("Range value must be between 2 and 16 and must be even")
		}

		imgFile, err := os.Open(inputFile)
		if err != nil {
			panic(err)
		}
		defer imgFile.Close()

		img, _, err := image.Decode(imgFile)
		if err != nil {
			panic(err)
		}

		// Checking for square images
		if img.Bounds().Dx() != img.Bounds().Dy() {
			panic("Image is not square, please use a square image")

		}

		err = ResizeTexture(img, outputFile, rangeValue)
		if err != nil {
			fmt.Println("Error:", err)
		}
	},
}

func TestResizeTexture(t *testing.T) {
	imgFile, err := os.Open("testdata/8k.png")
	if err != nil {
		t.Fatal(err)
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		t.Fatal(err)
	}

	outputFile := "testdata/output.png"
	err = ResizeTexture(img, outputFile, 4)
	if err != nil {
		t.Fatal(err)
	}

	// Check that the output file exists
	_, err = os.Stat(outputFile)
	if err != nil {
		t.Fatal(err)
	}

	// Check that the output file is a valid PNG
	f, err := os.Open(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	_, err = png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}

	// Check that the output file is the correct size
	imgOutput, _, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if imgOutput.Bounds().Dx() != 4000 || imgOutput.Bounds().Dy() != 4000 {
		t.Fatal("Output image is not the correct size")
	}
}

func init() {
	rootCmd.AddCommand(resizetextureCmd)

	// This is flags
	resizetextureCmd.Flags().String("input", "", "input file")
	resizetextureCmd.Flags().String("output", "", "output file")
	resizetextureCmd.Flags().Int("range", 4, "The range you want to resize from 16k to 2k")
}
