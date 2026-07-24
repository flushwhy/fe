// Package doctor checks that required tools are installed and configs are sane.
package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/flushwhy/fe/internal/ui"
)

type Result struct {
	Name    string
	OK      bool
	Message string
	Fix     string
}

func Run() []Result {
	cwd, err := os.Getwd()
	if err != nil {
		return []Result{{Name: "working directory", OK: false, Message: err.Error()}}
	}

	var results []Result

	_ = ui.RunSpinner("Detecting project type", func() error {
		time.Sleep(150 * time.Millisecond) // let spinner render
		return nil
	})

	lang := detectLang(cwd)
	results = append(results, Result{
		Name:    "project type",
		OK:      lang != "",
		Message: describeLang(lang, cwd),
	})

	_ = ui.RunSpinner("Checking common tools", func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	results = append(results, checkBinary("git", ""))
	results = append(results, checkBinary("make", "brew install make  /  apt install make"))

	switch lang {
	case "c", "cpp":
		_ = ui.RunSpinner("Checking C/C++ toolchain", func() error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
		results = append(results, checkBinary("clangd", "brew install llvm  /  apt install clangd"))
		results = append(results, checkBinary("bear", "brew install bear  /  apt install bear"))
		results = append(results, checkClangdConfig(cwd))
		results = append(results, checkCompileCommands(cwd))
		if lang == "cpp" {
			results = append(results, checkBinary("cmake", "brew install cmake  /  apt install cmake"))
			results = append(results, checkCMakeExportFlag(cwd))
		}
	case "odin":
		_ = ui.RunSpinner("Checking Odin toolchain", func() error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
		results = append(results, checkBinary("odin", "https://odin-lang.org/docs/install/"))
		results = append(results, checkBinary("ols", "https://github.com/DanielGavin/ols"))
		results = append(results, checkOLSConfig(cwd))
	case "zig":
		_ = ui.RunSpinner("Checking Zig toolchain", func() error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
		results = append(results, checkBinary("zig", "https://ziglang.org/download/"))
		results = append(results, checkBinary("zls", "https://github.com/zigtools/zls"))
		results = append(results, checkZLSConfig(cwd))
	case "go":
		_ = ui.RunSpinner("Checking Go toolchain", func() error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
		results = append(results, checkBinary("go", "https://go.dev/dl/"))
		results = append(results, checkBinary("gopls", "go install golang.org/x/tools/gopls@latest"))
		results = append(results, checkGoMod(cwd))
	}

	return results
}

func Print(results []Result) bool {
	allOK := true
	fmt.Println()

	for _, r := range results {
		if r.OK {
			fmt.Printf("  %s  %-32s %s\n",
				ui.StyleSuccess.Render(ui.IconOK),
				ui.StyleBold.Render(r.Name),
				ui.StyleDim.Render(r.Message),
			)
		} else {
			allOK = false
			fmt.Printf("  %s  %-32s %s\n",
				ui.StyleError.Render(ui.IconFail),
				ui.StyleBold.Render(r.Name),
				ui.StyleError.Render(r.Message),
			)
			if r.Fix != "" {
				fmt.Printf("     %s %s\n",
					ui.StyleDim.Render("fix:"),
					ui.StyleDim.Render(r.Fix),
				)
			}
		}
	}

	fmt.Println()
	if allOK {
		ui.OK("All checks passed")
	} else {
		fails := 0
		for _, r := range results {
			if !r.OK {
				fails++
			}
		}
		ui.Warn("%d check(s) failed — run 'fe add <tool>' to fix most issues", fails)
	}

	return allOK
}

func detectLang(root string) string {
	checks := []struct {
		file string
		lang string
	}{
		{"build.zig", "zig"},
		{"go.mod", "go"},
		{"CMakeLists.txt", "cpp"},
		{"*.cpp", "cpp"},
		{"src/*.cpp", "cpp"},
		{"Makefile", "c"},
		{"*.c", "c"},
		{"src/*.c", "c"},
		{"*.odin", "odin"},
		{"src/*.odin", "odin"},
	}
	for _, c := range checks {
		matches, _ := filepath.Glob(filepath.Join(root, c.file))
		if len(matches) > 0 {
			return c.lang
		}
	}
	return ""
}

func describeLang(lang, root string) string {
	if lang == "" {
		return "unknown — no recognised source files found"
	}
	return fmt.Sprintf("%s project at %s", lang, filepath.Base(root))
}

func checkBinary(name, installHint string) Result {
	path, err := exec.LookPath(name)
	if err != nil {
		fix := installHint
		if fix == "" {
			fix = fmt.Sprintf("install %s and ensure it is in PATH", name)
		}
		return Result{Name: name, OK: false, Message: "not found in PATH", Fix: fix}
	}
	return Result{Name: name, OK: true, Message: path}
}

func checkClangdConfig(root string) Result {
	name := ".clangd config"
	path := filepath.Join(root, ".clangd")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return Result{Name: name, OK: false, Message: ".clangd not found", Fix: "run: fe add clangd"}
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return Result{Name: name, OK: false, Message: err.Error()}
	}
	if !strings.Contains(string(content), "CompileFlags") {
		return Result{
			Name:    name,
			OK:      false,
			Message: ".clangd missing CompileFlags section",
			Fix:     "run: fe add clangd  (backs up existing file)",
		}
	}
	return Result{Name: name, OK: true, Message: ".clangd looks good"}
}

func checkCompileCommands(root string) Result {
	name := "compile_commands.json"
	candidates := []string{
		filepath.Join(root, "compile_commands.json"),
		filepath.Join(root, "build", "compile_commands.json"),
		filepath.Join(root, "cmake-build-debug", "compile_commands.json"),
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			content, _ := os.ReadFile(path)
			var arr []interface{}
			if err := json.Unmarshal(content, &arr); err != nil {
				return Result{
					Name:    name,
					OK:      false,
					Message: fmt.Sprintf("%s is not valid JSON", path),
					Fix:     "regenerate: bear -- make  or  cmake -B build ...",
				}
			}
			if len(arr) == 0 {
				return Result{
					Name:    name,
					OK:      false,
					Message: fmt.Sprintf("%s is an empty stub — clangd won't index", path),
					Fix:     "run: bear -- make  or  cmake -B build -DCMAKE_EXPORT_COMPILE_COMMANDS=ON",
				}
			}
			return Result{Name: name, OK: true, Message: fmt.Sprintf("%s (%d entries)", path, len(arr))}
		}
	}
	return Result{
		Name:    name,
		OK:      false,
		Message: "not found — clangd will not index your code",
		Fix:     "run: fe add clangd  then  bear -- make  or cmake",
	}
}

func checkCMakeExportFlag(root string) Result {
	name := "CMAKE_EXPORT_COMPILE_COMMANDS"
	content, err := os.ReadFile(filepath.Join(root, "CMakeLists.txt"))
	if err != nil {
		return Result{Name: name, OK: false, Message: "CMakeLists.txt not found"}
	}
	if strings.Contains(string(content), "CMAKE_EXPORT_COMPILE_COMMANDS") {
		return Result{Name: name, OK: true, Message: "set in CMakeLists.txt"}
	}
	return Result{
		Name:    name,
		OK:      false,
		Message: "not set in CMakeLists.txt",
		Fix:     "add: set(CMAKE_EXPORT_COMPILE_COMMANDS ON)",
	}
}

func checkOLSConfig(root string) Result {
	name := "ols.json"
	path := filepath.Join(root, "ols.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return Result{Name: name, OK: false, Message: "ols.json not found", Fix: "run: fe add ols"}
	}
	content, _ := os.ReadFile(path)
	var cfg map[string]interface{}
	if err := json.Unmarshal(content, &cfg); err != nil {
		return Result{Name: name, OK: false, Message: "ols.json is not valid JSON", Fix: "run: fe add ols"}
	}
	return Result{Name: name, OK: true, Message: "ols.json looks good"}
}

func checkZLSConfig(root string) Result {
	name := "zls.json"
	if _, err := os.Stat(filepath.Join(root, "zls.json")); os.IsNotExist(err) {
		return Result{Name: name, OK: false, Message: "zls.json not found", Fix: "run: fe add zls"}
	}
	return Result{Name: name, OK: true, Message: "zls.json present"}
}

func checkGoMod(root string) Result {
	name := "go.mod"
	if _, err := os.Stat(filepath.Join(root, "go.mod")); os.IsNotExist(err) {
		return Result{Name: name, OK: false, Message: "go.mod not found", Fix: "run: go mod init <module>"}
	}
	return Result{Name: name, OK: true, Message: "go.mod present"}
}
