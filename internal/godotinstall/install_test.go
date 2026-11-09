package godotinstall

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestReleaseTag(t *testing.T) {
	for input, want := range map[string]string{"4.3": "4.3-stable", "3.5.3": "3.5.3-stable", "4.3-stable": "4.3-stable"} {
		got, _, err := releaseTag(input)
		if err != nil || got != want {
			t.Fatalf("releaseTag(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
	for _, input := range []string{"", "4", "5.0", "4.x", "4.3.1.2"} {
		if _, _, err := releaseTag(input); err == nil {
			t.Errorf("releaseTag(%q) should fail", input)
		}
	}
}

func TestAssetFor(t *testing.T) {
	tests := []struct {
		version string
		major int
		goos string
		goarch string
		want string
	}{
		{"4.3-stable", 4, "linux", "amd64", "Godot_v4.3-stable_linux.x86_64.zip"},
		{"4.3-stable", 4, "darwin", "arm64", "Godot_v4.3-stable_macos.universal.zip"},
		{"4.3-stable", 4, "windows", "amd64", "Godot_v4.3-stable_win64.exe.zip"},
		{"3.5-stable", 3, "linux", "amd64", "Godot_v3.5-stable_x11.64.zip"},
		{"3.5-stable", 3, "darwin", "amd64", "Godot_v3.5-stable_osx.universal.zip"},
	}
	for _, test := range tests {
		got, _, err := assetFor(test.version, test.major, test.goos, test.goarch)
		if err != nil || got != test.want {
			t.Errorf("assetFor(%s, %s/%s) = %q, %v; want %q", test.version, test.goos, test.goarch, got, err, test.want)
		}
	}
	if _, _, err := assetFor("3.5-stable", 3, "linux", "arm64"); err == nil {
		t.Fatal("unsupported Godot 3 ARM64 build should fail")
	}
}

func TestExtractArchiveRejectsTraversal(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "bad.zip")
	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create("../outside.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("bad")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(t.TempDir(), "engine")
	if err := extractArchive(archive, destination); err == nil {
		t.Fatal("archive traversal was accepted")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(destination), "outside.txt")); !os.IsNotExist(err) {
		t.Fatalf("archive wrote outside destination: %v", err)
	}
}

func TestFindExecutable(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "Godot_v4.3-stable_linux.x86_64")
	if err := os.WriteFile(path, []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	console := filepath.Join(root, "Godot_v4.3-stable_linux.x86_64_console")
	if err := os.WriteFile(console, []byte("console"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := findExecutable(root, "linux")
	if err != nil || got != path {
		t.Fatalf("findExecutable = %q, %v; want %q", got, err, path)
	}
}

func TestWriteLauncher(t *testing.T) {
	root := t.TempDir()
	executable := filepath.Join(root, "Godot_v4.3-stable_linux.x86_64")
	if err := os.WriteFile(executable, []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeLauncher(root, executable, "linux"); err != nil {
		t.Fatal(err)
	}
	launcherInfo, err := os.Stat(filepath.Join(root, "bin", "godot"))
	if err != nil {
		t.Fatal(err)
	}
	executableInfo, err := os.Stat(executable)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(launcherInfo, executableInfo) {
		t.Fatal("Linux launcher should be a hard link to the managed executable")
	}
}

func TestEnsureCacheParentRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "godot")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := ensureCacheParent(root, filepath.Join(root, "godot", "4.3-stable")); err == nil {
		t.Fatal("cache setup followed a symlinked parent")
	}
	entries, err := os.ReadDir(outside)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("cache setup wrote outside the managed root: %v", entries)
	}
}

func TestCleanupEmptyCacheDirsRemovesOnlyEmptyChildren(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "godot", "4.3-stable")
	if err := os.MkdirAll(parent, 0o700); err != nil {
		t.Fatal(err)
	}
	cleanupEmptyCacheDirs(root, parent)
	if _, err := os.Stat(parent); !os.IsNotExist(err) {
		t.Fatalf("empty cache directories remain: %v", err)
	}
	if _, err := os.Stat(root); err != nil {
		t.Fatalf("shared cache root should remain: %v", err)
	}
}
