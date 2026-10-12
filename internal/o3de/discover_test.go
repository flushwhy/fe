package o3de

import (
	"os"
	"path/filepath"
	"testing"
)

func writeMarker(t *testing.T, root, marker, content string) {
	t.Helper()
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, marker), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestDiscoverFindsLocalUnregisteredEngine(t *testing.T) {
	root := t.TempDir()
	engine := filepath.Join(root, "engines", "o3de-main")
	writeMarker(t, engine, "engine.json", "{\"engine_name\":\"O3DE Main\",\"version\":\"24.09\"}")
	t.Setenv("O3DE_MANIFEST_PATH", filepath.Join(root, "missing-manifest.json"))

	got, err := Discover("engines", []string{root}, 4, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "O3DE Main" || got[0].Source != "local scan" {
		t.Fatalf("unexpected discovery results: %#v", got)
	}
}

func TestDiscoverDoesNotClassifyEveryExternalDirectoryAsGem(t *testing.T) {
	root := t.TempDir()
	gem := filepath.Join(root, "Gems", "SampleGem")
	writeMarker(t, gem, "gem.json", "{\"gem_name\":\"SampleGem\"}")
	other := filepath.Join(root, "external", "third-party")
	if err := os.MkdirAll(other, 0755); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(root, "o3de_manifest.json")
	if err := os.WriteFile(manifest, []byte("{\"external_subdirectories\":[\""+other+"\"]}"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("O3DE_MANIFEST_PATH", manifest)

	got, err := Discover("gems", []string{root}, 4, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "SampleGem" {
		t.Fatalf("expected only the directory with gem.json, got %#v", got)
	}
}

func TestDiscoverHonorsDepth(t *testing.T) {
	root := t.TempDir()
	engine := filepath.Join(root, "one", "two", "three", "engine")
	writeMarker(t, engine, "engine.json", "{\"engine_name\":\"Too Deep\"}")
	t.Setenv("O3DE_MANIFEST_PATH", filepath.Join(root, "missing-manifest.json"))

	got, err := Discover("engines", []string{root}, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("expected depth limit to exclude engine, got %#v", got)
	}
}

func TestDiscoverRejectsUnknownKind(t *testing.T) {
	if _, err := Discover("unknown", nil, 2, false); err == nil {
		t.Fatal("expected unknown kind error")
	}
}
