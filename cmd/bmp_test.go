package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestButlerPusher(t *testing.T) {
	// Test case: directory does not exist
	err := os.RemoveAll("test_dir")
	if err != nil {
		t.Fatalf("Could not remove test directory: %s", err)
	}
	err = Butler_pusher("test_user", "test_game", "test_dir", "")
	if err == nil {
		t.Errorf("Expected error for non-existent directory, but got nil")
	}

	// Test case: valid directory with no subdirectories
	err = os.Mkdir("test_dir", os.ModePerm)
	if err != nil {
		t.Fatalf("Could not create test directory: %s", err)
	}
	err = os.WriteFile(filepath.Join("test_dir", "test_file"), []byte("test"), os.ModePerm)
	if err != nil {
		t.Fatalf("Could not create test file: %s", err)
	}
	err = Butler_pusher("test_user", "test_game", "test_dir", "")
	if err == nil {
		t.Errorf("Expected error for invalid subdirectory, but got nil")
	}

	// Test case: valid directory with valid subdirectories
	err = os.Mkdir(filepath.Join("test_dir", "linux"), os.ModePerm)
	if err != nil {
		t.Fatalf("Could not create test subdirectory: %s", err)
	}
	err = os.WriteFile(filepath.Join("test_dir", "linux", "x64"), []byte("test"), os.ModePerm)
	if err != nil {
		t.Fatalf("Could not create test subfile: %s", err)
	}
	err = Butler_pusher("test_user", "test_game", "test_dir", "")
	if err != nil {
		t.Errorf("Expected nil error, but got: %s", err)
	}

	// Clean up test directory
	err = os.RemoveAll("test_dir")
	if err != nil {
		t.Fatalf("Could not remove test directory: %s", err)
	}
}
