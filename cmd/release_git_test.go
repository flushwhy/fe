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
		if !strings.Contains(strings.Join(commandErr.Details, "\n"), file) {
			t.Errorf("release error details %#v do not include %q", commandErr.Details, file)
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
