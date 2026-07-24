package butler

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

var discardLogger = log.New(io.Discard, "", 0)

// helperExecutor runs the test binary itself as a no-op subprocess.
func helperExecutor(t *testing.T, wantArgs []string) CommandExecutor {
	t.Helper()
	return func(name string, arg ...string) *exec.Cmd {
		got := append([]string{name}, arg...)
		if !reflect.DeepEqual(got, wantArgs) {
			t.Errorf("unexpected command args:\ngot  %v\nwant %v", got, wantArgs)
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

func TestPush(t *testing.T) {
	t.Run("fails when directory does not exist", func(t *testing.T) {
		err := Push("user", "game", "non-existent-dir", "", nil, discardLogger)
		if err == nil {
			t.Fatal("expected error for non-existent directory")
		}
	})

	t.Run("convention layout: windows-x64", func(t *testing.T) {
		tmp := t.TempDir()
		os.MkdirAll(filepath.Join(tmp, "windows-x64"), 0755)

		want := []string{"butler", "push", filepath.Join(tmp, "windows-x64"), "u/g:windows-x64"}
		err := Push("u", "g", tmp, "", helperExecutor(t, want), discardLogger)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("convention layout: linux-arm64 with userversion", func(t *testing.T) {
		tmp := t.TempDir()
		os.MkdirAll(filepath.Join(tmp, "linux-arm64"), 0755)

		want := []string{"butler", "push", filepath.Join(tmp, "linux-arm64"), "u/g:linux-arm64", "--userversion", "1.0.0"}
		err := Push("u", "g", tmp, "1.0.0", helperExecutor(t, want), discardLogger)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("legacy two-level layout: windows/x64", func(t *testing.T) {
		tmp := t.TempDir()
		os.MkdirAll(filepath.Join(tmp, "windows", "x64"), 0755)

		want := []string{"butler", "push", filepath.Join(tmp, "windows", "x64"), "u/g:windows-x64"}
		err := Push("u", "g", tmp, "", helperExecutor(t, want), discardLogger)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("skips non-platform folders", func(t *testing.T) {
		tmp := t.TempDir()
		os.MkdirAll(filepath.Join(tmp, "documentation"), 0755)

		failExecutor := func(name string, arg ...string) *exec.Cmd {
			t.Error("executor should not have been called")
			return nil
		}

		err := Push("u", "g", tmp, "", failExecutor, discardLogger)
		if err == nil {
			t.Fatal("expected error when no valid platform folders found")
		}
	})

	t.Run("fails when no platform folders found", func(t *testing.T) {
		tmp := t.TempDir()
		err := Push("u", "g", tmp, "", nil, discardLogger)
		if err == nil {
			t.Fatal("expected error for empty directory")
		}
	})
}
