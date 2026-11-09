package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateO3DEManifest(t *testing.T) {
	t.Run("missing directory", func(t *testing.T) {
		err := validateO3DEManifest(filepath.Join(t.TempDir(), "missing"), "project.json", "project")
		if err == nil || !strings.Contains(err.Error(), "does not exist") {
			t.Fatalf("expected missing-directory error, got %v", err)
		}
	})

	t.Run("path is not a directory", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "project")
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		err := validateO3DEManifest(path, "project.json", "project")
		if err == nil || !strings.Contains(err.Error(), "not a directory") {
			t.Fatalf("expected not-a-directory error, got %v", err)
		}
	})

	t.Run("missing manifest", func(t *testing.T) {
		err := validateO3DEManifest(t.TempDir(), "project.json", "project")
		if err == nil || !strings.Contains(err.Error(), "missing project.json") {
			t.Fatalf("expected missing-manifest error, got %v", err)
		}
	})

	t.Run("manifest is a directory", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, "project.json"), 0o700); err != nil {
			t.Fatal(err)
		}
		err := validateO3DEManifest(root, "project.json", "project")
		if err == nil || !strings.Contains(err.Error(), "not a regular file") {
			t.Fatalf("expected invalid-manifest-type error, got %v", err)
		}
	})

	t.Run("valid manifest", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "project.json"), []byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := validateO3DEManifest(root, "project.json", "project"); err != nil {
			t.Fatalf("expected valid manifest, got %v", err)
		}
	})
}

func TestLooksLikePath(t *testing.T) {
	for _, value := range []string{".", "./project", "../project", "/projects/game", `C:\Games\Project`, "Games/Project"} {
		if !looksLikePath(value) {
			t.Errorf("expected %q to be recognized as a path", value)
		}
	}
	if looksLikePath("MyGame") {
		t.Fatal("registered project name should not be treated as a path")
	}
}
