package inspect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectDetectsProjectTypesAndSkipsGeneratedDirectories(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"go.mod": "module example.com/demo\n",
		"addon/blender_manifest.toml": "id = \"demo\"\n",
		"addon/legacy.py": "bl_info = {\"name\": \"Demo\"}\n",
		"web/package.json": "{}\n",
		"vendor/ignored/Cargo.toml": "[package]\n",
		"dist/ignored/pyproject.toml": "[project]\n",
	}
	for name, content := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	report, err := Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]int{}
	for _, project := range report.Projects {
		got[project.Type] = len(project.Markers)
	}
	if got["Go"] != 1 || got["Blender add-on"] != 2 || got["Node.js"] != 1 {
		t.Fatalf("unexpected detected projects: %#v", got)
	}
	if _, ok := got["Rust"]; ok {
		t.Fatal("detected Rust marker inside vendor")
	}
	if _, ok := got["Python"]; ok {
		t.Fatal("detected Python marker inside dist")
	}
}

func TestInspectRejectsFilePath(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Inspect(file); err == nil {
		t.Fatal("expected error for file path")
	}
}

func TestInspectUnknownProjectReturnsEmptyReport(t *testing.T) {
	report, err := Inspect(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Projects) != 0 {
		t.Fatalf("expected no detected projects, got %#v", report.Projects)
	}
}
