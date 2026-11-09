package compdb

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidate(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "main.cpp")
	if err := os.WriteFile(source, []byte("int main() {}"), 0644); err != nil {
		t.Fatal(err)
	}
	db := filepath.Join(root, "compile_commands.json")
	valid := `[{"directory":"` + filepath.ToSlash(root) + `","file":"main.cpp","arguments":["clang++","-c","main.cpp"]}]`
	if err := os.WriteFile(db, []byte(valid), 0644); err != nil {
		t.Fatal(err)
	}
	if count, err := Validate(db); err != nil || count != 1 {
		t.Fatalf("Validate() = %d, %v; want 1, nil", count, err)
	}

	for _, data := range []string{
		"not json",
		"[]",
		`[{"file":"main.cpp","command":"clang++"}]`,
		`[{"directory":"` + root + `","file":"missing.cpp","command":"clang++"}]`,
	} {
		if err := os.WriteFile(db, []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := Validate(db); err == nil {
			t.Errorf("Validate(%q) succeeded; want error", data)
		}
	}
}

func TestVSCodeBuildDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".vscode"), 0755); err != nil {
		t.Fatal(err)
	}
	settings := `{"cmake.buildDirectory":"${workspaceFolder}/out/build/debug"}`
	if err := os.WriteFile(filepath.Join(root, ".vscode", "settings.json"), []byte(settings), 0644); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "out", "build", "debug")
	if got := vscodeBuildDir(root); got != want {
		t.Fatalf("vscodeBuildDir() = %q, want %q", got, want)
	}
}

func TestFindBuildDirUsesVSCodeConfiguration(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "CMakeLists.txt"), []byte("cmake_minimum_required(VERSION 3.16)"), 0644); err != nil {
		t.Fatal(err)
	}
	configured := filepath.Join(root, "out", "build", "debug")
	if err := os.MkdirAll(filepath.Join(root, ".vscode"), 0755); err != nil {
		t.Fatal(err)
	}
	settings := `{"cmake.buildDirectory":"${workspaceFolder}/out/build/debug"}`
	if err := os.WriteFile(filepath.Join(root, ".vscode", "settings.json"), []byte(settings), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := findBuildDir(root)
	if err != nil || got != configured {
		t.Fatalf("findBuildDir() = %q, %v; want %q", got, err, configured)
	}
}
