package preflight

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasAnyDetectsProjectMarkers(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "pyproject.toml"), []byte("[project]"), 0600); err != nil {
		t.Fatal(err)
	}
	if !hasAny(root, ".py", "pyproject.toml") {
		t.Fatal("expected Python project marker to be detected")
	}
}

func TestHasAnyIgnoresGeneratedDirectories(t *testing.T) {
	root := t.TempDir()
	generated := filepath.Join(root, "vendor")
	if err := os.Mkdir(generated, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(generated, "ignored.lua"), []byte(""), 0600); err != nil {
		t.Fatal(err)
	}
	if hasAny(root, ".lua") {
		t.Fatal("generated vendor files should not trigger language detection")
	}
}

func TestDetectPythonAndLuaOnlyWhenPresent(t *testing.T) {
	root := t.TempDir()
	if got := detectPython(root); len(got) != 0 {
		t.Fatalf("unexpected Python result in empty directory: %#v", got)
	}
	if got := detectLua(root); len(got) != 0 {
		t.Fatalf("unexpected Lua result in empty directory: %#v", got)
	}
}
