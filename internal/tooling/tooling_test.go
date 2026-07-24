package tooling

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func inTempDir(t *testing.T) (string, func()) {
	t.Helper()
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	return tmp, func() { os.Chdir(orig) }
}

func TestAddClangd(t *testing.T) {
	t.Run("writes .clangd and stub compile_commands.json", func(t *testing.T) {
		dir, cleanup := inTempDir(t)
		defer cleanup()

		if err := addClangd(dir); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// .clangd must exist.
		clangdPath := filepath.Join(dir, ".clangd")
		content, err := os.ReadFile(clangdPath)
		if err != nil {
			t.Fatalf(".clangd not written: %v", err)
		}
		if !strings.Contains(string(content), "CompileFlags") {
			t.Error(".clangd missing CompileFlags section")
		}

		// compile_commands.json must exist and be valid JSON.
		ccPath := filepath.Join(dir, "compile_commands.json")
		ccContent, err := os.ReadFile(ccPath)
		if err != nil {
			t.Fatalf("compile_commands.json not written: %v", err)
		}
		var arr []interface{}
		if err := json.Unmarshal(ccContent, &arr); err != nil {
			t.Errorf("compile_commands.json is not valid JSON: %v", err)
		}
	})

	t.Run("detects cmake build dir and sets CompilationDatabase", func(t *testing.T) {
		dir, cleanup := inTempDir(t)
		defer cleanup()

		// Simulate cmake output.
		buildDir := filepath.Join(dir, "build")
		os.MkdirAll(buildDir, 0755)
		os.WriteFile(filepath.Join(buildDir, "compile_commands.json"), []byte("[]\n"), 0644)
		os.WriteFile(filepath.Join(dir, "CMakeLists.txt"), []byte(""), 0644)

		if err := addClangd(dir); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, _ := os.ReadFile(filepath.Join(dir, ".clangd"))
		if !strings.Contains(string(content), "build") {
			t.Error("expected .clangd to reference build dir for cmake projects")
		}
	})

	t.Run("backs up existing .clangd", func(t *testing.T) {
		dir, cleanup := inTempDir(t)
		defer cleanup()

		// Write an existing .clangd.
		existing := filepath.Join(dir, ".clangd")
		os.WriteFile(existing, []byte("old config"), 0644)

		if err := addClangd(dir); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		bak := filepath.Join(dir, ".clangd.bak")
		if _, err := os.Stat(bak); os.IsNotExist(err) {
			t.Error("expected .clangd.bak to exist after backup")
		}
	})
}

func TestAddOLS(t *testing.T) {
	t.Run("writes ols.json", func(t *testing.T) {
		dir, cleanup := inTempDir(t)
		defer cleanup()

		if err := addOLS(dir); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(dir, "ols.json"))
		if err != nil {
			t.Fatalf("ols.json not written: %v", err)
		}

		var cfg map[string]interface{}
		if err := json.Unmarshal(content, &cfg); err != nil {
			t.Errorf("ols.json is not valid JSON: %v", err)
		}

		if _, ok := cfg["enable_document_hover"]; !ok {
			t.Error("ols.json missing enable_document_hover")
		}
	})

	t.Run("patches existing ols.json without overwriting", func(t *testing.T) {
		dir, cleanup := inTempDir(t)
		defer cleanup()

		existing := map[string]interface{}{
			"enable_snippets": true,
		}
		data, _ := json.Marshal(existing)
		os.WriteFile(filepath.Join(dir, "ols.json"), data, 0644)

		if err := addOLS(dir); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, _ := os.ReadFile(filepath.Join(dir, "ols.json"))
		var cfg map[string]interface{}
		json.Unmarshal(content, &cfg)

		// Original field preserved.
		if v, ok := cfg["enable_snippets"]; !ok || v != true {
			t.Error("existing field was overwritten")
		}
		// New field added.
		if _, ok := cfg["enable_document_hover"]; !ok {
			t.Error("missing field was not patched in")
		}
	})
}

func TestAddZLS(t *testing.T) {
	t.Run("writes zls.json", func(t *testing.T) {
		dir, cleanup := inTempDir(t)
		defer cleanup()

		if err := addZLS(dir); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(dir, "zls.json"))
		if err != nil {
			t.Fatalf("zls.json not written: %v", err)
		}

		var cfg map[string]interface{}
		if err := json.Unmarshal(content, &cfg); err != nil {
			t.Errorf("zls.json is not valid JSON: %v", err)
		}
	})
}

func TestFindCompileCommandsDir(t *testing.T) {
	t.Run("finds existing compile_commands.json in build/", func(t *testing.T) {
		dir, cleanup := inTempDir(t)
		defer cleanup()

		buildDir := filepath.Join(dir, "build")
		os.MkdirAll(buildDir, 0755)
		os.WriteFile(filepath.Join(buildDir, "compile_commands.json"), []byte("[]\n"), 0644)

		got := findCompileCommandsDir(dir)
		if got != buildDir {
			t.Errorf("got %s, want %s", got, buildDir)
		}
	})

	t.Run("detects cmake project and returns build/", func(t *testing.T) {
		dir, cleanup := inTempDir(t)
		defer cleanup()

		os.WriteFile(filepath.Join(dir, "CMakeLists.txt"), []byte(""), 0644)

		got := findCompileCommandsDir(dir)
		want := filepath.Join(dir, "build")
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})

	t.Run("defaults to project root", func(t *testing.T) {
		dir, cleanup := inTempDir(t)
		defer cleanup()

		got := findCompileCommandsDir(dir)
		if got != dir {
			t.Errorf("got %s, want %s (root)", got, dir)
		}
	})
}
