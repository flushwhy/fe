package scaffold

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"strings"
	"testing"
)

func makeTarGz(t *testing.T, headers []tar.Header) []byte {
	t.Helper()
	var data bytes.Buffer
	gz := gzip.NewWriter(&data)
	tw := tar.NewWriter(gz)
	for i := range headers {
		if err := tw.WriteHeader(&headers[i]); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}

func TestExtractTarGzCountsDirectoryEntries(t *testing.T) {
	headers := make([]tar.Header, 501)
	for i := range headers {
		headers[i] = tar.Header{Name: fmt.Sprintf("repo/dir-%03d/", i), Typeflag: tar.TypeDir, Mode: 0755}
	}
	_, err := extractTarGz(makeTarGz(t, headers))
	if err == nil || !strings.Contains(err.Error(), "too many entries") {
		t.Fatalf("extractTarGz error = %v, want entry limit error", err)
	}
}

func TestExtractTarGzRejectsAmbiguousPaths(t *testing.T) {
	data := makeTarGz(t, []tar.Header{{Name: `repo\\evil.txt`, Typeflag: tar.TypeDir, Mode: 0755}})
	if _, err := extractTarGz(data); err == nil {
		t.Fatal("extractTarGz accepted a path containing a backslash")
	}
}
