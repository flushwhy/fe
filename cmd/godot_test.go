package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGodotExtensionInitCreatesManifestAndBuildOutputs(t *testing.T) {
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "project.godot"), []byte("config_version=5\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(project)

	if err := godotExtensionInitCmd.RunE(godotExtensionInitCmd, nil); err != nil {
		t.Fatal(err)
	}

	manifest, err := os.ReadFile(filepath.Join(project, "fe_godot_extension.gdextension"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"entry_symbol = \"fe_extension_init\"",
		"linux.debug.x86_64 = \"res://bin/libfe_godot_extension.so\"",
		"windows.debug.x86_64 = \"res://bin/fe_godot_extension.dll\"",
		"macos.debug = \"res://bin/libfe_godot_extension.dylib\"",
	} {
		if !strings.Contains(string(manifest), want) {
			t.Errorf("manifest missing %q", want)
		}
	}

	cmake, err := os.ReadFile(filepath.Join(project, "native", "CMakeLists.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"RUNTIME_OUTPUT_DIRECTORY", "SUFFIX \".dll\"", "SUFFIX \".dylib\"", "SUFFIX \".so\""} {
		if !strings.Contains(string(cmake), want) {
			t.Errorf("CMakeLists.txt missing %q", want)
		}
	}
}

func TestGodotExportRequiresPreset(t *testing.T) {
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "project.godot"), []byte("config_version=5\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(project)

	err := godotExportCmd.RunE(godotExportCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "export preset is required") {
		t.Fatalf("expected missing-preset error, got %v", err)
	}
}

func TestGodotExportRequiresExportPresetsFile(t *testing.T) {
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "project.godot"), []byte("config_version=5\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(project)
	if err := godotExportCmd.Flags().Set("preset", "Linux"); err != nil {
		t.Fatal(err)
	}
	if err := godotExportCmd.Flags().Set("output", "build/game.x86_64"); err != nil {
		t.Fatal(err)
	}

	err := godotExportCmd.RunE(godotExportCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "export presets not found") {
		t.Fatalf("expected missing-export-presets error, got %v", err)
	}
}
