package godot

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlatformDir(t *testing.T) {
	tests := map[string]string{
		"Windows Desktop": "windows-x64",
		"Linux/X11": "linux-x64",
		"Linux": "linux-x64",
		"macOS": "macos-arm64",
	}
	for platform, want := range tests {
		got, err := PlatformDir(platform)
		if want == "" {
			if err == nil {
				t.Errorf("PlatformDir(%q) expected error, got %q", platform, got)
			}
			continue
		}
		if err != nil || got != want {
			t.Errorf("PlatformDir(%q) = %q, %v; want %q", platform, got, err, want)
		}
	}
}

func TestReadPresets(t *testing.T) {
	path := filepath.Join(t.TempDir(), "export_presets.cfg")
	content := "[preset.0]\nname=\"Windows Desktop\"\nplatform=\"Windows Desktop\"\nexport_path=\"build/game.exe\"\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	presets, err := ReadPresets(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(presets) != 1 || presets[0].Name != "Windows Desktop" || presets[0].Platform != "Windows Desktop" || presets[0].ExportPath != "build/game.exe" {
		t.Fatalf("unexpected presets: %#v", presets)
	}
}

func TestReadPresetsDetectsDedicatedServer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "export_presets.cfg")
	content := "[preset.0]\nname=\"Linux Server\"\nplatform=\"Linux\"\ndedicated_server=true\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	presets, err := ReadPresets(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(presets) != 1 || !presets[0].DedicatedServer {
		t.Fatalf("expected dedicated-server preset, got %#v", presets)
	}
}

func TestFindLinuxServerPreset(t *testing.T) {
	presets := []Preset{
		{Name: "Linux Desktop", Platform: "Linux", DedicatedServer: false},
		{Name: "Windows Server", Platform: "Windows Desktop", DedicatedServer: true},
		{Name: "Linux Dedicated", Platform: "Linux", DedicatedServer: true},
	}
	got, err := FindLinuxServerPreset(presets)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Linux Dedicated" {
		t.Fatalf("selected %q, want Linux Dedicated", got.Name)
	}
}

func TestFindLinuxServerPresetRejectsMissingAndAmbiguousMatches(t *testing.T) {
	_, err := FindLinuxServerPreset([]Preset{{Name: "Linux", Platform: "Linux"}})
	if err == nil || !strings.Contains(err.Error(), "no Linux dedicated-server") {
		t.Fatalf("expected missing server preset error, got %v", err)
	}
	_, err = FindLinuxServerPreset([]Preset{
		{Name: "Server A", Platform: "Linux", DedicatedServer: true},
		{Name: "Server B", Platform: "Linux", DedicatedServer: true},
	})
	if err == nil || !strings.Contains(err.Error(), "multiple Linux dedicated-server presets") {
		t.Fatalf("expected ambiguous server preset error, got %v", err)
	}
}

func TestExpectedOutputPath(t *testing.T) {
	got, err := ExpectedOutputPath("/projects/my-game", "builds", Preset{
		Name: "Windows Desktop", Platform: "Windows Desktop",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("builds", "windows-x64", "my-game.exe")
	if got != want {
		t.Fatalf("ExpectedOutputPath() = %q, want %q", got, want)
	}
	got, err = ExpectedOutputPath("/projects/my-game", "builds", Preset{
		Name: "Linux", Platform: "Linux", ExportPath: "game.x86_64",
	})
	if err != nil {
		t.Fatal(err)
	}
	want = filepath.Join("builds", "linux-x64", "game.x86_64")
	if got != want {
		t.Fatalf("ExpectedOutputPath() = %q, want %q", got, want)
	}
}


func TestExportResolvesRelativeOutputFromProjectRoot(t *testing.T) {
	projectDir := t.TempDir()
	presetFile := filepath.Join(projectDir, "export_presets.cfg")
	presets := "[preset.0]\nname=\"Windows Desktop\"\nplatform=\"Windows Desktop\"\nexport_path=\"game.exe\"\n"
	if err := os.WriteFile(presetFile, []byte(presets), 0o600); err != nil {
		t.Fatal(err)
	}

	var gotArgs []string
	runner := func(_ string, args ...string) *exec.Cmd {
		gotArgs = append([]string(nil), args...)
		return exec.Command("go", "version")
	}
	if err := Export(projectDir, "godot", "artifacts", "", runner); err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(projectDir, "artifacts", "windows-x64", "game.exe")
	if got := gotArgs[len(gotArgs)-1]; got != want {
		t.Fatalf("export destination = %q, want %q", got, want)
	}
	if _, err := os.Stat(filepath.Dir(want)); err != nil {
		t.Fatalf("export directory was not created in project: %v", err)
	}
}

