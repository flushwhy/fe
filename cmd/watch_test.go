package cmd

import "testing"

func TestWatchDirIgnored(t *testing.T) {
	for _, name := range []string{".git", "build", "node_modules"} {
		if !watchDirIgnored(name, nil) {
			t.Errorf("expected default directory %q to be ignored", name)
		}
	}
	if !watchDirIgnored("Generated", []string{"generated"}) {
		t.Fatal("custom ignored directory names should be case-insensitive")
	}
	if watchDirIgnored("src", []string{"generated"}) {
		t.Fatal("unexpectedly ignored src")
	}
}

func TestWatchIgnoreDirPositionalValues(t *testing.T) {
	cmd := watchCmd
	if err := cmd.Args(cmd, []string{"cache"}); err == nil {
		t.Fatal("expected positional directory names without --ignore-dir to fail")
	}
	if err := cmd.Flags().Set("ignore-dir", "generated"); err != nil { t.Fatal(err) }
	if err := cmd.Args(cmd, []string{"generated", "cache"}); err != nil {
		t.Fatalf("expected positional directory names with --ignore-dir to be accepted: %v", err)
	}
}
