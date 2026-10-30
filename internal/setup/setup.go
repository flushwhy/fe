package setup

import (
	"os"
	"os/exec"
	"path/filepath"

	projectinspect "github.com/flushwhy/fe/internal/inspect"
)

type Tool struct {
	Name     string
	Required bool
	Reason   string
	Found    string
}

func Inspect(root string) ([]Tool, error) {
	report, err := projectinspect.Inspect(root)
	if err != nil {
		return nil, err
	}

	var tools []Tool
	add := func(name string, required bool, reason string, alternatives ...string) {
		tool := Tool{Name: name, Required: required, Reason: reason}
		for _, candidate := range alternatives {
			if path, err := exec.LookPath(candidate); err == nil {
				tool.Found = path
				break
			}
		}
		tools = append(tools, tool)
	}

	for _, project := range report.Projects {
		switch project.Type {
		case "Go":
			add("Go", true, "build and test the Go project", "go")
		case "Rust":
			add("Rust toolchain", true, "build and test the Cargo project", "cargo")
		case "Python":
			add("Python", true, "run Python checks and scripts", "python3", "python")
		case "Node.js":
			add("Node.js", true, "run project scripts", "node")
			if _, err := os.Stat(filepath.Join(report.Root, "pnpm-lock.yaml")); err == nil {
				add("pnpm", true, "install dependencies using the project's lockfile", "pnpm")
			} else if _, err := os.Stat(filepath.Join(report.Root, "yarn.lock")); err == nil {
				add("Yarn", true, "install dependencies using the project's lockfile", "yarn")
			} else {
				add("npm", true, "install dependencies and run project scripts", "npm")
			}
		case "C/C++":
			add("C/C++ compiler", true, "compile the project", "cc", "gcc", "clang", "c++", "g++", "clang++")
			if _, err := os.Stat(filepath.Join(report.Root, "CMakeLists.txt")); err == nil {
				add("CMake", true, "configure and build this CMake project", "cmake")
			}
			if _, err := os.Stat(filepath.Join(report.Root, "Makefile")); err == nil {
				add("Make", true, "run the project's Makefile targets", "make", "gmake")
			}
		case "Godot":
			add("Godot", true, "run the project and perform engine exports", "godot", "godot4")
		case "Blender add-on":
			add("Blender", false, "validate and build the extension with Blender's extension commands", "blender")
		case "Lua":
			add("Lua", true, "run Lua scripts and checks", "lua")
		}
	}
	return tools, nil
}
