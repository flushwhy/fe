package cmd

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func mockExecutor(t *testing.T, expectedArgs []string) commandExecutor {
	return func(name string, arg ...string) *exec.Cmd {

		allArgs := append([]string{name}, arg...)

		if !reflect.DeepEqual(allArgs, expectedArgs) {
			t.Errorf("unexpected command args:\ngot:  %v\nwant: %v", allArgs, expectedArgs)
		}

		cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess")
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	os.Exit(0)
}

func TestButlerPusher(t *testing.T) {

	discardLogger := log.New(io.Discard, "", 0)

	t.Run("fails when directory does not exist", func(t *testing.T) {

		err := Butler_pusher("user", "game", "non-existent-dir", "", nil, discardLogger)

		if err == nil {
			t.Error("expected an error for a non-existent directory, but got nil")
		}
	})

	t.Run("successfully pushes a valid directory structure", func(t *testing.T) {

		tempDir := t.TempDir()

		platformDir := filepath.Join(tempDir, "windows")
		archDir := filepath.Join(platformDir, "x64")
		os.MkdirAll(archDir, 0755)

		expectedCmd := []string{
			"butler",
			"push",
			archDir,
			"test_user/test_game:windowsx64",
		}

		err := Butler_pusher("test_user", "test_game", tempDir, "", mockExecutor(t, expectedCmd), discardLogger)

		if err != nil {
			t.Errorf("expected no error, but got: %v", err)
		}
	})

	t.Run("successfully pushes with a userversion", func(t *testing.T) {
		tempDir := t.TempDir()
		platformDir := filepath.Join(tempDir, "linux")
		archDir := filepath.Join(platformDir, "arm64")
		os.MkdirAll(archDir, 0755)

		expectedCmd := []string{
			"butler",
			"push",
			archDir,
			"test_user/test_game:linuxarm64",
			"--userversion",
			"1.2.3",
		}

		err := Butler_pusher("test_user", "test_game", tempDir, "1.2.3", mockExecutor(t, expectedCmd), discardLogger)

		if err != nil {
			t.Errorf("expected no error, but got: %v", err)
		}
	})

	t.Run("skips invalid platform folders", func(t *testing.T) {
		tempDir := t.TempDir()

		os.MkdirAll(filepath.Join(tempDir, "documentation"), 0755)

		failingExecutor := func(name string, arg ...string) *exec.Cmd {
			t.Error("executor was called unexpectedly")
			return nil
		}

		err := Butler_pusher("user", "game", tempDir, "", failingExecutor, discardLogger)

		if err != nil {
			t.Errorf("expected no error when skipping folders, but got: %v", err)
		}
	})
}
