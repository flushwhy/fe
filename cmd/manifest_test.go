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

func TestWriteArtifactManifestRejectsSymlinkOutput(t *testing.T) {
	root := t.TempDir()
	artifact := filepath.Join(root, "game.bin")
	if err := os.WriteFile(artifact, []byte("game"), 0o600); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "keep.txt")
	const sentinel = "do not overwrite"
	if err := os.WriteFile(target, []byte(sentinel), 0o600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(root, "manifest.json")
	if err := os.Symlink(target, output); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	if err := writeArtifactManifest(root, output, "", ""); err == nil {
		t.Fatal("expected symlink output to be rejected")
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != sentinel {
		t.Fatalf("symlink target was modified: got %q", got)
	}
}

func TestWriteManifestAtomicallyReplacesFile(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "manifest.json")
	if err := os.WriteFile(output, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeManifestAtomically(output, []byte("new")); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Fatalf("got %q, want new content", got)
	}
}

func TestWriteArtifactManifestAcceptsSymlinkedRoot(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "artifacts")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "game.bin"), []byte("game"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(parent, "artifact-link")
	if err := os.Symlink(root, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	output := filepath.Join(parent, "manifest.json")
	if err := writeArtifactManifest(link, output, "", ""); err != nil {
		t.Fatalf("manifest from symlinked root: %v", err)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	var got artifactManifest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Artifacts) != 1 || got.Artifacts[0].Path != "game.bin" {
		t.Fatalf("unexpected artifacts: %#v", got.Artifacts)
	}
}
