//go:build !windows

package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultRunnerReportsProcessExitCode(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "fe-test-failing-command")
	script := "#!/bin/sh\nexit 23\n"
	if err := os.WriteFile(binary, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	err := DefaultRunner("fe-test-failing-command")
	if err == nil {
		t.Fatal("expected command failure")
	}
	if !strings.Contains(err.Error(), "exit code 23") {
		t.Fatalf("DefaultRunner() error = %q, want exit code 23", err)
	}
}
