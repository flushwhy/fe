package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func initReleaseGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, output)
		}
		return strings.TrimSpace(string(runOutput(cmd)))
	}
	_ = run
	if output, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}
	for _, args := range [][]string{
		{"-C", dir, "config", "user.email", "fe-test@example.invalid"},
		{"-C", dir, "config", "user.name", "Fe Test"},
	} {
		if output, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "tracked.txt"), []byte("clean\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"-C", dir, "add", "tracked.txt"}, {"-C", dir, "commit", "-qm", "initial"}} {
		if output, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	return dir
}

func runGitAt(t *testing.T, dir string, args ...string) {
	t.Helper()
	full := append([]string{"-C", dir}, args...)
	if output, err := exec.Command("git", full...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
}

func runOutput(cmd *exec.Cmd) []byte { return nil }

func TestValidateReleaseGitAllowsNonRepository(t *testing.T) {
	state := releaseGitState{}
	if err := validateReleaseGit(state, false, false); err != nil {
		t.Fatalf("non-repository should remain supported: %v", err)
	}
}

func TestInspectReleaseGitDetectsDirtyTree(t *testing.T) {
	dir := initReleaseGitRepo(t)
	old, err := os.Getwd()
	if err != nil { t.Fatal(err) }
	if err := os.Chdir(dir); err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = os.Chdir(old) })

	if err := os.WriteFile(filepath.Join(dir, "tracked.txt"), []byte("changed\n"), 0o600); err != nil { t.Fatal(err) }
	state, err := inspectReleaseGit()
	if err != nil { t.Fatal(err) }
	if !state.IsRepo || !state.Dirty || state.Commit == "" { t.Fatalf("unexpected Git state: %+v", state) }
	if err := validateReleaseGit(state, false, true); err == nil || !strings.Contains(err.Error(), "uncommitted") {
		t.Fatalf("expected dirty-tree blocker, got %v", err)
	}
	if err := validateReleaseGit(state, true, true); err != nil { t.Fatalf("explicit override should allow dirty tree: %v", err) }
}

func TestInspectReleaseGitUntrackedFilesAreDirty(t *testing.T) {
	dir := initReleaseGitRepo(t)
	old, err := os.Getwd()
	if err != nil { t.Fatal(err) }
	if err := os.Chdir(dir); err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = os.Chdir(old) })
	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new\n"), 0o600); err != nil { t.Fatal(err) }
	state, err := inspectReleaseGit()
	if err != nil { t.Fatal(err) }
	if !state.Dirty { t.Fatal("untracked file should make the working tree dirty") }
}
