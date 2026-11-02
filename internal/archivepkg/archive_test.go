package archivepkg

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
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
