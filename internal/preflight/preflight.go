package preflight

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flushwhy/fe/internal/config"
)

type Result struct {
	Name    string
	OK      bool
	Message string
	Fix     string
}

func Run(cfg config.PreflightConfig, stage, root string) ([]Result, error) {
	if root == "" {
		root = "."
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve project directory: %w", err)
	}
	results := make([]Result, 0)
	if cfg.Detect {
		results = append(results, detectPython(root)...)
		results = append(results, detectLua(root)...)
	}
	for _, check := range cfg.Checks {
		if check.Stage != "" && check.Stage != stage && stage != "all" {
			continue
		}
		if check.Command == "" {
			results = append(results, Result{Name: check.Name, OK: false, Message: "no command configured", Fix: check.Fix})
			continue
		}
		bin, err := exec.LookPath(check.Command)
		if err != nil {
			results = append(results, Result{Name: check.Name, OK: false, Message: fmt.Sprintf("%q is not installed or not in PATH", check.Command), Fix: check.Fix})
			continue
		}
		run := exec.Command(bin, check.Args...)
		run.Dir = root
		output, err := run.CombinedOutput()
		message := strings.TrimSpace(string(output))
		if err != nil {
			if message == "" {
				message = err.Error()
			}
			results = append(results, Result{Name: check.Name, OK: false, Message: message, Fix: check.Fix})
			continue
		}
		if message == "" {
			message = "command succeeded"
		}
		name := check.Name
		if name == "" {
			name = check.Command
		}
		results = append(results, Result{Name: name, OK: true, Message: message})
	}
	return results, nil
}

func detectPython(root string) []Result {
	if !hasAny(root, ".py", "pyproject.toml", "requirements.txt", "Pipfile", "setup.py") {
		return nil
	}
	for _, name := range []string{"python3", "python"} {
		if path, err := exec.LookPath(name); err == nil {
			return []Result{{Name: "Python", OK: true, Message: path}}
		}
	}
	return []Result{{Name: "Python", OK: false, Message: "Python project detected, but no Python interpreter was found", Fix: "Install Python 3 from https://www.python.org/downloads/ or your OS package manager"}}
}

func detectLua(root string) []Result {
	if !hasAny(root, ".lua", ".rockspec") {
		return nil
	}
	for _, name := range []string{"lua", "luajit"} {
		if path, err := exec.LookPath(name); err == nil {
			return []Result{{Name: "Lua", OK: true, Message: path}}
		}
	}
	return []Result{{Name: "Lua", OK: false, Message: "Lua project detected, but neither lua nor luajit was found", Fix: "Install Lua or LuaJIT using your OS package manager; install LuaRocks if the project has a .rockspec"}}
}

func hasAny(root string, extensions ...string) bool {
	found := false
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != root && ignoredDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		for _, item := range extensions {
			if strings.HasPrefix(item, ".") && strings.EqualFold(filepath.Ext(path), item) || filepath.Base(path) == item {
				found = true
				return filepath.SkipAll
			}
		}
		return nil
	})
	return found
}

func ignoredDir(name string) bool {
	switch name {
	case ".git", ".fe", "build", "dist", "vendor", "node_modules", "__pycache__", ".venv", "venv":
		return true
	default:
		return strings.HasPrefix(name, ".")
	}
}
