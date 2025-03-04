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
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// pngjoinerCmd represents the pngjoiner command
var pngjoinerCmd = &cobra.Command{
	Use:   "pngjoiner",
	Short: "This takes multiple PNGs and combines them into a single PNG.",
	Long:  `This takes multiple PNGs and combines them into a single PNG.`,
	Run: func(cmd *cobra.Command, args []string) {
		inputFile, _ := cmd.Flags().GetString("input")
		outputFile, _ := cmd.Flags().GetString("output")
		rows, _ := cmd.Flags().GetInt("rows")
		cols, _ := cmd.Flags().GetInt("cols")

		if inputFile == "" || outputFile == "you_forgot_to_specify_an_output_file.png" || rows == 0 || cols == 0 {
			cmd.Help()
			return
		}

		err := PngJoiner(inputFile, outputFile, rows, cols)
		if err != nil {
			fmt.Println("Error:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(pngjoinerCmd)

	pngjoinerCmd.Flags().String("output", "you_forgot_to_specify_an_output_file.png", "output file")
	pngjoinerCmd.Flags().String("input", "", "input file(s) or directory")
	pngjoinerCmd.Flags().Int("rows", 0, "rows")
	pngjoinerCmd.Flags().Int("cols", 0, "cols")
}

func PngJoiner(inputFile string, outputFile string, rows int, cols int) error {

	imageFiles := []string{}
	if strings.HasSuffix(strings.ToLower(inputFile), ".png") {
		imageFiles = append(imageFiles, inputFile)
	} else {
		files, err := filepath.Glob(filepath.Join(inputFile, "*.png"))
		if err != nil {
			return err
		}
		imageFiles = append(imageFiles, files...)
	}

	img, err := LoadImageFromFiles(imageFiles)
	if err != nil {
		return err
	}

	err = SaveImage(outputFile, img, rows, cols)
	if err != nil {
		return err
	}

	return nil
}

func LoadImageFromFiles(files []string) (image.Image, error) {
	var images []image.Image
	for _, file := range files {
		img, err := readImage(file)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	bounds := images[0].Bounds()
	for _, img := range images {
		if !img.Bounds().Eq(bounds) {
			return nil, fmt.Errorf("different bounds")
		}
	}
	return images[0], nil
}

func readImage(file string) (image.Image, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func SaveImage(outputFile string, img image.Image, rows int, cols int) error {
	bounds := img.Bounds()
	dx := bounds.Dx() / cols
	dy := bounds.Dy() / rows
	newImg := image.NewRGBA(image.Rect(0, 0, dx*cols, dy*rows))
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			r := image.NewRGBA(image.Rect(0, 0, dx, dy))
			draw.Draw(r, r.Bounds(), img, image.Point{dx * x, dy * y}, draw.Src)
			draw.Draw(newImg, newImg.Bounds(), r, image.Point{dx * x, dy * y}, draw.Over)
		}
	}
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, newImg)
}
