package godot

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlatformDir(t *testing.T) {
	tests := map[string]string{
		"Windows Desktop": "windows-x64",
		"Linux/X11": "",
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
