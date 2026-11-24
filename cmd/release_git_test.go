package cmd

import (
	"os"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flushwhy/fe/internal/ui"
)

func initReleaseGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
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

func TestValidateReleaseGitAllowsNonRepository(t *testing.T) {
	state := releaseGitState{}
	if err := validateReleaseGit(state, false, false); err != nil {
		t.Fatalf("non-repository should remain supported: %v", err)
	}
}

func TestInspectReleaseGitDetectsDirtyTree(t *testing.T) {
	t.Setenv("GITHUB_SHA", "")
	dir := initReleaseGitRepo(t)
	old, err := os.Getwd()
	if err != nil { t.Fatal(err) }
	if err := os.Chdir(dir); err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = os.Chdir(old) })

	if err := os.WriteFile(filepath.Join(dir, "tracked.txt"), []byte("changed\n"), 0o600); err != nil { t.Fatal(err) }
	state, err := inspectReleaseGit()
	if err != nil { t.Fatal(err) }
	if !state.IsRepo || !state.Dirty || state.Commit == "" { t.Fatalf("unexpected Git state: %+v", state) }
	if err := validateReleaseGit(state, false, true); err == nil || !strings.Contains(err.Error(), "tracked.txt") {
		t.Fatalf("expected dirty-tree blocker to list tracked.txt, got %v", err)
	}
	if err := validateReleaseGit(state, true, true); err != nil { t.Fatalf("explicit override should allow dirty tree: %v", err) }
}

func TestInspectReleaseGitUntrackedFilesAreDirty(t *testing.T) {
	t.Setenv("GITHUB_SHA", "")
	dir := initReleaseGitRepo(t)
	old, err := os.Getwd()
	if err != nil { t.Fatal(err) }
	if err := os.Chdir(dir); err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = os.Chdir(old) })
	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new\n"), 0o600); err != nil { t.Fatal(err) }
	state, err := inspectReleaseGit()
	if err != nil { t.Fatal(err) }
	if !state.Dirty { t.Fatal("untracked file should make the working tree dirty") }
	err = validateReleaseGit(state, false, true)
	if err == nil || !strings.Contains(err.Error(), "new.txt") {
		t.Fatalf("expected dirty-tree blocker to list new.txt, got %v", err)
	}
}

func TestValidateReleaseGitProvidesDirtyTreeDetailsAndRecoveryHint(t *testing.T) {
	state := releaseGitState{
		IsRepo:     true,
		Dirty:      true,
		DirtyFiles: []string{"M tracked.txt", "?? new file.txt"},
	}
	err := validateReleaseGit(state, false, true)
	if err == nil {
		t.Fatal("expected dirty working tree to block release")
	}
	var commandErr *ui.CommandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("expected ui.CommandError, got %T: %v", err, err)
	}
	for _, file := range []string{"M tracked.txt", "?? new file.txt"} {
		if !strings.Contains(commandErr.Message, file) {
			t.Errorf("release error %q does not list %q", commandErr.Message, file)
		}
	}
	if !strings.Contains(commandErr.Hint, "--allow-dirty") {
		t.Fatalf("recovery hint should explain the explicit override: %q", commandErr.Hint)
	}
}

func TestValidateReleaseGitExplainsUnpushedCommits(t *testing.T) {
	state := releaseGitState{IsRepo: true, UnpushedKnown: true, Unpushed: 3}
	err := validateReleaseGit(state, true, false)
	if err == nil {
		t.Fatal("expected unpushed commits to block release")
	}
	var commandErr *ui.CommandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("expected ui.CommandError, got %T: %v", err, err)
	}
	if !strings.Contains(commandErr.Message, "3 commit(s)") || !strings.Contains(commandErr.Hint, "push") {
		t.Fatalf("unexpected unpushed-commit guidance: %#v", commandErr)
	}
}

func TestValidateReleaseGitExplainsUnknownRemoteState(t *testing.T) {
	state := releaseGitState{IsRepo: true}
	err := validateReleaseGit(state, true, false)
	if err == nil {
		t.Fatal("expected unknown remote state to block release")
	}
	var commandErr *ui.CommandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("expected ui.CommandError, got %T: %v", err, err)
	}
	if !strings.Contains(commandErr.Message, "cannot verify") || !strings.Contains(commandErr.Hint, "upstream") {
		t.Fatalf("unexpected remote-state guidance: %#v", commandErr)
	}
}
