/*
Copyright © 2024 Ryan Flush <roflush@pm.me>
*/
package cmd

import (
	"fmt"
	"image"

	"image/png"

	"os"

	"github.com/nfnt/resize"
	"github.com/spf13/cobra"
)

// resizetextureCmd represents the resizetexture command
var resizetextureCmd = &cobra.Command{
	Use:   "resize",
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

func init() {
	rootCmd.AddCommand(resizetextureCmd)

	// This is flags
	resizetextureCmd.Flags().String("input", "", "input file")
	resizetextureCmd.Flags().String("output", "", "output file")
	resizetextureCmd.Flags().Int("range", 4, "The range you want to resize from 16k to 2k")
}

func ResizeTexture(img image.Image, outputFile string, rangeValue int) error {
	res := map[int]int{2: 2000, 4: 4000, 6: 6000, 8: 8000, 12: 12000, 16: 16000}

	for key, val := range res {
		if res[key] <= rangeValue {

			newImg := resize.Resize(uint(res[val]), uint(res[val]), img, resize.Bilinear)
			f, err := os.Create(outputFile)
			if err != nil {
				return err
			}
			defer f.Close()
			return png.Encode(f, newImg)
		}
	}

	return nil
}
