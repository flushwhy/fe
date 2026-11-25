package cmd

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDiscoverCheckTasksDetectsGoProject(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/test\n\ngo 1.25\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\nfunc main() {}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	tasks, _, err := discoverCheckTasks(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) == 0 || tasks[0].name != "go vet ./..." {
		t.Fatalf("expected Go vet task, got %#v", tasks)
	}
}

func TestDiscoverCheckTasksSkipsGeneratedDirectories(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "node_modules"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "node_modules", "ignored.js"), []byte(""), 0600); err != nil {
		t.Fatal(err)
	}
	tasks, skipped, err := discoverCheckTasks(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 0 || len(skipped) != 0 {
		t.Fatalf("generated files should not trigger checks: tasks=%#v skipped=%#v", tasks, skipped)
	}
}

func TestFindCompileCommands(t *testing.T) {
	root := t.TempDir()
	build := filepath.Join(root, "build")
	if err := os.Mkdir(build, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(build, "compile_commands.json"), []byte("[]"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := findCompileCommands(root); got != build {
		t.Fatalf("findCompileCommands() = %q, want %q", got, build)
	}
}

func TestAnalyzerFailureDetailReportsExitCode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a POSIX test process")
	}
	err := exec.Command("sh", "-c", "exit 12").Run()
	if err == nil {
		t.Fatal("expected analyzer process to fail")
	}
	got := analyzerFailureDetail("go vet ./...", err)
	if !strings.Contains(got, "go vet ./...") || !strings.Contains(got, "exited with code 12") {
		t.Fatalf("analyzerFailureDetail() = %q, want task and exit code", got)
	}
}

func TestAnalyzerFailureDetailExplainsStartFailure(t *testing.T) {
	got := analyzerFailureDetail("ruff check .", errors.New("executable not found"))
	if !strings.Contains(got, "ruff check .") || !strings.Contains(got, "could not start") {
		t.Fatalf("analyzerFailureDetail() = %q, want task and start-failure context", got)
	}
}
