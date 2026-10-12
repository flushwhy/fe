package o3de

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

var markerByKind = map[string]string{
	"engines":   "engine.json",
	"projects":  "project.json",
	"gems":      "gem.json",
	"templates": "template.json",
}

// Discover finds registered objects and optionally scans a bounded set of local roots.
// It is read-only; it does not register or modify any O3DE objects.
func Discover(kind string, roots []string, maxDepth int, registeredOnly bool) ([]Entry, error) {
	marker, ok := markerByKind[kind]
	if !ok {
		return nil, fmt.Errorf("unknown O3DE object type %q (use engines, projects, gems, or templates)", kind)
	}
	var entries []Entry
	if kind != "gems" {
		registered, err := Registered(kind)
		if err == nil {
			for _, entry := range registered {
				if isObject(entry.Path, marker) {
					entries = append(entries, entry)
				}
			}
		} else if registeredOnly {
			return nil, err
		}
	} else {
		// O3DE's external_subdirectories list can contain more than Gems.
		// Keep only entries with an actual gem.json marker.
		registered, err := Registered("gems")
		if err == nil {
			for _, entry := range registered {
				if isObject(entry.Path, marker) {
					entries = append(entries, entry)
				}
			}
		} else if registeredOnly {
			return nil, err
		}
	}
	if registeredOnly {
		return dedupeEntries(entries), nil
	}
	if maxDepth < 0 {
		return nil, fmt.Errorf("depth must be zero or greater")
	}
	if len(roots) == 0 {
		roots = DefaultSearchRoots()
	}
	for _, root := range roots {
		abs, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		info, err := os.Stat(abs)
		if err != nil || !info.IsDir() {
			continue
		}
		rootDepth := pathDepth(abs)
		err = filepath.WalkDir(abs, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return nil // An unreadable subtree should not abort discovery elsewhere.
			}
			if path != abs && d.IsDir() && shouldSkipDirectory(d.Name()) {
				return filepath.SkipDir
			}
			depth := pathDepth(path) - rootDepth
			if depth > maxDepth {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if !d.IsDir() && d.Name() == marker {
				entry, err := inspectObject(filepath.Dir(path), kind)
				if err == nil {
					entry.Source = "local scan"
					entries = append(entries, entry)
				}
			}
			return nil
		})
		if err != nil {
			continue
		}
	}
	return dedupeEntries(entries), nil
}

func DefaultSearchRoots() []string {
	cwd, _ := os.Getwd()
	roots := []string{}
	if cwd != "" {
		roots = append(roots, cwd)
	}
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		for _, name := range []string{"O3DE", "o3de", "Projects", "projects", "Development", "development", "dev", "src", "source", "Games", "games"} {
			path := filepath.Join(home, name)
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				roots = append(roots, path)
			}
		}
	}
	if runtime.GOOS != "windows" {
		for _, path := range []string{"/opt", "/usr/local/src"} {
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				roots = append(roots, path)
			}
		}
	}
	return uniquePaths(roots)
}

func inspectObject(root, kind string) (Entry, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return Entry{}, err
	}
	marker := markerByKind[kind]
	data, err := os.ReadFile(filepath.Join(abs, marker))
	if err != nil {
		return Entry{}, err
	}
	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		return Entry{}, err
	}
	name := filepath.Base(abs)
	for _, key := range []string{"engine_name", "project_name", "gem_name", "template_name", "display_name", "name"} {
		if value, ok := fields[key].(string); ok && strings.TrimSpace(value) != "" {
			name = value
			break
		}
	}
	version, _ := fields["version"].(string)
	if version == "" {
		version, _ = fields["engine_version"].(string)
	}
	return Entry{Kind: strings.TrimSuffix(kind, "s"), Name: name, Path: abs, Version: version}, nil
}

func isObject(root, marker string) bool {
	info, err := os.Stat(filepath.Join(root, marker))
	return err == nil && !info.IsDir()
}

func shouldSkipDirectory(name string) bool {
	switch strings.ToLower(name) {
	case ".git", ".hg", ".svn", ".cache", "cache", "node_modules", "build", "builds", "bin", "obj", "cmake-build-debug", "cmake-build-release", "downloads", ".venv":
		return true
	default:
		return false
	}
}

func pathDepth(path string) int {
	clean := filepath.Clean(path)
	volume := filepath.VolumeName(clean)
	rel := strings.TrimPrefix(strings.TrimPrefix(clean, volume), string(filepath.Separator))
	if rel == "" || rel == "." {
		return 0
	}
	return len(strings.Split(rel, string(filepath.Separator)))
}

func uniquePaths(paths []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, path := range paths {
		abs, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		key := strings.ToLower(filepath.Clean(abs))
		if !seen[key] {
			seen[key] = true
			out = append(out, abs)
		}
	}
	return out
}

func dedupeEntries(entries []Entry) []Entry {
	seen := map[string]bool{}
	out := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		abs, err := filepath.Abs(entry.Path)
		if err == nil {
			entry.Path = abs
		}
		key := strings.ToLower(filepath.Clean(entry.Path))
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, entry)
	}
	sort.Slice(out, func(i, j int) bool {
		if strings.EqualFold(out[i].Name, out[j].Name) {
			return out[i].Path < out[j].Path
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out
}
