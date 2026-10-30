package inspect

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const maxPythonMetadataBytes = 64 * 1024

// Project describes one project ecosystem recognized from files on disk.
type Project struct {
	Type    string
	Markers []string
}

// Report is the result of inspecting a project directory.
type Report struct {
	Root     string
	Projects []Project
	Files    int
}

// Inspect detects common project types without executing project code.
func Inspect(root string) (*Report, error) {
	if root == "" {
		root = "."
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve project path: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("access project path %q: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("project path %q is not a directory", root)
	}

	markers := map[string][]string{}
	add := func(kind, path string) {
		markers[kind] = append(markers[kind], filepath.ToSlash(path))
	}

	err = filepath.WalkDir(abs, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == abs {
			return nil
		}
		rel, err := filepath.Rel(abs, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", ".fe", ".venv", "venv", "node_modules", "vendor", "build", "builds", "dist", "__pycache__", ".idea", ".vscode":
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}

		reported := false
		switch entry.Name() {
		case "go.mod":
			add("Go", rel); reported = true
		case "Cargo.toml":
			add("Rust", rel); reported = true
		case "pyproject.toml", "requirements.txt", "Pipfile", "setup.py":
			add("Python", rel); reported = true
		case "package.json":
			add("Node.js", rel); reported = true
		case "CMakeLists.txt", "Makefile":
			add("C/C++", rel); reported = true
		case "project.godot":
			add("Godot", rel); reported = true
		case "blender_manifest.toml":
			add("Blender add-on", rel); reported = true
		case "engine.json":
			add("O3DE", rel); reported = true
		}
		if reported {
			return nil
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".rockspec") {
			add("Lua", rel)
			return nil
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".py") {
			legacy, err := hasBlenderMetadata(path)
			if err != nil {
				return fmt.Errorf("inspect %s: %w", rel, err)
			}
			if legacy {
				add("Blender add-on", rel)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan project: %w", err)
	}

	types := make([]string, 0, len(markers))
	for kind := range markers {
		types = append(types, kind)
	}
	sort.Strings(types)

	report := &Report{Root: abs}
	for _, kind := range types {
		sort.Strings(markers[kind])
		report.Projects = append(report.Projects, Project{Type: kind, Markers: markers[kind]})
	}
	return report, nil
}

func hasBlenderMetadata(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	reader := bufio.NewReader(ioLimitReader{r: file, remaining: maxPythonMetadataBytes})
	for {
		line, err := reader.ReadString('\n')
		if strings.Contains(line, "bl_info") {
			return true, nil
		}
		if err != nil {
			return false, nil
		}
	}
}

type ioLimitReader struct {
	r         *os.File
	remaining int
}

func (r ioLimitReader) Read(p []byte) (int, error) {
	if r.remaining <= 0 {
		return 0, os.ErrClosed
	}
	if len(p) > r.remaining {
		p = p[:r.remaining]
	}
	n, err := r.r.Read(p)
	r.remaining -= n
	return n, err
}
