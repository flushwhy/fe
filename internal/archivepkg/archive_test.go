package archivepkg

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCreateZipAndTarGz(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "bin", "fe"), []byte("binary"), 0755); err != nil {
		t.Fatal(err)
	}

	zipPath := filepath.Join(t.TempDir(), "fe.zip")
	if err := Create(root, zipPath, "zip"); err != nil {
		t.Fatal(err)
	}
	z, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(z.File) != 1 || z.File[0].Name != "bin/fe" {
		t.Fatalf("unexpected ZIP entries: %#v", z.File)
	}
	_ = z.Close()

	tarPath := filepath.Join(t.TempDir(), "fe.tar.gz")
	if err := Create(root, tarPath, "tar.gz"); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(tarPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	entry, err := tar.NewReader(gz).Next()
	if err != nil {
		t.Fatal(err)
	}
	if entry.Name != "bin/fe" {
		t.Fatalf("unexpected TAR entry: %q", entry.Name)
	}
}

func TestCreateRejectsWindowsAmbiguousArchiveNames(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("backslash is a path separator on Windows")
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "bad\\name.txt"), []byte("unsafe"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Create(root, filepath.Join(t.TempDir(), "unsafe.zip"), "zip"); err == nil {
		t.Fatal("Create() accepted a filename containing a backslash")
	}
}

func TestCreateRejectsSymlinkOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation may require elevated privileges")
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "input"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "target")
	if err := os.WriteFile(target, []byte("keep me"), 0644); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "archive.zip")
	if err := os.Symlink(target, output); err != nil {
		t.Fatal(err)
	}
	if err := Create(root, output, "zip"); err == nil {
		t.Fatal("Create() accepted a symlink output")
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "keep me" {
		t.Fatalf("symlink target was modified: %q", got)
	}
}

func TestCreateDoesNotReplaceExistingArchiveOnFailure(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "bad\\name"), []byte("unsafe"), 0644); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "archive.zip")
	if err := os.WriteFile(output, []byte("previous archive"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Create(root, output, "zip"); err == nil {
		t.Fatal("Create() unexpectedly succeeded")
	}
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "previous archive" {
		t.Fatalf("existing archive was modified: %q", got)
	}
}
