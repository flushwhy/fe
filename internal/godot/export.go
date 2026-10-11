package godot

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

// Preset describes the fields fe needs from a Godot export preset.
type Preset struct {
	Name       string
	Platform   string
	ExportPath string
}

// ReadPresets loads named export presets from export_presets.cfg.
func ReadPresets(path string) ([]Preset, error) {
	cfg, err := ini.Load(path)
	if err != nil {
		return nil, fmt.Errorf("read Godot export presets %s: %w", path, err)
	}
	presets := make([]Preset, 0)
	for _, section := range cfg.Sections() {
		sectionName := section.Name()
		if !strings.HasPrefix(sectionName, "preset.") || strings.Contains(strings.TrimPrefix(sectionName, "preset."), ".") {
			continue
		}
		name := section.Key("name").String()
		platform := section.Key("platform").String()
		if name == "" || platform == "" {
			continue
		}
		presets = append(presets, Preset{Name: name, Platform: platform, ExportPath: section.Key("export_path").String()})
	}
	if len(presets) == 0 {
		return nil, fmt.Errorf("no valid presets found in %s", path)
	}
	return presets, nil
}

// PlatformDir maps Godot's built-in platform labels to fe/Butler's folder convention.
func PlatformDir(platform string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case "windows desktop":
		return "windows-x64", nil
	case "linux", "linux/x11", "linuxbsd", "linuxbsd headless":
		return "linux-x64", nil
	case "macos", "mac os x":
		return "macos-arm64", nil
	default:
		return "", fmt.Errorf("unsupported Godot export platform %q; configure a supported desktop preset", platform)
	}
}

// Export exports each configured preset into a Butler-compatible platform folder.
// A preset's configured export_path supplies the filename/extension, while fe controls its directory.
func Export(projectDir, executable, outputDir, onlyPreset string, runner func(string, ...string) *exec.Cmd) error {
	absoluteDir, err := filepath.Abs(projectDir)
	if err != nil {
		return fmt.Errorf("resolve Godot project directory: %w", err)
	}
	projectDir = absoluteDir
	presetFile := filepath.Join(projectDir, "export_presets.cfg")
	presets, err := ReadPresets(presetFile)
	if err != nil {
		return err
	}
	if executable == "" {
		executable = "godot"
	}
	if outputDir == "" {
		outputDir = "builds"
	}

	projectName := filepath.Base(projectDir)
	found := onlyPreset == ""
	for _, preset := range presets {
		if onlyPreset != "" && preset.Name != onlyPreset {
			continue
		}
		found = true
		platformDir, err := PlatformDir(preset.Platform)
		if err != nil {
			return err
		}
		filename := filepath.Base(preset.ExportPath)
		if filename == "." || filename == string(filepath.Separator) || filename == "" {
			filename = projectName
			switch platformDir {
			case "windows-x64":
				filename += ".exe"
			case "web":
				filename += ".zip"
			case "macos-arm64":
				filename += ".zip"
			}
		}
		destination := filepath.Join(outputDir, platformDir, filename)
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return fmt.Errorf("create export directory: %w", err)
		}
		args := []string{"--headless", "--path", projectDir, "--export-release", preset.Name, destination}
		cmd := runner(executable, args...)
		cmd.Dir = projectDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("Godot export preset %q failed: %w\n%s", preset.Name, err, strings.TrimSpace(string(output)))
		}
		fmt.Printf("Exported %s → %s\n", preset.Name, destination)
	}
	if !found {
		return fmt.Errorf("Godot export preset %q not found", onlyPreset)
	}
	return nil
}
