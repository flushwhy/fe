/*
Copyright 2024 Ryan Flush <roflush@pm.me>
*/
package cmd

import (
	"image"
	"image/png"
	"os"
	"testing"
)

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
