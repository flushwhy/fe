// Package compdb repairs C/C++ compilation databases using the project's build system.
package compdb

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Entry struct {
	Directory string   `json:"directory"`
	File      string   `json:"file"`
	Command   string   `json:"command"`
	Arguments []string `json:"arguments"`
}

// Validate checks the compilation database structure and paths clangd relies on.
func Validate(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return 0, fmt.Errorf("invalid JSON: %w", err)
	}
	if len(entries) == 0 {
		return 0, fmt.Errorf("database is empty")
	}
	for i, entry := range entries {
		if entry.Directory == "" || entry.File == "" || (entry.Command == "" && len(entry.Arguments) == 0) {
			return 0, fmt.Errorf("entry %d must contain directory, file, and command or arguments", i)
		}
		dir := entry.Directory
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(filepath.Dir(path), dir)
		}
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			return 0, fmt.Errorf("entry %d has missing working directory %q", i, entry.Directory)
		}
		source := entry.File
		if !filepath.IsAbs(source) {
			source = filepath.Join(dir, source)
		}
		if _, err := os.Stat(source); err != nil {
			return 0, fmt.Errorf("entry %d references missing source file %q", i, entry.File)
		}
	}
	return len(entries), nil
}

// Repair configures CMake to emit compile_commands.json. It never guesses compiler flags.
func Repair(root string) (string, int, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return "", 0, err
	}
	if _, err := os.Stat(filepath.Join(root, "CMakeLists.txt")); err != nil {
		return "", 0, fmt.Errorf("no CMakeLists.txt found in %s; for Make projects, use Bear with your actual build command", root)
	}

	buildDir, err := findBuildDir(root)
	if err != nil {
		return "", 0, err
	}
	dbPath := filepath.Join(buildDir, "compile_commands.json")
	if err := rejectSymlink(dbPath); err != nil {
		return "", 0, err
	}

	if _, err := Validate(dbPath); err == nil {
		return buildDir, mustCount(dbPath), nil
	}

	backup := ""
	if _, err := os.Lstat(dbPath); err == nil {
		backup = dbPath + ".bak"
		if err := copyFile(dbPath, backup); err != nil {
			return "", 0, fmt.Errorf("backing up existing compilation database: %w", err)
		}
	}

	cmd := exec.Command("cmake", "-S", root, "-B", buildDir, "-DCMAKE_EXPORT_COMPILE_COMMANDS=ON")
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		if backup != "" {
			_ = copyFile(backup, dbPath)
		}
		return "", 0, fmt.Errorf("CMake configure failed: %w\n%s", runErr, strings.TrimSpace(string(output)))
	}
	count, err := Validate(dbPath)
	if err != nil {
		if backup != "" {
			_ = copyFile(backup, dbPath)
		}
		return "", 0, fmt.Errorf("CMake completed but produced an unusable compilation database: %w", err)
	}
	return buildDir, count, nil
}

func mustCount(path string) int {
	count, _ := Validate(path)
	return count
}

func findBuildDir(root string) (string, error) {
	if configured := vscodeBuildDir(root); configured != "" {
		if cacheMatches(configured, root) {
			return configured, nil
		}
		if _, err := os.Stat(configured); os.IsNotExist(err) {
			return configured, nil
		}
		if empty, _ := isEmptyDir(configured); empty {
			return configured, nil
		}
	}

	var matches []string
	known := []string{"build", "cmake-build-debug", "cmake-build-release", "out/build"}
	for _, name := range known {
		dir := filepath.Join(root, name)
		if cacheMatches(dir, root) {
			return dir, nil
		}
	}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != root && (d.Name() == ".git" || d.Name() == "node_modules" || d.Name() == ".cache" || d.Name() == "vendor") {
				return filepath.SkipDir
			}
			if path != root && depth(root, path) > 4 {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == "CMakeCache.txt" && cacheMatches(filepath.Dir(path), root) {
			matches = append(matches, filepath.Dir(path))
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("found multiple CMake build directories (%s); set cmake.buildDirectory in .vscode/settings.json or remove stale build caches", strings.Join(matches, ", "))
	}

	buildDir := filepath.Join(root, "build")
	if info, err := os.Stat(buildDir); err == nil && info.IsDir() {
		if empty, _ := isEmptyDir(buildDir); !empty {
			buildDir = filepath.Join(buildDir, "fe-clangd")
		}
	}
	return buildDir, nil
}

func vscodeBuildDir(root string) string {
	data, err := os.ReadFile(filepath.Join(root, ".vscode", "settings.json"))
	if err != nil {
		return ""
	}
	var settings map[string]json.RawMessage
	if json.Unmarshal(data, &settings) != nil {
		return ""
	}
	var value string
	if json.Unmarshal(settings["cmake.buildDirectory"], &value) != nil || value == "" {
		return ""
	}
	value = strings.ReplaceAll(value, "${workspaceFolder}", root)
	value = strings.ReplaceAll(value, "${workspaceRoot}", root)
	if !filepath.IsAbs(value) {
		value = filepath.Join(root, value)
	}
	return filepath.Clean(value)
}

func cacheMatches(buildDir, root string) bool {
	data, err := os.ReadFile(filepath.Join(buildDir, "CMakeCache.txt"))
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "CMAKE_HOME_DIRECTORY:INTERNAL=") {
			source := strings.TrimPrefix(line, "CMAKE_HOME_DIRECTORY:INTERNAL=")
			a, _ := filepath.Abs(source)
			b, _ := filepath.Abs(root)
			return samePath(a, b)
		}
	}
	return false
}

func samePath(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

func depth(root, path string) int {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return 99
	}
	return len(strings.Split(rel, string(os.PathSeparator)))
}

func isEmptyDir(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

func rejectSymlink(path string) error {
	info, err := os.Lstat(path)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to overwrite symlink %s", path)
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := rejectSymlink(dst); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0600)
}
