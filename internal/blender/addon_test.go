package blender

import (
	"archive/zip"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)
const validManifest = `schema_version = "1.0.0"
id = "sample_addon"
version = "1.2.3"
name = "Sample Add-on"
tagline = "A sample Blender add-on"
maintainer = "Example Developer"
type = ["add-on"]
blender_version_min = "4.2.0"
license = ["SPDX:GPL-3.0-or-later"]
`

func writeFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestInspectAndPackageExtension(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "blender_manifest.toml", validManifest)
	writeFile(t, root, "__init__.py", "bl_info = {}\n")
	writeFile(t, root, "operators.py", "def register(): pass\n")
	writeFile(t, root, ".git/config", "ignored")
	writeFile(t, root, "__pycache__/operators.pyc", "ignored")
	writeFile(t, root, "dist/old.zip", "ignored")

	addon, err := Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if addon.Legacy || addon.ID != "sample_addon" || addon.Version != "1.2.3" {
		t.Fatalf("unexpected add-on metadata: %#v", addon)
	}
	if err := validateManifest(addon); err != nil {
		t.Fatal(err)
	}

	output := filepath.Join(t.TempDir(), "sample_addon.zip")
	if err := Package(root, output); err != nil {
		t.Fatal(err)
	}
	archive, err := zip.OpenReader(output)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()

	var names []string
	for _, file := range archive.File {
		names = append(names, file.Name)
	}
	want := []string{"__init__.py", "blender_manifest.toml", "operators.py"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("archive entries = %v, want %v", names, want)
	}
}

func TestInspectAndPackageLegacyAddon(t *testing.T) {
	root := filepath.Join(t.TempDir(), "legacy_addon")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, root, "__init__.py", "bl_info = { 'name': 'Legacy' }\n")
	writeFile(t, root, "module.py", "VALUE = 1\n")

	addon, err := Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if !addon.Legacy || addon.ID != "legacy_addon" {
		t.Fatalf("unexpected legacy add-on: %#v", addon)
	}

	output := filepath.Join(t.TempDir(), "legacy.zip")
	if err := Package(root, output); err != nil {
		t.Fatal(err)
	}
	archive, err := zip.OpenReader(output)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	var names []string
	for _, file := range archive.File {
		names = append(names, file.Name)
	}
	want := []string{"legacy_addon/__init__.py", "legacy_addon/module.py"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("archive entries = %v, want %v", names, want)
	}
}

func TestValidateManifestRejectsInvalidMetadata(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "__init__.py", "bl_info = {}\n")
	writeFile(t, root, "blender_manifest.toml", `schema_version = "2.0.0"
id = "Bad-ID"
version = "not-a-version"
name = ""
tagline = "A tagline."
maintainer = ""
type = ["theme"]
blender_version_min = "4"
license = []
`)
	addon, err := Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateManifest(addon); err == nil {
		t.Fatal("expected invalid manifest error")
	}
}

func TestInspectRejectsUnknownDirectory(t *testing.T) {
	if _, err := Inspect(t.TempDir()); err == nil {
		t.Fatal("expected unrecognized add-on error")
	}
}


func TestPackageRejectsWindowsAmbiguousArchiveNames(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("backslash is a path separator on Windows")
	}
	root := t.TempDir()
	writeFile(t, root, "blender_manifest.toml", validManifest)
	writeFile(t, root, "__init__.py", "bl_info = {}\n")
	writeFile(t, root, "bad\\name.py", "VALUE = 1\n")
	if err := Package(root, filepath.Join(t.TempDir(), "unsafe.zip")); err == nil {
		t.Fatal("Package() accepted a filename containing a backslash")
	}
}
