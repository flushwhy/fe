package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func testTemplateFS(lang Lang) fstest.MapFS {
	files := fstest.MapFS{".gitignore": &fstest.MapFile{Data: []byte("build/\n")}}
	switch lang {
	case LangOdin:
		files["main.odin"] = &fstest.MapFile{Data: []byte("package main\n")}
		files["ols.json"] = &fstest.MapFile{Data: []byte("{}\n")}
	case LangC:
		files["main.c"] = &fstest.MapFile{Data: []byte("int main(void) { return 0; }\n")}
		files["Makefile"] = &fstest.MapFile{Data: []byte("all:\n\tcc main.c -o app\n")}
	case LangCpp:
		files["main.cpp"] = &fstest.MapFile{Data: []byte("int main() { return 0; }\n")}
		files["CMakeLists.txt"] = &fstest.MapFile{Data: []byte("cmake_minimum_required(VERSION 3.20)\n")}
	case LangGo:
		files["main.go"] = &fstest.MapFile{Data: []byte("package main\n// {{.Name}}\n")}
		files["go.mod.tmpl"] = &fstest.MapFile{Data: []byte("module {{.Module}}\n\ngo {{.GoVersion}}\n")}
	case LangZig:
		files["build.zig"] = &fstest.MapFile{Data: []byte("pub fn build() void {}\n")}
		files["src/main.zig"] = &fstest.MapFile{Data: []byte("pub fn main() void {}\n")}
	}
	return files
}

func TestInit(t *testing.T) {
	t.Run("creates standard directory structure for odin", func(t *testing.T) {
		tmp := t.TempDir()

		err := Init(Options{
			Lang:       LangOdin,
			Name:       "my-game",
			Dir:        tmp,
			templateFS: testTemplateFS(LangOdin),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedFiles := []string{
			"my-game/main.odin",
			"my-game/ols.json",
			"my-game/.gitignore",
		}
		for _, f := range expectedFiles {
			p := filepath.Join(tmp, f)
			if _, err := os.Stat(p); os.IsNotExist(err) {
				t.Errorf("expected file %s to exist", p)
			}
		}
	})

	t.Run("creates compile_commands.json stub for c project", func(t *testing.T) {
		tmp := t.TempDir()

		err := Init(Options{
			Lang:       LangC,
			Name:       "my-lib",
			Dir:        tmp,
			templateFS: testTemplateFS(LangC),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stub := filepath.Join(tmp, "my-lib", "compile_commands.json")
		if _, err := os.Stat(stub); os.IsNotExist(err) {
			t.Errorf("expected compile_commands.json stub at %s", stub)
		}
	})

	t.Run("creates compile_commands.json stub for cpp project", func(t *testing.T) {
		tmp := t.TempDir()

		err := Init(Options{
			Lang:       LangCpp,
			Name:       "my-engine",
			Dir:        tmp,
			templateFS: testTemplateFS(LangCpp),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stub := filepath.Join(tmp, "my-engine", "compile_commands.json")
		if _, err := os.Stat(stub); os.IsNotExist(err) {
			t.Errorf("expected compile_commands.json stub for cpp project")
		}
	})

	t.Run("renders project name into template files", func(t *testing.T) {
		tmp := t.TempDir()

		err := Init(Options{
			Lang:       LangGo,
			Name:       "myapp",
			Dir:        tmp,
			Module:     "github.com/test/myapp",
			templateFS: testTemplateFS(LangGo),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(tmp, "myapp", "main.go"))
		if err != nil {
			t.Fatalf("main.go not written: %v", err)
		}
		if !strings.Contains(string(content), "myapp") {
			t.Errorf("expected project name in main.go, got:\n%s", content)
		}
	})

	t.Run("fails when project directory already exists", func(t *testing.T) {
		tmp := t.TempDir()
		os.Mkdir(filepath.Join(tmp, "existing"), 0755)

		err := Init(Options{
			Lang:       LangOdin,
			Name:       "existing",
			Dir:        tmp,
			templateFS: testTemplateFS(LangOdin),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("fails for unsupported language", func(t *testing.T) {
		tmp := t.TempDir()

		err := Init(Options{
			Lang: Lang("cobol"),
			Name: "my-project",
			Dir:  tmp,
		})
		if err == nil {
			t.Fatal("expected error for unsupported lang, got nil")
		}
		if !strings.Contains(err.Error(), "unsupported language") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("fails when name is empty", func(t *testing.T) {
		err := Init(Options{Lang: LangC, Name: ""})
		if err == nil {
			t.Fatal("expected error for empty name, got nil")
		}
	})
}
