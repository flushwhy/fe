package cmd

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestUpdateProcessErrorIncludesExitCodeAndRecoveryHint(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a POSIX test process")
	}
	err := exec.Command("sh", "-c", "exit 9").Run()
	if err == nil {
		t.Fatal("expected package manager process to fail")
	}
	got := updateProcessError("apt-get", err)
	if !strings.Contains(got.Message, "exit code 9") {
		t.Fatalf("updateProcessError() message = %q, want exit code", got.Message)
	}
	if !strings.Contains(got.Hint, "package manager output") {
		t.Fatalf("updateProcessError() hint = %q, want recovery guidance", got.Hint)
	}
}

func TestUpdateProcessErrorExplainsStartFailure(t *testing.T) {
	got := updateProcessError("brew", errors.New("executable not found"))
	if !strings.Contains(got.Message, "could not start brew") {
		t.Fatalf("updateProcessError() message = %q, want start-failure context", got.Message)
	}
}
