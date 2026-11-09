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
	t.Run("writes .clangd without creating an empty compilation database", func(t *testing.T) {
		dir, cleanup := inTempDir(t)
		defer cleanup()

		if err := addClangd(dir); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(dir, ".clangd"))
		if err != nil {
			t.Fatalf(".clangd not written: %v", err)
		}
		if !strings.Contains(string(content), "CompileFlags") {
			t.Error(".clangd missing CompileFlags section")
		}
		if _, err := os.Stat(filepath.Join(dir, "compile_commands.json")); !os.IsNotExist(err) {
			t.Error("add clangd should not create an empty compilation database")
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


func TestConfigWritersRejectSymlinks(t *testing.T) {
	if os.PathSeparator == '\\' {
		t.Skip("symlink creation may require elevated privileges on Windows")
	}
	dir := t.TempDir()
	target := filepath.Join(t.TempDir(), "outside.json")
	if err := os.WriteFile(target, []byte("keep me"), 0644); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name string
		file string
		write func(string) error
	}{
		{"ols", "ols.json", addOLS},
		{"zls", "zls.json", addZLS},
		{"gopls", ".gopls.yaml", addGopls},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, tc.file)
			if err := os.Symlink(target, path); err != nil {
				t.Fatal(err)
			}
			if err := tc.write(dir); err == nil {
				t.Fatal("writer accepted a symlink destination")
			}
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
		})
	}

	clangd := filepath.Join(dir, ".clangd")
	if err := os.Symlink(target, clangd); err != nil {
		t.Fatal(err)
	}
	if err := writeFileWithBackup(clangd, []byte("new config")); err == nil {
		t.Fatal("backup writer accepted a symlink destination")
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "keep me" {
		t.Fatalf("symlink target was modified: %q", got)
	}
}

func TestConfigWritersRejectDanglingSymlinks(t *testing.T) {
	if os.PathSeparator == '\\' {
		t.Skip("symlink creation may require elevated privileges on Windows")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "zls.json")
	if err := os.Symlink(filepath.Join(dir, "missing-target"), path); err != nil {
		t.Fatal(err)
	}
	if err := addZLS(dir); err == nil {
		t.Fatal("writer accepted a dangling symlink destination")
	}
}
