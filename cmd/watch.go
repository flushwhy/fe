package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch a workspace and automate its development loop.",
	Long: `Detects an O3DE project or Gem from its manifests and watches source files.
C++ changes trigger an incremental CMake build. Lua and asset changes are
reported for O3DE's existing Asset Processor/reload pipeline; Fe does not
restart the editor or claim to hot-reload native C++ modules.

Run from a project or Gem directory:
  fe watch
  fe watch --build-dir build/linux
  fe watch --target MyGem --target MyGem.Editor`,
	Args: cobra.NoArgs,
	RunE: runWatch,
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().String("path", ".", "workspace directory to detect and watch")
	watchCmd.Flags().String("build-dir", "", "configured CMake build directory (auto-detected when possible)")
	watchCmd.Flags().StringArray("target", nil, "CMake target to build for C++ changes (repeatable)")
	watchCmd.Flags().Duration("debounce", 500*time.Millisecond, "quiet period after file changes before rebuilding")
}

func runWatch(cmd *cobra.Command, _ []string) error {
	path, _ := cmd.Flags().GetString("path")
	root, kind, err := detectWatchWorkspace(path)
	if err != nil {
		return err
	}
	buildDir, _ := cmd.Flags().GetString("build-dir")
	if buildDir != "" {
		if !filepath.IsAbs(buildDir) {
			buildDir = filepath.Join(root, buildDir)
		}
	} else {
		buildDir = findWatchBuildDir(root)
	}
	if buildDir != "" {
		if _, err := os.Stat(filepath.Join(buildDir, "CMakeCache.txt")); err != nil {
			return fmt.Errorf("build directory %q is not a configured CMake build (CMakeCache.txt missing); pass --build-dir", buildDir)
		}
	}
	targets, _ := cmd.Flags().GetStringArray("target")
	debounce, _ := cmd.Flags().GetDuration("debounce")
	if debounce < 0 {
		return fmt.Errorf("--debounce cannot be negative")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create file watcher: %w", err)
	}
	defer watcher.Close()
	if err := addWatchTree(watcher, root); err != nil {
		return err
	}

	fmt.Printf("Watching %s workspace: %s\n", kind, root)
	if buildDir != "" {
		fmt.Printf("C++ build directory: %s\n", buildDir)
	} else {
		fmt.Println("No configured CMake build found; C++ changes will be reported but not built. Pass --build-dir to enable builds.")
	}
	fmt.Println("Press Ctrl-C to stop.")
	var pending []string
	var timer *time.Timer
	var timerC <-chan time.Time
	building := false

	for {
		select {
		case <-cmd.Context().Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op& (fsnotify.Create|fsnotify.Write|fsnotify.Rename|fsnotify.Remove) == 0 {
				continue
			}
			if info, statErr := os.Stat(event.Name); statErr == nil && info.IsDir() && event.Op&fsnotify.Create != 0 {
				if err := addWatchTree(watcher, event.Name); err != nil {
					fmt.Fprintf(os.Stderr, "watch directory: %v\n", err)
				}
				continue
			}
			category := classifyWatchFile(event.Name)
			if category == "" {
				continue
			}
			pending = append(pending, event.Name)
			if timer == nil {
				timer = time.NewTimer(debounce)
			} else {
				if !timer.Stop() {
					select { case <-timer.C: default: }
				}
				timer.Reset(debounce)
			}
			timerC = timer.C
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "watcher: %v\n", err)
		case <-timerC:
			timerC = nil
			changed := uniqueStrings(pending)
			pending = nil
			hasCPP, hasLua, hasAsset := false, false, false
			for _, file := range changed {
				switch classifyWatchFile(file) {
				case "cpp": hasCPP = true
				case "lua": hasLua = true
				case "asset": hasAsset = true
				}
				fmt.Printf("Changed: %s\n", file)
			}
			if hasLua {
				fmt.Println("Lua changed: O3DE's running Asset Processor/runtime pipeline may pick this up; Fe does not restart the editor.")
			}
			if hasAsset {
				fmt.Println("Asset changed: O3DE's Asset Processor handles asset processing and supported live reload.")
			}
			if hasCPP {
				if buildDir == "" {
					fmt.Println("C++ changed; configure a CMake build and pass --build-dir to enable automatic builds.")
				} else if !building {
					building = true
					if err := runWatchBuild(cmd.Context(), buildDir, targets); err != nil {
						fmt.Fprintf(os.Stderr, "Build failed: %v\n", err)
					} else {
						fmt.Println("Build succeeded. Native C++ changes may still require an editor or game restart to load.")
					}
					building = false
				}
			}
		}
	}
}

func detectWatchWorkspace(path string) (string, string, error) {
	abs, err := filepath.Abs(path)
	if err != nil { return "", "", fmt.Errorf("resolve workspace: %w", err) }
	info, err := os.Stat(abs)
	if err != nil { return "", "", fmt.Errorf("inspect workspace %q: %w", abs, err) }
	if !info.IsDir() { abs = filepath.Dir(abs) }
	for dir := abs; ; dir = filepath.Dir(dir) {
		if fileExists(filepath.Join(dir, "project.json")) { return dir, "O3DE project", nil }
		if fileExists(filepath.Join(dir, "gem.json")) { return dir, "O3DE Gem", nil }
		if fileExists(filepath.Join(dir, "engine.json")) { return dir, "O3DE engine", nil }
		parent := filepath.Dir(dir)
		if parent == dir { break }
	}
	return "", "", fmt.Errorf("could not detect an O3DE project, Gem, or engine from %q (expected project.json, gem.json, or engine.json in this directory or a parent)", abs)
}

func findWatchBuildDir(root string) string {
	candidates := []string{
		filepath.Join(root, "build"),
		filepath.Join(root, "build", "windows"),
		filepath.Join(root, "build", "linux"),
		filepath.Join(root, "build", "mac"),
		filepath.Join(root, "build", "macos"),
	}
	for _, candidate := range candidates {
		if fileExists(filepath.Join(candidate, "CMakeCache.txt")) { return candidate }
	}
	return ""
}

func addWatchTree(watcher *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil { return nil }
		if entry.IsDir() && path != root && ignoredWatchDir(entry.Name()) { return filepath.SkipDir }
		if entry.IsDir() {
			if err := watcher.Add(path); err != nil { return fmt.Errorf("watch %q: %w", path, err) }
		}
		return nil
	})
}

func ignoredWatchDir(name string) bool {
	switch strings.ToLower(name) {
	case ".git", ".fe", "build", "cache", "user", "bin", "node_modules", "thirdparty", "3rdparty":
		return true
	default:
		return strings.HasPrefix(name, ".")
	}
}

func classifyWatchFile(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".cpp", ".cc", ".cxx", ".c", ".h", ".hh", ".hpp", ".hxx", ".inl", ".cmake":
		return "cpp"
	case ".lua":
		return "lua"
	case ".azasset", ".assetinfo", ".prefab", ".slice", ".material", ".mtl", ".shader", ".shaderoption", ".png", ".jpg", ".jpeg", ".tif", ".tiff", ".bmp", ".dds", ".fbx", ".obj", ".gltf", ".glb", ".wav", ".ogg":
		return "asset"
	default:
		if strings.EqualFold(filepath.Base(path), "CMakeLists.txt") { return "cpp" }
		return ""
	}
}

func runWatchBuild(ctx context.Context, buildDir string, targets []string) error {
	args := []string{"--build", buildDir, "--parallel"}
	if len(targets) > 0 {
		args = append(args, "--target")
		args = append(args, targets...)
	}
	fmt.Printf("Building: cmake %s\n", strings.Join(args, " "))
	build := exec.CommandContext(ctx, "cmake", args...)
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil { return fmt.Errorf("cmake --build: %w", err) }
	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if !seen[value] { seen[value] = true; result = append(result, value) }
	}
	return result
}
