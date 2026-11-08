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
	Name            string
	Platform        string
	ExportPath      string
	DedicatedServer  bool
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
		dedicated, _ := section.Key("dedicated_server").Bool()
		presets = append(presets, Preset{
			Name: name, Platform: platform, ExportPath: section.Key("export_path").String(),
			DedicatedServer: dedicated,
		})
	}
	if len(presets) == 0 {
		return nil, fmt.Errorf("no valid presets found in %s", path)
	}
	return presets, nil
}

// FindLinuxServerPreset selects the only Linux preset explicitly configured for dedicated-server export.
func FindLinuxServerPreset(presets []Preset) (Preset, error) {
	var matches []Preset
	for _, preset := range presets {
		platform := strings.ToLower(strings.TrimSpace(preset.Platform))
		if strings.Contains(platform, "linux") && preset.DedicatedServer {
			matches = append(matches, preset)
		}
	}
	switch len(matches) {
	case 0:
		return Preset{}, fmt.Errorf("no Linux dedicated-server export preset found; create a Linux preset in Godot and enable 'Export as dedicated server'")
	case 1:
		return matches[0], nil
	default:
		names := make([]string, len(matches))
		for i, preset := range matches {
			names[i] = preset.Name
		}
		return Preset{}, fmt.Errorf("multiple Linux dedicated-server presets found (%s); choose one with --preset", strings.Join(names, ", "))
	}
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
	if !filepath.IsAbs(outputDir) {
		outputDir = filepath.Join(projectDir, outputDir)
	}

	found := onlyPreset == ""
	for _, preset := range presets {
		if onlyPreset != "" && preset.Name != onlyPreset {
			continue
		}
		found = true
		destination, err := ExpectedOutputPath(projectDir, outputDir, preset)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o750); err != nil {
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

// ExpectedOutputPath returns the path Fe will write for a configured preset.
func ExpectedOutputPath(projectDir, outputDir string, preset Preset) (string, error) {
	if outputDir == "" {
		outputDir = "builds"
	}
	platformDir, err := PlatformDir(preset.Platform)
	if err != nil {
		return "", err
	}
	filename := filepath.Base(preset.ExportPath)
	if filename == "." || filename == string(filepath.Separator) || filename == "" {
		filename = filepath.Base(projectDir)
		switch platformDir {
		case "windows-x64":
			filename += ".exe"
		case "web", "macos-arm64":
			filename += ".zip"
		}
	}
	return filepath.Join(outputDir, platformDir, filename), nil
}
