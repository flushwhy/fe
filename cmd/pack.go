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
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"sync" // Import the sync package
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// packCmd represents the pack command
var packCmd = &cobra.Command{
	Use:   "pack",
	Short: "Packs lots of PNGs into a single PNG.",
	Long:  `This packs all the PNGs in a given folder do a single PNG.`,
	Run: func(cmd *cobra.Command, args []string) {
		inputdir := viper.GetString("pack.input")
		output := viper.GetString("pack.output")

		if inputdir == "" {
			log.Fatal("Error: --input must be set via flag or config file.")
		}

		outputdir := filepath.Dir(output)
		if err := os.MkdirAll(outputdir, os.ModePerm); err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}

		var mu sync.Mutex
		var attempts int = 0
		const maxAttempts = 5

		for attempts < maxAttempts {
			mu.Lock()
			defer mu.Unlock()

			if err := os.RemoveAll(output); err != nil {
				log.Printf("Attempt %d: Failed to remove output file: %v", attempts, err)
				time.Sleep(1 * time.Second)
				attempts++
				continue
			}

			break
		}

		if attempts == maxAttempts {
			log.Fatalf("Failed to remove output file after %d attempts", maxAttempts)
		}

		tempFile, err := os.CreateTemp(outputdir, "temp-write-test-*.tmp")
		if err != nil {
			log.Fatalf("No write permissions for output directory %s: %v", outputdir, err)
		}
		defer tempFile.Close()
		os.Remove(tempFile.Name())

		files, err := os.ReadDir(inputdir)
		if err != nil {
			log.Fatal("Failed to read input directory: ", err)
		}

		var images []image.Image
		totalWidth := 0
		maxHeight := 0

		for _, file := range files {
			if file.IsDir() || filepath.Ext(file.Name()) != ".png" {
				continue
			}

			filePath := filepath.Join(inputdir, file.Name())
			f, err := os.Open(filePath)
			if err != nil {
				log.Printf("Could not open file %s: %v", filePath, err)
				continue
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				log.Printf("Could not decode image %s: %v", filePath, err)
				continue
			}

			images = append(images, img)
			totalWidth += img.Bounds().Dx()
			if img.Bounds().Dy() > maxHeight {
				maxHeight = img.Bounds().Dy()
			}
		}

		if len(images) == 0 {
			log.Fatal("No PNG images found in the input directory.")
		}

		canvas := image.NewRGBA(image.Rect(0, 0, totalWidth, maxHeight))

		currentX := 0
		for _, img := range images {
			rect := image.Rect(currentX, 0, currentX+img.Bounds().Dx(), img.Bounds().Dy())
			draw.Draw(canvas, rect, img, image.Point{}, draw.Src)
			currentX += img.Bounds().Dx()
		}

		outFile, err := os.Create(output)
		if err != nil {
			log.Fatalf("Failed to create output file: %v", err)
		}
		defer outFile.Close()

		err = png.Encode(outFile, canvas)
		if err != nil {
			log.Fatalf("Failed to encode final image: %v", err)
		}

		fmt.Printf("✅ Successfully packed %d images into %s\n", len(images), output)

	},
}

func init() {
	rootCmd.AddCommand(packCmd)

	packCmd.Flags().String("input", "", "Place to PNGs")
	packCmd.Flags().String("output", "out.png", "Output to single .png")

	viper.BindPFlag("pack.input", packCmd.Flags().Lookup("input"))
	viper.BindPFlag("pack.output", packCmd.Flags().Lookup("output"))
}
