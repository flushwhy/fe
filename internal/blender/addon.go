package blender

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

const maxMetadataBytes = 64 * 1024

var (
	extensionIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	versionPattern    = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)
)

type Manifest struct {
	SchemaVersion    string   `toml:"schema_version"`
	ID               string   `toml:"id"`
	Version          string   `toml:"version"`
	Name             string   `toml:"name"`
	Tagline          string   `toml:"tagline"`
	Maintainer       string   `toml:"maintainer"`
	Type             string   `toml:"type"`
	BlenderVersionMin string  `toml:"blender_version_min"`
	License          []string `toml:"license"`
}

type Addon struct {
	Root       string
	ID         string
	Name       string
	Version    string
	Legacy     bool
	Manifest   Manifest
}

type Validation struct {
	Addon    Addon
	Warnings []string
}

// Inspect recognizes a Blender extension manifest or a legacy add-on entry point.
func Inspect(root string) (Addon, error) {
	if root == "" {
		root = "."
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return Addon{}, fmt.Errorf("resolve add-on path: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return Addon{}, fmt.Errorf("access add-on path %q: %w", root, err)
	}
	if !info.IsDir() {
		return Addon{}, fmt.Errorf("add-on path %q is not a directory", root)
	}

	addon := Addon{Root: abs, ID: filepath.Base(abs), Name: filepath.Base(abs), Version: "dev"}
	manifestPath := filepath.Join(abs, "blender_manifest.toml")
	data, err := os.ReadFile(manifestPath)
	if err == nil {
		if err := toml.Unmarshal(data, &addon.Manifest); err != nil {
			return Addon{}, fmt.Errorf("parse %s: %w", manifestPath, err)
		}
		addon.ID = addon.Manifest.ID
		addon.Name = addon.Manifest.Name
		addon.Version = addon.Manifest.Version
		return addon, nil
	}
	if !os.IsNotExist(err) {
		return Addon{}, fmt.Errorf("read %s: %w", manifestPath, err)
	}

	entry := filepath.Join(abs, "__init__.py")
	data, err = os.ReadFile(entry)
	if err != nil {
		return Addon{}, fmt.Errorf("no blender_manifest.toml or legacy __init__.py found in %q", abs)
	}
	if len(data) > maxMetadataBytes {
		data = data[:maxMetadataBytes]
	}
	if !bytes.Contains(data, []byte("bl_info")) {
		return Addon{}, fmt.Errorf("legacy add-on %q has no bl_info metadata in __init__.py", abs)
	}
	addon.Legacy = true
	return addon, nil
}

// Validate checks Blender extension metadata and parses Python source without
// importing or executing add-on code.
func Validate(root string) (Validation, error) {
	addon, err := Inspect(root)
	if err != nil {
		return Validation{}, err
	}
	if !addon.Legacy {
		if err := validateManifest(addon); err != nil {
			return Validation{}, err
		}
	}

	report := Validation{Addon: addon}
	python, err := findPython()
	if err != nil {
		report.Warnings = append(report.Warnings, "Python syntax was not checked: install Python or add it to PATH")
		return report, nil
	}
	files, err := pythonFiles(addon.Root)
	if err != nil {
		return Validation{}, err
	}
	if len(files) == 0 {
		return report, nil
	}
	script := "import ast, pathlib, sys; [ast.parse(pathlib.Path(p).read_text(encoding='utf-8'), filename=p) for p in sys.argv[1:]]"
	args := append([]string{"-c", script}, files...)
	cmd := exec.Command(python, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return Validation{}, fmt.Errorf("Python syntax check failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return report, nil
}

func validateManifest(addon Addon) error {
	m := addon.Manifest
	var problems []string
	required := map[string]string{
		"schema_version":      m.SchemaVersion,
		"id":                  m.ID,
		"version":             m.Version,
		"name":                m.Name,
		"tagline":             m.Tagline,
		"maintainer":          m.Maintainer,
		"type":                m.Type,
		"blender_version_min": m.BlenderVersionMin,
	}
	for field, value := range required {
		if strings.TrimSpace(value) == "" {
			problems = append(problems, field+" is required")
		}
	}
	if m.SchemaVersion != "" && m.SchemaVersion != "1.0.0" {
		problems = append(problems, "schema_version must be 1.0.0")
	}
	if m.ID != "" && !extensionIDPattern.MatchString(m.ID) {
		problems = append(problems, "id must start with a lowercase letter and contain only lowercase letters, digits, and underscores")
	}
	if m.Version != "" && !versionPattern.MatchString(m.Version) {
		problems = append(problems, "version must use semantic versioning (major.minor.patch)")
	}
	if m.BlenderVersionMin != "" && !versionPattern.MatchString(m.BlenderVersionMin) {
		problems = append(problems, "blender_version_min must use semantic versioning (major.minor.patch)")
	}
	if len(m.License) == 0 {
		problems = append(problems, "license must contain at least one SPDX license identifier")
	}
	if m.Type != "" && m.Type != "add-on" {
		problems = append(problems, "only Blender add-on extensions (type = \"add-on\") are supported")
	}
	if len(m.Tagline) > 64 {
		problems = append(problems, "tagline must be 64 characters or fewer")
	}
	if strings.HasSuffix(strings.TrimSpace(m.Tagline), ".") {
		problems = append(problems, "tagline must not end with a period")
	}
	if _, err := os.Stat(filepath.Join(addon.Root, "__init__.py")); err != nil {
		problems = append(problems, "add-on extension must contain __init__.py")
	}
	if len(problems) > 0 {
		return fmt.Errorf("invalid Blender extension manifest: %s", strings.Join(problems, "; "))
	}
	return nil
}

func findPython() (string, error) {
	for _, name := range []string{"python3", "python"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("Python interpreter not found")
}

func pythonFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", ".fe", ".venv", "venv", "__pycache__", "build", "dist", "builds":
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type().IsRegular() && strings.EqualFold(filepath.Ext(entry.Name()), ".py") {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

// Package creates a deterministic Blender-installable ZIP. Extension files
// live at the archive root; legacy add-ons retain their package directory.
func Package(root, output string) error {
	addon, err := Inspect(root)
	if err != nil {
		return err
	}
	if !addon.Legacy {
		if err := validateManifest(addon); err != nil {
			return err
		}
	}

	output, err = filepath.Abs(output)
	if err != nil {
		return fmt.Errorf("resolve package output: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return fmt.Errorf("create package output directory: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(output), ".blender-addon-*.zip")
	if err != nil {
		return fmt.Errorf("create temporary package: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	writer := zip.NewWriter(tmp)
	count := 0
	err = filepath.WalkDir(addon.Root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == addon.Root {
			return nil
		}
		if entry.IsDir() {
			if excludedDirectory(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() || excludedFile(entry.Name()) {
			return nil
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if abs == output {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(addon.Root, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		if addon.Legacy {
			name = filepath.ToSlash(filepath.Join(filepath.Base(addon.Root), rel))
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = name
		header.Method = zip.Deflate
		header.Modified = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)
		target, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}
		source, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(target, source)
		closeErr := source.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		count++
		return nil
	})
	closeErr := writer.Close()
	fileCloseErr := tmp.Close()
	if err != nil {
		return fmt.Errorf("package Blender add-on: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("finish Blender package: %w", closeErr)
	}
	if fileCloseErr != nil {
		return fmt.Errorf("close Blender package: %w", fileCloseErr)
	}
	if count == 0 {
		return fmt.Errorf("add-on directory %q contains no packageable files", addon.Root)
	}
	if err := os.Rename(tmpPath, output); err != nil {
		return fmt.Errorf("publish Blender package %q: %w", output, err)
	}
	return nil
}

func excludedDirectory(name string) bool {
	switch name {
	case ".git", ".fe", ".venv", "venv", "__pycache__", ".pytest_cache", ".mypy_cache", "build", "builds", "dist":
		return true
	default:
		return false
	}
}

func excludedFile(name string) bool {
	lower := strings.ToLower(name)
	return lower == ".ds_store" || lower == "thumbs.db" ||
		strings.HasSuffix(lower, ".pyc") || strings.HasSuffix(lower, ".pyo") ||
		strings.HasSuffix(lower, ".zip")
}
