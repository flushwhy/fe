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

	"github.com/flushwhy/fe/internal/godot"
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
		if strings.TrimSpace(output) == "" {
			if optimize != "server" || selectedPreset == nil || strings.TrimSpace(selectedPreset.ExportPath) == "" {
				return errors.New("an output path is required; pass --output or configure export_path in the selected server preset")
			}
			output = selectedPreset.ExportPath
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
	Long:  "Run a headless Godot export. Use --optimize server to discover and export a Linux dedicated-server preset.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := findGodotProject(".")
		if err != nil {
			return err
		}
		preset, _ := cmd.Flags().GetString("preset")
		optimize, _ := cmd.Flags().GetString("optimize")
		if optimize != "" && optimize != "server" {
			return fmt.Errorf("unsupported optimization profile %q; currently supported: server", optimize)
		}
		output, _ := cmd.Flags().GetString("output")
		presetFile := filepath.Join(project, "export_presets.cfg")
		if _, err := os.Stat(presetFile); err != nil {
			return errors.New("export presets not found; create one in the Godot editor and save export_presets.cfg")
		}
		presets, err := godot.ReadPresets(presetFile)
		if err != nil {
			return err
		}
		var selectedPreset *godot.Preset
		if optimize == "server" {
			if preset == "" {
				selected, err := godot.FindLinuxServerPreset(presets)
				if err != nil {
					return err
				}
				preset = selected.Name
				selectedPreset = &selected
				fmt.Fprintf(cmd.OutOrStdout(), "Selected Linux dedicated-server preset: %s\n", preset)
			} else {
				var selected *godot.Preset
				for i := range presets {
					if presets[i].Name == preset {
						selected = &presets[i]
						break
					}
				}
				if selected == nil {
					return fmt.Errorf("Godot export preset %q not found in export_presets.cfg", preset)
				}
				if !strings.Contains(strings.ToLower(selected.Platform), "linux") || !selected.DedicatedServer {
					return fmt.Errorf("preset %q is not configured for Linux dedicated-server export; choose a preset with 'Export as dedicated server' enabled", preset)
				}
				selectedPreset = selected
			}
		} else {
			if strings.TrimSpace(preset) == "" {
				return errors.New("an export preset is required; pass --preset or use --optimize server")
			}
			found := false
			for i := range presets {
				if presets[i].Name == preset {
					found = true
					selectedPreset = &presets[i]
					break
				}
			}
			if !found {
				return fmt.Errorf("Godot export preset %q not found in export_presets.cfg", preset)
			}
		}
		requested, _ := cmd.Flags().GetString("version")
		engine, err := selectGodotEngine(project, requested)
		if err != nil {
			return err
		}
		if !filepath.IsAbs(output) {
			output = filepath.Join(project, output)
		}
		if err := os.MkdirAll(filepath.Dir(output), 0o750); err != nil {
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
		out := cmd.OutOrStdout()
		project, err := findGodotProject(".")
		if err != nil {
			fmt.Fprintln(out, "Project: not found (no project.godot in this directory or its parents)")
		} else {
			fmt.Fprintln(out, "Project:", project)
			fmt.Fprintln(out, "project.godot: found")
			nativeScaffold := filepath.Join(project, "native", "CMakeLists.txt")
			if _, err := os.Stat(nativeScaffold); err == nil {
				fmt.Fprintln(out, "Native extension: native/CMakeLists.txt found")
				if _, err := os.Stat(filepath.Join(project, "godot-cpp", "CMakeLists.txt")); err == nil {
					fmt.Fprintln(out, "godot-cpp: found")
				} else {
					fmt.Fprintln(out, "godot-cpp: not found (needed by the included C++ GDExtension scaffold)")
				}
			}
			if b, readErr := os.ReadFile(filepath.Join(project, "project.godot")); readErr == nil && strings.Contains(string(b), "gdnative") {
				fmt.Fprintln(out, "Legacy GDNative configuration detected: Godot 3 uses GDNative; Godot 4 uses GDExtension.")
			}
		}
		fmt.Fprintln(out, "Detected engines:")
		engines := detectGodotEngines()
		if len(engines) == 0 {
			fmt.Fprintln(out, "  none on PATH")
		}
		for _, e := range engines {
			fmt.Fprintf(out, "  %s  %s\n", e.Version, e.Path)
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
	godotExportCmd.Flags().String("preset", "", "export preset name from export_presets.cfg (auto-detected for --optimize server)")
	godotExportCmd.Flags().String("optimize", "", "apply a target optimization profile (currently: server)")
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
	return selectGodotEngineFrom(project, requested, detectGodotEngines())
}

func selectGodotEngineFrom(project, requested string, engines []godotEngine) (*godotEngine, error) {
	if len(engines) == 0 {
		return nil, errors.New("no Godot executable found on PATH; install Godot or add its directory to PATH")
	}
	if requested != "" {
		for i := range engines {
			if godotVersionMatches(engines[i].Version, requested) {
				return &engines[i], nil
			}
		}
		return nil, fmt.Errorf("no installed Godot engine matches version %q; run 'fe godot engines'", requested)
	}

	projectVersion := projectGodotVersion(project)
	if projectVersion != "" {
		for i := range engines {
			if godotVersionMatches(engines[i].Version, projectVersion) {
				return &engines[i], nil
			}
		}
		return nil, fmt.Errorf("project requires Godot %s, but no installed engine matches; run 'fe godot engines' or pass --version", projectVersion)
	}
	if len(engines) == 1 {
		return &engines[0], nil
	}
	return nil, errors.New("could not choose an installed engine; project version could not be detected, so pass --version (for example --version 4.3) or run 'fe godot engines'")
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

func projectGodotVersion(project string) string {
	data, err := os.ReadFile(filepath.Join(project, "project.godot"))
	if err != nil {
		return ""
	}
	// Godot 4 stores the major/minor version in config/features, for example
	// PackedStringArray("4.3", "GL Compatibility").
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.Contains(line, "config/features") {
			continue
		}
		for _, field := range strings.FieldsFunc(line, func(r rune) bool {
			return !(r >= '0' && r <= '9' || r == '.')
		}) {
			if strings.Contains(field, ".") && (strings.HasPrefix(field, "3.") || strings.HasPrefix(field, "4.")) {
				return field
			}
		}
	}
	return ""
}

func godotVersionMatches(installed, requested string) bool {
	requested = strings.TrimSpace(requested)
	return requested != "" && (installed == requested || strings.HasPrefix(installed, requested+"."))
}
