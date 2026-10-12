// Package o3de provides read-only discovery and diagnostics for an existing O3DE setup.
// It deliberately delegates registration and lifecycle operations to O3DE's own CLI.
package o3de

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type Entry struct {
	Kind    string
	Name    string
	Path    string
	Source  string
	Version string
}

type Check struct {
	Name   string
	OK     bool
	Detail string
	Fix    string
}

type objectManifest struct {
	Name          string `json:"engine_name"`
	Version       string `json:"version"`
	EngineVersion string `json:"engine_version"`
}

func ManifestPath() string {
	if p := os.Getenv("O3DE_MANIFEST_PATH"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, ".o3de", "o3de_manifest.json")
}

// Registered returns objects from O3DE's own user manifest. It never modifies registration.
func Registered(kind string) ([]Entry, error) {
	key := map[string]string{
		"engines":   "engines",
		"projects":  "projects",
		"gems":      "external_subdirectories",
		"templates": "templates",
	}[kind]
	if key == "" {
		return nil, fmt.Errorf("unknown O3DE object type %q (use engines, projects, gems, or templates)", kind)
	}
	manifest := ManifestPath()
	if manifest == "" {
		return nil, fmt.Errorf("could not determine O3DE manifest path; set O3DE_MANIFEST_PATH")
	}
	data, err := os.ReadFile(manifest)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("O3DE manifest not found at %s; no registered objects to list", manifest)
		}
		return nil, fmt.Errorf("read O3DE manifest %s: %w", manifest, err)
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parse O3DE manifest %s: %w", manifest, err)
	}
	raw, ok := root[key]
	if !ok {
		return nil, nil
	}
	entries, err := decodeEntries(raw, kind)
	if err != nil {
		return nil, fmt.Errorf("parse %s from O3DE manifest: %w", kind, err)
	}
	sort.Slice(entries, func(i, j int) bool {
		if strings.EqualFold(entries[i].Name, entries[j].Name) {
			return entries[i].Path < entries[j].Path
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	return entries, nil
}

func decodeEntries(raw json.RawMessage, kind string) ([]Entry, error) {
	var list []string
	if err := json.Unmarshal(raw, &list); err == nil {
		entries := make([]Entry, 0, len(list))
		for _, path := range list {
			if strings.TrimSpace(path) == "" {
				continue
			}
			entries = append(entries, Entry{Kind: kind, Name: filepath.Base(filepath.Clean(path)), Path: path, Source: "O3DE manifest"})
		}
		return entries, nil
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err == nil {
		var entries []Entry
		for name, value := range object {
			path, version := decodeEntryValue(value)
			if path == "" {
				continue
			}
			entries = append(entries, Entry{Kind: kind, Name: name, Path: path, Version: version, Source: "O3DE manifest"})
		}
		return entries, nil
	}
	return nil, fmt.Errorf("expected an array or object")
}

func decodeEntryValue(raw json.RawMessage) (string, string) {
	var path string
	if json.Unmarshal(raw, &path) == nil {
		return path, ""
	}
	var value map[string]any
	if json.Unmarshal(raw, &value) != nil {
		return "", ""
	}
	for _, key := range []string{"path", "engine_path", "project_path", "gem_path", "template_path"} {
		if p, ok := value[key].(string); ok && p != "" {
			version, _ := value["engine_version"].(string)
			return p, version
		}
	}
	return "", ""
}

// FindEngine checks an explicit engine path or discovers a registered engine/current engine.
func FindEngine(explicit string) (Entry, error) {
	if explicit != "" {
		return inspectEngine(explicit, "explicit path")
	}
	cwd, err := os.Getwd()
	if err == nil {
		for dir := cwd; ; dir = filepath.Dir(dir) {
			if _, statErr := os.Stat(filepath.Join(dir, "engine.json")); statErr == nil {
				return inspectEngine(dir, "current directory")
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
		}
	}
	engines, err := Registered("engines")
	if err != nil {
		return Entry{}, err
	}
	for _, engine := range engines {
		if _, statErr := os.Stat(filepath.Join(engine.Path, "engine.json")); statErr == nil {
			return inspectEngine(engine.Path, "registered engine")
		}
	}
	return Entry{}, fmt.Errorf("no O3DE engine found; pass --engine <path> or register an engine with O3DE's official CLI")
}

func inspectEngine(root, source string) (Entry, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return Entry{}, fmt.Errorf("resolve engine path: %w", err)
	}
	data, err := os.ReadFile(filepath.Join(abs, "engine.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return Entry{}, fmt.Errorf("%s has no engine.json; this does not look like an O3DE engine root", abs)
		}
		return Entry{}, fmt.Errorf("read engine manifest: %w", err)
	}
	var manifest objectManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Entry{}, fmt.Errorf("parse %s: %w", filepath.Join(abs, "engine.json"), err)
	}
	name := manifest.Name
	if name == "" {
		name = filepath.Base(abs)
	}
	version := manifest.Version
	if version == "" {
		version = manifest.EngineVersion
	}
	return Entry{Kind: "engine", Name: name, Path: abs, Source: source, Version: version}, nil
}

// Doctor reports discoverability and basic native build prerequisites. It does not mutate setup.
func Doctor(enginePath string) []Check {
	var checks []Check
	engine, err := FindEngine(enginePath)
	if err != nil {
		checks = append(checks, Check{Name: "O3DE engine", OK: false, Detail: err.Error(), Fix: "install/build an engine, or pass --engine /path/to/o3de"})
	} else {
		detail := engine.Path
		if engine.Version != "" {
			detail += " (version " + engine.Version + ")"
		}
		checks = append(checks, Check{Name: "O3DE engine", OK: true, Detail: detail})
		cli := EngineCLI(engine.Path)
		if cli == "" {
			checks = append(checks, Check{Name: "O3DE CLI", OK: false, Detail: "scripts/o3de CLI not found", Fix: "check that this is a complete O3DE engine checkout or SDK"})
		} else {
			checks = append(checks, Check{Name: "O3DE CLI", OK: true, Detail: cli})
		}
	}
	checks = append(checks, binaryCheck("cmake", "install CMake using your platform's package manager"))
	if runtime.GOOS == "windows" {
		if p, err := exec.LookPath("cl"); err == nil {
			checks = append(checks, Check{Name: "C++ compiler", OK: true, Detail: p})
		} else if p, err := exec.LookPath("clang-cl"); err == nil {
			checks = append(checks, Check{Name: "C++ compiler", OK: true, Detail: p})
		} else {
			checks = append(checks, Check{Name: "C++ compiler", OK: false, Detail: "MSVC cl.exe or clang-cl not found in PATH", Fix: "run from a Visual Studio Developer Command Prompt or install the supported C++ toolchain"})
		}
	} else {
		found := false
		for _, compiler := range []string{"clang++", "g++", "c++"} {
			if p, err := exec.LookPath(compiler); err == nil {
				checks = append(checks, Check{Name: "C++ compiler", OK: true, Detail: p})
				found = true
				break
			}
		}
		if !found {
			checks = append(checks, Check{Name: "C++ compiler", OK: false, Detail: "clang++, g++, or c++ not found in PATH", Fix: "install a C++ compiler supported by your O3DE version"})
		}
	}
	return checks
}

func binaryCheck(name, fix string) Check {
	path, err := exec.LookPath(name)
	if err != nil {
		return Check{Name: name, OK: false, Detail: "not found in PATH", Fix: fix}
	}
	return Check{Name: name, OK: true, Detail: path}
}

func EngineCLI(engineRoot string) string {
	names := []string{"o3de.sh", "o3de.bat"}
	if runtime.GOOS == "windows" {
		names = []string{"o3de.bat", "o3de.sh"}
	}
	for _, name := range names {
		path := filepath.Join(engineRoot, "scripts", name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}
	if path, err := exec.LookPath("o3de"); err == nil {
		return path
	}
	return ""
}
