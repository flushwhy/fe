package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

type godotEngine struct {
	Path    string
	Version string
}

var godotCmd = &cobra.Command{
	Use:   "godot",
	Short: "Discover Godot engines and manage native extensions.",
	Long:  "Convenience commands for Godot 3 GDNative and Godot 4 GDExtension projects.",
}

var godotEnginesCmd = &cobra.Command{
	Use:   "engines",
	Short: "List detected Godot installations and their versions.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		engines := detectGodotEngines()
		if len(engines) == 0 {
			return errors.New("no Godot executable found on PATH; install Godot or add its directory to PATH")
		}
		for _, e := range engines {
			fmt.Printf("%-18s %s\n", e.Version, e.Path)
		}
		return nil
	},
}

var godotOpenCmd = &cobra.Command{
	Use:   "open",
	Short: "Open the current Godot project with a selected or matching engine.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := findGodotProject(".")
		if err != nil {
			return err
		}
		requested, _ := cmd.Flags().GetString("version")
		chosen, err := selectGodotEngine(project, requested)
		if err != nil {
			return err
		}
		c := exec.Command(chosen.Path, "--editor", "--path", project)
		c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
		return c.Run()
	},
}

var godotExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export the current Godot project using an export preset.",
	Long:  "Run a headless Godot export. Create export presets in the Godot editor first.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := findGodotProject(".")
		if err != nil {
			return err
		}
		preset, _ := cmd.Flags().GetString("preset")
		if strings.TrimSpace(preset) == "" {
			return errors.New("an export preset is required; pass --preset")
		}
		output, _ := cmd.Flags().GetString("output")
		if strings.TrimSpace(output) == "" {
			return errors.New("an output path is required; pass --output")
		}
		if _, err := os.Stat(filepath.Join(project, "export_presets.cfg")); err != nil {
			return errors.New("export presets not found; create one in the Godot editor and save export_presets.cfg")
		}
		requested, _ := cmd.Flags().GetString("version")
		engine, err := selectGodotEngine(project, requested)
		if err != nil {
			return err
		}
		if !filepath.IsAbs(output) {
			output = filepath.Join(project, output)
		}
		if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
			return fmt.Errorf("create export output directory: %w", err)
		}
		mode := "--export-release"
		if debug, _ := cmd.Flags().GetBool("debug"); debug {
			mode = "--export-debug"
		}
		c := exec.Command(engine.Path, "--headless", "--path", project, mode, preset, output)
		c.Stdin, c.Stdout, c.Stderr = cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr()
		return c.Run()
	},
}

var godotDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check the current Godot project and native extension setup.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := findGodotProject(".")
		if err != nil {
			fmt.Println("Project: not found (no project.godot in this directory or its parents)")
		} else {
			fmt.Println("Project:", project)
			if _, err := os.Stat(filepath.Join(project, "project.godot")); err == nil {
				fmt.Println("project.godot: found")
			}
			if _, err := os.Stat(filepath.Join(project, "native", "CMakeLists.txt")); err == nil {
				fmt.Println("Native extension: native/CMakeLists.txt found")
			}
			if _, err := os.Stat(filepath.Join(project, "godot-cpp")); err == nil {
				fmt.Println("godot-cpp: found")
			} else {
				fmt.Println("godot-cpp: not found (needed by the included C++ GDExtension scaffold)")
			}
			if _, err := os.Stat(filepath.Join(project, "project.godot")); err == nil {
				if b, readErr := os.ReadFile(filepath.Join(project, "project.godot")); readErr == nil && strings.Contains(string(b), "gdnative") {
					fmt.Println("Legacy GDNative configuration detected: Godot 3 uses GDNative; Godot 4 uses GDExtension.")
				}
			}
		}
		fmt.Println("Detected engines:")
		engines := detectGodotEngines()
		if len(engines) == 0 {
			fmt.Println("  none on PATH")
		}
		for _, e := range engines {
			fmt.Printf("  %s  %s\n", e.Version, e.Path)
		}
		return nil
	},
}

var godotExtensionCmd = &cobra.Command{
	Use:     "extension",
	Aliases: []string{"ex", "ext"},
	Short:   "Scaffold, build, and diagnose a native Godot extension (aliases: ex, ext).",
	Long: `Manage a Godot 4 C++ GDExtension scaffold. Godot 3 GDNative is a separate,
legacy API and is detected by doctor but is not silently converted.

Examples:
  fe godot ex init
  fe godot ext setup
  fe godot ex build
  fe godot ex doctor`,
}

var godotExtensionInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a minimal C++ GDExtension scaffold in ./native.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := findGodotProject(".")
		if err != nil {
			return err
		}
		dir := filepath.Join(project, "native")
		if _, err := os.Stat(dir); err == nil {
			return fmt.Errorf("%s already exists; refusing to overwrite it", dir)
		} else if !os.IsNotExist(err) {
			return err
		}
		files := map[string]string{
			"CMakeLists.txt": `cmake_minimum_required(VERSION 3.20)
project(fe_godot_extension LANGUAGES CXX)
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
if(NOT EXISTS "${CMAKE_CURRENT_SOURCE_DIR}/../godot-cpp/CMakeLists.txt")
  message(FATAL_ERROR "godot-cpp missing. Run: fe godot ex setup")
endif()
add_subdirectory("${CMAKE_CURRENT_SOURCE_DIR}/../godot-cpp" godot-cpp)
add_library(fe_godot_extension SHARED src/register_types.cpp)
target_link_libraries(fe_godot_extension PRIVATE godot-cpp)
set_target_properties(fe_godot_extension PROPERTIES
  LIBRARY_OUTPUT_DIRECTORY "${CMAKE_CURRENT_SOURCE_DIR}/../bin"
  LIBRARY_OUTPUT_DIRECTORY_DEBUG "${CMAKE_CURRENT_SOURCE_DIR}/../bin"
  LIBRARY_OUTPUT_DIRECTORY_RELEASE "${CMAKE_CURRENT_SOURCE_DIR}/../bin"
  RUNTIME_OUTPUT_DIRECTORY "${CMAKE_CURRENT_SOURCE_DIR}/../bin"
  RUNTIME_OUTPUT_DIRECTORY_DEBUG "${CMAKE_CURRENT_SOURCE_DIR}/../bin"
  RUNTIME_OUTPUT_DIRECTORY_RELEASE "${CMAKE_CURRENT_SOURCE_DIR}/../bin")
if(WIN32)
  set_target_properties(fe_godot_extension PROPERTIES PREFIX "" SUFFIX ".dll")
elseif(APPLE)
  set_target_properties(fe_godot_extension PROPERTIES PREFIX "lib" SUFFIX ".dylib")
else()
  set_target_properties(fe_godot_extension PROPERTIES PREFIX "lib" SUFFIX ".so")
endif()
`,
			"../fe_godot_extension.gdextension": `[configuration]
entry_symbol = "fe_extension_init"
compatibility_minimum = "4.3"
reloadable = true

[libraries]
linux.debug.x86_64 = "res://bin/libfe_godot_extension.so"
linux.release.x86_64 = "res://bin/libfe_godot_extension.so"
windows.debug.x86_64 = "res://bin/fe_godot_extension.dll"
windows.release.x86_64 = "res://bin/fe_godot_extension.dll"
macos.debug = "res://bin/libfe_godot_extension.dylib"
macos.release = "res://bin/libfe_godot_extension.dylib"
`,
			"src/register_types.cpp": `#include <godot_cpp/core/class_db.hpp>
#include <godot_cpp/godot.hpp>

using namespace godot;

void initialize_fe_extension(ModuleInitializationLevel level) {
    if (level != MODULE_INITIALIZATION_LEVEL_SCENE) return;
}

void uninitialize_fe_extension(ModuleInitializationLevel level) {
    if (level != MODULE_INITIALIZATION_LEVEL_SCENE) return;
}

extern "C" {
GDExtensionBool GDE_EXPORT fe_extension_init(
    GDExtensionInterfaceGetProcAddress get_proc_address,
    GDExtensionClassLibraryPtr library,
    GDExtensionInitialization *initialization) {
    GDExtensionBinding::InitObject init_obj(get_proc_address, library, initialization);
    init_obj.register_initializer(initialize_fe_extension);
    init_obj.register_terminator(uninitialize_fe_extension);
    init_obj.set_minimum_library_initialization_level(MODULE_INITIALIZATION_LEVEL_SCENE);
    return init_obj.init();
}
}
`,
			"README.md": `# Native Godot extension

This scaffold targets Godot 4 GDExtension and C++ via godot-cpp.

1. Run "fe godot ex setup" to fetch godot-cpp.
2. Run "fe godot ex build" to configure and compile.
3. The generated fe_godot_extension.gdextension manifest is ready to use for x86_64 Linux and Windows, and macOS builds.

Godot 3 GDNative uses a different API and is not compatible with this scaffold.
`,
		}
		for name, data := range files {
			path := filepath.Join(dir, filepath.FromSlash(name))
			if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
				return err
			}
			if err := os.WriteFile(path, []byte(data), 0o640); err != nil {
				return fmt.Errorf("write %s: %w", path, err)
			}
		}
		fmt.Println("Created native GDExtension scaffold in", dir)
		fmt.Println("Next: fe godot ex setup")
		return nil
	},
}

var godotExtensionSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Fetch the godot-cpp dependency for the native scaffold.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := findGodotProject(".")
		if err != nil {
			return err
		}
		native := filepath.Join(project, "native")
		if _, err := os.Stat(filepath.Join(native, "CMakeLists.txt")); err != nil {
			return fmt.Errorf("native scaffold not found; run 'fe godot ex init' first")
		}
		if _, err := os.Stat(filepath.Join(project, "godot-cpp", "CMakeLists.txt")); err == nil {
			fmt.Println("godot-cpp is already present")
			return nil
		}
		if _, err := exec.LookPath("git"); err != nil {
			return errors.New("git is required to fetch godot-cpp")
		}
		c := exec.Command("git", "clone", "--depth", "1", "--branch", "4.3", "https://github.com/godotengine/godot-cpp.git", filepath.Join(project, "godot-cpp"))
		c.Stdin, c.Stdout, c.Stderr = cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr()
		return c.Run()
	},
}

var godotExtensionBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the native GDExtension using CMake.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := findGodotProject(".")
		if err != nil {
			return err
		}
		native := filepath.Join(project, "native")
		if _, err := os.Stat(filepath.Join(native, "CMakeLists.txt")); err != nil {
			return fmt.Errorf("native scaffold not found; run 'fe godot ex init'")
		}
		if _, err := os.Stat(filepath.Join(project, "godot-cpp", "CMakeLists.txt")); err != nil {
			return fmt.Errorf("godot-cpp not found; run 'fe godot ex setup'")
		}
		if _, err := exec.LookPath("cmake"); err != nil {
			return errors.New("cmake is required to build the extension")
		}
		buildDir := filepath.Join(native, "build")
		configure := exec.Command("cmake", "-S", native, "-B", buildDir)
		configure.Stdin, configure.Stdout, configure.Stderr = cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr()
		if err := configure.Run(); err != nil {
			return fmt.Errorf("configure extension: %w", err)
		}
		build := exec.Command("cmake", "--build", buildDir, "--config", "Debug")
		build.Stdin, build.Stdout, build.Stderr = cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr()
		return build.Run()
	},
}

var godotExtensionDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check native extension build prerequisites.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := findGodotProject(".")
		if err != nil {
			return err
		}
		checks := []struct{ name, path string }{
			{"native/CMakeLists.txt", filepath.Join(project, "native", "CMakeLists.txt")},
			{"godot-cpp/CMakeLists.txt", filepath.Join(project, "godot-cpp", "CMakeLists.txt")},
		}
		failed := false
		for _, check := range checks {
			if _, err := os.Stat(check.path); err == nil {
				fmt.Println("OK:", check.name)
			} else {
				fmt.Println("MISSING:", check.name)
				failed = true
			}
		}
		for _, tool := range []string{"cmake", "git"} {
			if path, err := exec.LookPath(tool); err == nil {
				fmt.Printf("OK: %s (%s)\n", tool, path)
			} else {
				fmt.Println("MISSING:", tool)
				failed = true
			}
		}
		if failed {
			return errors.New("native extension prerequisites are incomplete")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(godotCmd)
	godotCmd.AddCommand(godotEnginesCmd, godotOpenCmd, godotExportCmd, godotDoctorCmd, godotExtensionCmd)
	godotOpenCmd.Flags().String("version", "", "installed Godot version to use (for example 4.3 or 3.5)")
	godotExportCmd.Flags().String("preset", "", "export preset name from export_presets.cfg")
	godotExportCmd.Flags().String("output", "", "export output path (relative to the project or absolute)")
	godotExportCmd.Flags().Bool("debug", false, "create a debug export instead of a release export")
	godotExportCmd.Flags().String("version", "", "installed Godot version to use (for example 4.3 or 3.5)")
	godotExtensionCmd.AddCommand(godotExtensionInitCmd, godotExtensionSetupCmd, godotExtensionBuildCmd, godotExtensionDoctorCmd)
}

func detectGodotEngines() []godotEngine {
	names := []string{"godot", "godot4", "godot-mono", "godot4-mono", "Godot"}
	if runtime.GOOS == "windows" {
		names = []string{"godot.exe", "godot4.exe", "Godot.exe", "godot-mono.exe"}
	}
	seen := map[string]bool{}
	var engines []godotEngine
	for _, name := range names {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		path, _ = filepath.Abs(path)
		if seen[path] {
			continue
		}
		seen[path] = true
		out, err := exec.Command(path, "--version").Output()
		version := "unknown"
		if err == nil {
			version = strings.TrimSpace(string(out))
			scanner := bufio.NewScanner(strings.NewReader(version))
			if scanner.Scan() {
				version = scanner.Text()
			}
		}
		engines = append(engines, godotEngine{Path: path, Version: version})
	}
	return engines
}

func selectGodotEngine(project, requested string) (*godotEngine, error) {
	engines := detectGodotEngines()
	for i := range engines {
		if requested != "" && !strings.Contains(engines[i].Version, requested) {
			continue
		}
		if requested == "" && !projectVersionMatches(project, engines[i].Version) {
			continue
		}
		return &engines[i], nil
	}
	if requested == "" && len(engines) == 1 {
		return &engines[0], nil
	}
	if requested == "" {
		return nil, fmt.Errorf("could not choose an installed engine matching this project; use --version (for example --version 4.3) or run 'fe godot engines'")
	}
	return nil, fmt.Errorf("no installed Godot engine matches version %q; run 'fe godot engines'", requested)
}

func findGodotProject(start string) (string, error) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for dir := abs; ; dir = filepath.Dir(dir) {
		if info, err := os.Stat(filepath.Join(dir, "project.godot")); err == nil && !info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return "", fmt.Errorf("no project.godot found from %q upward", abs)
}

func projectVersionMatches(project, version string) bool {
	data, err := os.ReadFile(filepath.Join(project, "project.godot"))
	if err != nil {
		return false
	}
	// Godot project files commonly record config/features = PackedStringArray("4.3", ...).
	var projectVersion string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, "config/features") || strings.Contains(line, "config_version") {
			for _, field := range strings.FieldsFunc(line, func(r rune) bool {
				return !(r >= '0' && r <= '9' || r == '.')
			}) {
				if strings.Contains(field, ".") && field[0] >= '3' && field[0] <= '4' {
					projectVersion = field
					break
				}
			}
		}
	}
	if projectVersion == "" {
		return false
	}
	// Engine --version strings often look like 4.3.stable.official.
	return strings.HasPrefix(version, projectVersion)
}

