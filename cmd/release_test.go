package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateReleaseBuildDir(t *testing.T) {
	t.Run("missing directory", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "builds")
		err := validateReleaseBuildDir(dir)
		if err == nil {
			t.Fatal("expected missing build directory to fail")
		}
		if !strings.Contains(err.Error(), "configure build.targets") {
			t.Fatalf("expected actionable configuration guidance, got: %v", err)
		}
	})

	t.Run("directory without platform folders", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "README.txt"), []byte("not a build"), 0o600); err != nil {
			t.Fatal(err)
		}
		err := validateReleaseBuildDir(dir)
		if err == nil || !strings.Contains(err.Error(), "contains no platform folders") {
			t.Fatalf("expected missing platform folders error, got: %v", err)
		}
	})

	t.Run("prebuilt platform folder", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "builds")
		if err := os.MkdirAll(filepath.Join(dir, "linux-x64"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := validateReleaseBuildDir(dir); err != nil {
			t.Fatalf("expected prebuilt output to be accepted, got: %v", err)
		}
	})
}
