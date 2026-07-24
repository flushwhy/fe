package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	t.Run("creates standard directory structure for odin", func(t *testing.T) {
		tmp := t.TempDir()

		err := Init(Options{
			Lang: LangOdin,
			Name: "my-game",
			Dir:  tmp,
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
			Lang: LangC,
			Name: "my-lib",
			Dir:  tmp,
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
			Lang: LangCpp,
			Name: "my-engine",
			Dir:  tmp,
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
			Lang:   LangGo,
			Name:   "myapp",
			Dir:    tmp,
			Module: "github.com/test/myapp",
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
			Lang: LangOdin,
			Name: "existing",
			Dir:  tmp,
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
