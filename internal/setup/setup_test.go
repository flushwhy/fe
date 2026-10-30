package setup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectBlenderAddonDoesNotRequireBlenderForPackaging(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "blender_manifest.toml"), []byte("schema_version = \"1.0.0\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	tools, err := Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 1 || tools[0].Name != "Blender" || tools[0].Required {
		t.Fatalf("expected optional Blender requirement, got %#v", tools)
	}
}

func TestInspectGoProject(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example\n"), 0600); err != nil {
		t.Fatal(err)
	}
	tools, err := Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 1 || tools[0].Name != "Go" || !tools[0].Required {
		t.Fatalf("expected required Go tool, got %#v", tools)
	}
}
