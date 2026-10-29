package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check [path]",
	Short: "Run available static analyzers for the detected project languages.",
	Long: `Detects project languages and runs installed static analyzers without modifying files.

  Go:         go vet ./...
  C/C++:      clang-tidy (requires compile_commands.json)
  Python:     ruff check .
  JavaScript: eslint .

If a supported language is found but its analyzer is missing, the command prints
what to install to enable that check. The command fails if an analyzer finds
problems or no applicable analyzer can run.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCheck,
}

type checkTask struct {
	name string
	bin  string
	args []string
	dir  string
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	root := "."
	if len(args) == 1 {
		root = args[0]
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve project path: %w", err)
	}
	tasks, skipped, err := discoverCheckTasks(root)
	if err != nil {
		return err
	}
	for _, message := range skipped {
		fmt.Printf("SKIP  %s\n", message)
	}
	if len(tasks) == 0 {
		return fmt.Errorf("no static analyzer could run; install a supported analyzer or configure the project first")
	}

	failed := 0
	for _, task := range tasks {
		fmt.Printf("\nCHECK %s\n", task.name)
		run := exec.CommandContext(cmd.Context(), task.bin, task.args...)
		run.Dir = task.dir
		run.Stdout = os.Stdout
		run.Stderr = os.Stderr
		if err := run.Run(); err != nil {
			failed++
			fmt.Printf("FAIL  %s: %v\n", task.name, err)
		} else {
			fmt.Printf("PASS  %s\n", task.name)
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d static check(s) failed", failed)
	}
	return nil
}

func discoverCheckTasks(root string) ([]checkTask, []string, error) {
	tasks := []checkTask{}
	skipped := []string{}
	hasGo, hasPython, hasJS, hasCPP := false, false, false, false
	var cppFiles []string

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != root && ignoredCheckDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".go":
			hasGo = true
		case ".py":
			hasPython = true
		case ".js", ".jsx", ".mjs", ".cjs", ".ts", ".tsx":
			hasJS = true
		case ".c", ".cc", ".cpp", ".cxx", ".h", ".hh", ".hpp", ".hxx":
			hasCPP = true
			cppFiles = append(cppFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("scan project: %w", err)
	}

	if hasGo && fileExists(filepath.Join(root, "go.mod")) {
		if bin, err := exec.LookPath("go"); err == nil {
			tasks = append(tasks, checkTask{name: "go vet ./...", bin: bin, args: []string{"vet", "./..."}, dir: root})
		} else {
			skipped = append(skipped, "Go files found, but Go is not installed; install Go from https://go.dev/dl/ to enable 'go vet ./...'")
		}
	}
	if hasCPP {
		if bin, err := exec.LookPath("clang-tidy"); err != nil {
			skipped = append(skipped, "C/C++ files found, but clang-tidy is not installed; install clang-tidy to enable C/C++ checks")
		} else if buildDir := findCompileCommands(root); buildDir == "" {
			skipped = append(skipped, "C/C++ files found, but compile_commands.json is missing; configure CMake with -DCMAKE_EXPORT_COMPILE_COMMANDS=ON to enable clang-tidy")
		} else {
			args := []string{"-p", buildDir}
			args = append(args, cppFiles...)
			tasks = append(tasks, checkTask{name: "clang-tidy", bin: bin, args: args, dir: root})
		}
	}
	if hasPython {
		if bin, err := exec.LookPath("ruff"); err == nil {
			tasks = append(tasks, checkTask{name: "ruff check .", bin: bin, args: []string{"check", "."}, dir: root})
		} else {
			skipped = append(skipped, "Python files found, but Ruff is not installed; run 'python -m pip install ruff' to enable 'ruff check .'")
		}
	}
	if hasJS {
		if bin, err := exec.LookPath("eslint"); err == nil {
			tasks = append(tasks, checkTask{name: "eslint .", bin: bin, args: []string{"."}, dir: root})
		} else {
			skipped = append(skipped, "JavaScript/TypeScript files found, but ESLint is not installed; run 'npm install --save-dev eslint' in the project to enable ESLint checks")
		}
	}
	return tasks, skipped, nil
}

func findCompileCommands(root string) string {
	for _, dir := range []string{"build", "build/debug", "build/release", "cmake-build-debug", "cmake-build-release"} {
		candidate := filepath.Join(root, dir)
		if fileExists(filepath.Join(candidate, "compile_commands.json")) {
			return candidate
		}
	}
	return ""
}

func ignoredCheckDir(name string) bool {
	switch strings.ToLower(name) {
	case ".git", ".fe", "build", "dist", "vendor", "node_modules", "__pycache__", ".venv", "venv", "target":
		return true
	default:
		return strings.HasPrefix(name, ".")
	}
}
