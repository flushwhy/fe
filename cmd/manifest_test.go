package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteArtifactManifest(t *testing.T) {
	root := t.TempDir()
	linuxDir := filepath.Join(root, "linux-x64")
	if err := os.MkdirAll(linuxDir, 0o755); err != nil {
		t.Fatal(err)
	}
	payload := []byte("game binary")
	if err := os.WriteFile(filepath.Join(linuxDir, "game"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(root, "artifact-manifest.json")
	if err := writeArtifactManifest(root, output, "1.2.3", "abc123"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	var got artifactManifest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Version != "1.2.3" || got.Commit != "abc123" {
		t.Fatalf("unexpected metadata: version=%q commit=%q", got.Version, got.Commit)
	}
	if len(got.Artifacts) != 1 {
		t.Fatalf("expected one artifact, got %d", len(got.Artifacts))
	}
	sum := sha256.Sum256(payload)
	wantHash := hex.EncodeToString(sum[:])
	item := got.Artifacts[0]
	if item.Platform != "linux-x64" || item.Path != "linux-x64/game" || item.Size != int64(len(payload)) || item.SHA256 != wantHash {
		t.Fatalf("unexpected artifact record: %+v", item)
	}
}

func TestWriteArtifactManifestRejectsEmptyDirectory(t *testing.T) {
	root := t.TempDir()
	err := writeArtifactManifest(root, filepath.Join(root, "manifest.json"), "", "")
	if err == nil {
		t.Fatal("expected empty artifact directory to fail")
	}
}
