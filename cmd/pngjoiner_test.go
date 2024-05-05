package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPngJoiner(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test_pngjoiner")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir) // clean up

	// Copy test PNG files into the temporary directory
	files := []string{"test1.png", "test2.png", "test3.png"}
	for _, file := range files {
		err := copyFile(file, filepath.Join(tmpDir, file))
		if err != nil {
			t.Fatalf("Failed to copy test file: %v", err)
		}
	}

	// Call PngJoiner with the temporary directory as the input file
	outputFile := "output.png"
	err = PngJoiner(tmpDir, outputFile, 2, 2)
	if err != nil {
		t.Fatalf("PngJoiner failed: %v", err)
	}

	// Add your assertions here to check the output file or any other conditions
}

func copyFile(file, s string) error {
	panic("unimplemented")
}
