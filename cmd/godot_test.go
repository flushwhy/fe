package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/flushwhy/fe/internal/projectstate"
	"github.com/spf13/cobra"
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
	if err := godotExportCmd.Flags().Set("preset", ""); err != nil {
		t.Fatal(err)
	}
	if err := godotExportCmd.Flags().Set("output", ""); err != nil {
		t.Fatal(err)
	}
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

func TestGodotExportRejectsUnknownPreset(t *testing.T) {
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "project.godot"), []byte("config_version=5\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	presets := `[preset.0]
name="Linux"
platform="Linux"
export_path="game.x86_64"
`
	if err := os.WriteFile(filepath.Join(project, "export_presets.cfg"), []byte(presets), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(project)
	if err := godotExportCmd.Flags().Set("preset", "Windows"); err != nil {
		t.Fatal(err)
	}
	if err := godotExportCmd.Flags().Set("output", "build/game.exe"); err != nil {
		t.Fatal(err)
	}

	err := godotExportCmd.RunE(godotExportCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "not found in export_presets.cfg") {
		t.Fatalf("expected unknown-preset error, got %v", err)
	}
}

func TestSelectGodotEngineRejectsSingleIncompatibleEngine(t *testing.T) {
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "project.godot"), []byte(`config/features=PackedStringArray("4.3", "GL Compatibility")`), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := selectGodotEngineFrom(project, "", []godotEngine{{Path: "godot-4.2", Version: "4.2.stable.official"}})
	if err == nil || !strings.Contains(err.Error(), "project requires Godot 4.3") {
		t.Fatalf("expected incompatible engine error, got %v", err)
	}
}

func TestSelectGodotEngineMatchesProjectVersion(t *testing.T) {
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "project.godot"), []byte(`config/features=PackedStringArray("4.3", "GL Compatibility")`), 0600); err != nil {
		t.Fatal(err)
	}
	engines := []godotEngine{
		{Path: "godot-4.2", Version: "4.2.stable.official"},
		{Path: "godot-4.3", Version: "4.3.1.stable.official"},
	}
	got, err := selectGodotEngineFrom(project, "", engines)
	if err != nil {
		t.Fatal(err)
	}
	if got.Path != "godot-4.3" {
		t.Fatalf("selected %q, want godot-4.3", got.Path)
	}
}

func TestSelectGodotEngineUsesOnlyEngineWhenProjectVersionIsUnknown(t *testing.T) {
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "project.godot"), []byte("config_version=5\n"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := selectGodotEngineFrom(project, "", []godotEngine{{Path: "godot", Version: "4.3.stable.official"}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Path != "godot" {
		t.Fatalf("selected %q, want godot", got.Path)
	}
}

func TestGodotVersionMatchesWholeComponent(t *testing.T) {
	if !godotVersionMatches("4.3.1.stable.official", "4.3") {
		t.Fatal("4.3 should match 4.3.1")
	}
	if godotVersionMatches("4.30.stable.official", "4.3") {
		t.Fatal("4.3 must not match 4.30")
	}
	for _, prerelease := range []string{"4.3.rc1.official", "4.3.beta2.official", "4.3.dev.custom_build"} {
		if godotVersionMatches(prerelease, "4.3") {
			t.Errorf("stable request must not match prerelease %q", prerelease)
		}
	}
}


func TestGodotDoctorDoesNotRequireNativeExtensionForRegularProject(t *testing.T) {
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "project.godot"), []byte("config_version=5\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(project)

	var output bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&output)
	if err := godotDoctorCmd.RunE(cmd, nil); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "godot-cpp: not found") {
		t.Fatalf("regular Godot project should not require godot-cpp; output:\n%s", output.String())
	}
}

func TestGodotDoctorReportsMissingGodotCppForNativeScaffold(t *testing.T) {
	project := t.TempDir()
	for path, content := range map[string]string{
		"project.godot": "config_version=5\n",
		"native/CMakeLists.txt": "project(test)",
	} {
		fullPath := filepath.Join(project, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(project)

	var output bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&output)
	if err := godotDoctorCmd.RunE(cmd, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "godot-cpp: not found") {
		t.Fatalf("native scaffold should report missing godot-cpp; output:\n%s", output.String())
	}
}

func TestSelectGodotEngineUsesProjectManagedVersion(t *testing.T) {
	if os.PathSeparator == '\\' {
		t.Skip("uses a Unix executable fixture")
	}
	t.Setenv("FE_HOME", t.TempDir())
	t.Setenv("PATH", t.TempDir())
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "project.godot"), []byte(`config/features=PackedStringArray("4.3", "GL Compatibility")`), 0o644); err != nil {
		t.Fatal(err)
	}
	cache, err := projectstate.ToolCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	platform := "linux-x86_64"
	executableName := "Godot_v4.3-stable_linux.x86_64"
	if runtime.GOOS == "darwin" {
		platform = "macos-universal"
		executableName = filepath.Join("Godot.app", "Contents", "MacOS", "Godot")
	}
	toolPath := filepath.Join(cache, "godot", "4.3-stable", platform)
	executable := filepath.Join(toolPath, executableName)
	if err := os.MkdirAll(filepath.Dir(executable), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(executable, []byte("#!/bin/sh\necho 4.3.stable.official\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(toolPath, ".fe-managed"), []byte("godot:4.3-stable\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := projectstate.Sync(project); err != nil {
		t.Fatal(err)
	}
	if err := projectstate.RegisterTool(project, "godot:4.3-stable", toolPath); err != nil {
		t.Fatal(err)
	}
	got, err := selectGodotEngine(project, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Path != executable || !godotVersionMatches(got.Version, "4.3") {
		t.Fatalf("selected %#v, want managed engine %s", got, executable)
	}
}

func TestGodotInstallIsDisabledInCI(t *testing.T) {
	t.Setenv("CI", "true")
	if err := godotInstallCmd.Flags().Set("version", "4.3"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = godotInstallCmd.Flags().Set("version", "") })
	err := godotInstallCmd.RunE(godotInstallCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "disabled in CI") {
		t.Fatalf("expected CI guard, got %v", err)
	}
}

func TestActivateProjectGodotPathIsTemporary(t *testing.T) {
	for _, key := range []string{"CI", "GITHUB_ACTIONS", "JENKINS_URL", "BUILDKITE", "BUILD_ID", "TF_BUILD"} {
		t.Setenv(key, "")
	}
	t.Setenv("FE_HOME", t.TempDir())
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "project.godot"), []byte(`config/features=PackedStringArray("4.3", "GL Compatibility")`), 0o644); err != nil {
		t.Fatal(err)
	}
	cache, err := projectstate.ToolCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	toolPath := filepath.Join(cache, "godot", "4.3-stable", "linux-x86_64")
	bin := filepath.Join(toolPath, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(toolPath, ".fe-managed"), []byte("godot:4.3-stable"), 0o600); err != nil {
		t.Fatal(err)
	}
	launcher := "godot"
	if runtime.GOOS == "windows" {
		launcher = "godot.exe"
	}
	if err := os.WriteFile(filepath.Join(bin, launcher), []byte("launcher"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := projectstate.Sync(project); err != nil {
		t.Fatal(err)
	}
	if err := projectstate.RegisterTool(project, "godot:4.3-stable", toolPath); err != nil {
		t.Fatal(err)
	}
	oldPath := os.Getenv("PATH")
	restore, err := activateProjectGodotPath(project)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(os.Getenv("PATH"), bin+string(os.PathListSeparator)) {
		restore()
		t.Fatalf("project bin was not prepended to PATH: %s", os.Getenv("PATH"))
	}
	restore()
	if os.Getenv("PATH") != oldPath {
		t.Fatal("PATH was not restored after project command")
	}
}
