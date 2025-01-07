package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPngJoiner(t *testing.T) {
	// Test case 1: Single PNG file input
	inputFile := "test.png"
	outputFile := "output.png"
	rows := 2
	cols := 2

	// Create a temporary input file
	inputFilePath := createTempFile(t, inputFile)
	defer os.Remove(inputFilePath)

	// Call the PngJoiner function
	err := PngJoiner(inputFilePath, outputFile, rows, cols)
	if err != nil {
		t.Errorf("PngJoiner returned an error for single PNG input: %v", err)
	}

	// Check if the output file exists
	_, err = os.Stat(outputFile)
	if os.IsNotExist(err) {
		t.Errorf("Output file does not exist for single PNG input")
	}

	// Test case 2: Multiple PNG files in a directory input
	inputDir := "test_dir"
	outputFile = "output.png"
	rows = 2
	cols = 2

	// Create a temporary input directory and files
	inputDirPath := createTempDir(t, inputDir)
	defer os.RemoveAll(inputDirPath)
	createTempFile(t, filepath.Join(inputDirPath, "test1.png"))
	createTempFile(t, filepath.Join(inputDirPath, "test2.png"))

	// Call the PngJoiner function
	err = PngJoiner(inputDirPath, outputFile, rows, cols)
	if err != nil {
		t.Errorf("PngJoiner returned an error for multiple PNG files in a directory input: %v", err)
	}

	// Check if the output file exists
	_, err = os.Stat(outputFile)
	if os.IsNotExist(err) {
		t.Errorf("Output file does not exist for multiple PNG files in a directory input")
	}

	// Test case 3: Invalid input file
	inputFile = "invalid.png"
	outputFile = "output.png"
	rows = 2
	cols = 2

	// Call the PngJoiner function
	err = PngJoiner(inputFile, outputFile, rows, cols)
	if err == nil {
		t.Errorf("PngJoiner did not return an error for invalid input file")
	}
}

// Helper function to create a temporary file
func createTempFile(t *testing.T, fileName string) string {
	file, err := os.CreateTemp("", fileName)
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	file.Close()
	return file.Name()
}

// Helper function to create a temporary directory
func createTempDir(t *testing.T, dirName string) string {
	dir, err := os.MkdirTemp("", dirName)
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	return dir
}
