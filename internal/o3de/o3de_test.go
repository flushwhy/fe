package o3de

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegisteredReadsO3DEManifestArray(t *testing.T) {
	manifest := filepath.Join(t.TempDir(), "o3de_manifest.json")
	t.Setenv("O3DE_MANIFEST_PATH", manifest)
	content := `{"engines":["/opt/o3de","/work/another-engine"]}`
	if err := os.WriteFile(manifest, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	entries, err := Registered("engines")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Name != "o3de" || entries[0].Path != "/opt/o3de" {
		t.Fatalf("unexpected entries: %#v", entries)
	}
}

func TestRegisteredRejectsUnknownKind(t *testing.T) {
	if _, err := Registered("unknown"); err == nil {
		t.Fatal("expected unknown kind error")
	}
}

func TestFindEngineRequiresManifest(t *testing.T) {
	dir := t.TempDir()
	if _, err := FindEngine(dir); err == nil {
		t.Fatal("expected missing engine.json error")
	}
}

func TestFindEngineReadsEngineManifest(t *testing.T) {
	dir := t.TempDir()
	content := `{"engine_name":"TestEngine","version":"24.09"}`
	if err := os.WriteFile(filepath.Join(dir, "engine.json"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	entry, err := FindEngine(dir)
	if err != nil {
		t.Fatal(err)
	}
	if entry.Name != "TestEngine" || entry.Version != "24.09" {
		t.Fatalf("unexpected engine: %#v", entry)
	}
}
