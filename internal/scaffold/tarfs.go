package scaffold

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

// memFile implements fs.File for in-memory content.
type memFile struct {
	name    string
	content []byte
	reader  *bytes.Reader
}

func (f *memFile) Read(p []byte) (int, error)  { return f.reader.Read(p) }
func (f *memFile) Close() error                 { return nil }
func (f *memFile) Stat() (fs.FileInfo, error)   { return &memFileInfo{f}, nil }

type memFileInfo struct{ f *memFile }

func (i *memFileInfo) Name() string      { return filepath.Base(i.f.name) }
func (i *memFileInfo) Size() int64       { return int64(len(i.f.content)) }
func (i *memFileInfo) Mode() fs.FileMode { return 0444 }
func (i *memFileInfo) ModTime() time.Time { return time.Time{} }
func (i *memFileInfo) IsDir() bool       { return false }
func (i *memFileInfo) Sys() interface{}  { return nil }

// memFS is an in-memory fs.FS built from a tar.gz payload.
type memFS map[string][]byte

func (m memFS) Open(name string) (fs.File, error) {
	content, ok := m[name]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return &memFile{name: name, content: content, reader: bytes.NewReader(content)}, nil
}

func (m memFS) ReadDir(name string) ([]fs.DirEntry, error) {
	prefix := name
	if prefix != "." {
		prefix += "/"
	}

	seen := map[string]bool{}
	var entries []fs.DirEntry

	for path := range m {
		rel := path
		if prefix != "./" {
			if !strings.HasPrefix(path, prefix) {
				continue
			}
			rel = strings.TrimPrefix(path, prefix)
		}

		parts := strings.SplitN(rel, "/", 2)
		entryName := parts[0]
		if seen[entryName] {
			continue
		}
		seen[entryName] = true

		isDir := len(parts) > 1
		entries = append(entries, &memDirEntry{name: entryName, isDir: isDir})
	}

	return entries, nil
}

type memDirEntry struct {
	name  string
	isDir bool
}

func (e *memDirEntry) Name() string               { return e.name }
func (e *memDirEntry) IsDir() bool                { return e.isDir }
func (e *memDirEntry) Type() fs.FileMode          { return 0 }
func (e *memDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

// extractTarGz reads a gzipped tar archive into an in-memory fs.FS.
func extractTarGz(data []byte) (fs.FS, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("not a valid gzip archive: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	mem := make(memFS)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading tar: %w", err)
		}
		if hdr.Typeflag == tar.TypeDir {
			continue
		}

		content, err := io.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("reading tar entry %s: %w", hdr.Name, err)
		}

		// Strip any leading directory component (e.g. "odin/main.odin" → "main.odin").
		name := hdr.Name
		if idx := strings.Index(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}

		mem[name] = content
	}

	return mem, nil
}
