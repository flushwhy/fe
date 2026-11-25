//go:build windows

package o3de

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIInvocationPassesShellMetacharactersAsLiteralBatchArguments(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "record-args.bat")
	output := filepath.Join(dir, "args.txt")
	body := "@echo off\r\nsetlocal DisableDelayedExpansion\r\n> \"%~dp0args.txt\" echo %1\r\n>> \"%~dp0args.txt\" echo %2\r\n"
	if err := os.WriteFile(script, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}

	malicious := "Game & echo INJECTED > pwned.txt & rem"
	cmd, err := NewCLICommand(context.Background(), script, []string{"register", malicious})
	if err != nil {
		t.Fatal(err)
	}
	cmd.Dir = dir
	if combined, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("O3DE batch invocation failed: %v\n%s", err, combined)
	}

	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read captured arguments: %v", err)
	}
	if !strings.Contains(string(got), `"Game & echo INJECTED > pwned.txt & rem"`) {
		t.Fatalf("batch script did not receive the argument literally: %q", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "pwned.txt")); !os.IsNotExist(err) {
		t.Fatalf("shell metacharacters executed as a command; pwned.txt stat error = %v", err)
	}
}
