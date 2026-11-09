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
	buildType := vscodeBuildType(root)
	dbPath := filepath.Join(buildDir, "compile_commands.json")
	if err := rejectSymlink(dbPath); err != nil {
		return "", 0, err
	}
	if count, err := Validate(dbPath); err == nil {
		return buildDir, count, nil
	}

	generator := cacheGenerator(buildDir)
	cmakeArgs := []string{"-S", root, "-B", buildDir, "-DCMAKE_EXPORT_COMPILE_COMMANDS=ON"}
	if generator != "" && !supportsCompileCommands(generator) {
		if _, err := exec.LookPath("ninja"); err != nil {
			return "", 0, fmt.Errorf("build directory uses %q, which cannot generate compile_commands.json; install Ninja and rerun", generator)
		}
		buildDir = ninjaBuildDir(root)
		cmakeArgs = []string{"-S", root, "-B", buildDir, "-G", "Ninja", "-DCMAKE_BUILD_TYPE=" + buildType, "-DCMAKE_EXPORT_COMPILE_COMMANDS=ON"}
	} else if generator == "" && runtime.GOOS == "windows" {
		if _, err := exec.LookPath("ninja"); err == nil {
			buildDir = ninjaBuildDir(root)
			cmakeArgs = []string{"-S", root, "-B", buildDir, "-G", "Ninja", "-DCMAKE_EXPORT_COMPILE_COMMANDS=ON"}
		} else {
			return "", 0, fmt.Errorf("CMake's default Windows generator may be Visual Studio, which cannot generate compile_commands.json; install Ninja and rerun")
		}
	}
	dbPath = filepath.Join(buildDir, "compile_commands.json")
	if err := rejectSymlink(dbPath); err != nil {
		return "", 0, err
	}
	backup := ""
	if _, err := os.Lstat(dbPath); err == nil {
		backup = dbPath + ".bak"
		if err := copyFile(dbPath, backup); err != nil {
			return "", 0, fmt.Errorf("backing up existing compilation database: %w", err)
		}
	}

	cmd := exec.Command("cmake", cmakeArgs...)
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

func findBuildDir(root string) (string, error) {
	if configured := vscodeBuildDir(root); configured != "" {
		if _, err := os.Stat(filepath.Join(configured, "CMakeCache.txt")); os.IsNotExist(err) || cacheMatches(configured, root) {
			return configured, nil
		}
		return "", fmt.Errorf("VS Code's configured CMake build directory %q belongs to a different project; update cmake.buildDirectory before repairing", configured)
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

func vscodeBuildType(root string) string {
	data, err := os.ReadFile(filepath.Join(root, ".vscode", "settings.json"))
	if err != nil {
		return "Debug"
	}
	var settings map[string]json.RawMessage
	var buildType string
	if json.Unmarshal(data, &settings) != nil || json.Unmarshal(settings["cmake.buildType"], &buildType) != nil || buildType == "" {
		return "Debug"
	}
	return buildType
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
	var buildType string
	_ = json.Unmarshal(settings["cmake.buildType"], &buildType)
	if buildType == "" {
		buildType = "Debug"
	}
	value = strings.ReplaceAll(value, "${workspaceFolder}", root)
	value = strings.ReplaceAll(value, "${workspaceRoot}", root)
	value = strings.ReplaceAll(value, "${buildType}", buildType)
	if !filepath.IsAbs(value) {
		value = filepath.Join(root, value)
	}
	return filepath.Clean(value)
}

func ninjaBuildDir(root string) string {
	base := filepath.Join(root, "build", "fe-clangd-ninja")
	candidate := base
	for i := 2; ; i++ {
		_, cacheErr := os.Stat(filepath.Join(candidate, "CMakeCache.txt"))
		if os.IsNotExist(cacheErr) {
			if info, err := os.Stat(candidate); err != nil || !info.IsDir() {
				return candidate
			}
			if empty, err := isEmptyDir(candidate); err == nil && empty {
				return candidate
			}
		} else if cacheErr == nil && cacheMatches(candidate, root) && supportsCompileCommands(cacheGenerator(candidate)) {
			return candidate
		}
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
}

func cacheGenerator(buildDir string) string {
	data, err := os.ReadFile(filepath.Join(buildDir, "CMakeCache.txt"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "CMAKE_GENERATOR:INTERNAL=") {
			return strings.TrimPrefix(line, "CMAKE_GENERATOR:INTERNAL=")
		}
	}
	return ""
}

func supportsCompileCommands(generator string) bool {
	return strings.Contains(generator, "Ninja") || strings.Contains(generator, "Makefiles")
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
