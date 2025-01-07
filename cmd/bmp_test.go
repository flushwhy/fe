<<<<<<< HEAD
package cmd_test

import (
	"testing"

	"codeberg.org/flush/fe/cmd"
)

func TestButlerPusher(t *testing.T) {
	// Test case when the directory can be read successfully
	directory := "testdata/cmd/butler_pusher"
	username := "flush"
	game := "xo"
	userversion := ""
	t.Run("Directory Read Success", func(t *testing.T) {
		testButlerPusher(t, directory, username, game, userversion)
	})

	// Test case when the subdirectory can be read successfully
	directory = "testdata/cmd/butler_pusher"
	username = "ff"
	game = "xo"
	userversion = ""
	t.Run("Subdirectory Read Success", func(t *testing.T) {
		testButlerPusher(t, directory, username, game, userversion)
	})

	// Test case when user version is empty
	directory = "testdata/cmd/butler_pusher"
	username = "ff"
	game = "xo"
	userversion = ""
	t.Run("UserVersion Empty", func(t *testing.T) {
		testButlerPusher(t, directory, username, game, userversion)
	})

	// Test case when user version is not empty
	directory = "testdata/cmd/butler_pusher"
	username = "ff"
	game = "xo"
	userversion = "1.0.0"
	t.Run("UserVersion Not Empty", func(t *testing.T) {
		testButlerPusher(t, directory, username, game, userversion)
	})
}

func testButlerPusher(t *testing.T, directory, username, game, userversion string) {
	cmd.Butler_pusher(username, game, directory, userversion)
=======
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
>>>>>>> flush_dev
}
