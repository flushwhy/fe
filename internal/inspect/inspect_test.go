package inspect

import (
	"os"
	"path/filepath"
	"reflect"
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
	if len(report.Suggestions) != 1 || report.Suggestions[0].Command != "go" {
		t.Fatalf("unexpected build suggestions: %#v", report.Suggestions)
	}
}

func TestInspectSuggestsRootGoAndRustBuilds(t *testing.T) {
	root := t.TempDir()
	for name, content := range map[string]string{
		"go.mod":    "module example.com/demo\n",
		"Cargo.toml": "[package]\nname = \"demo\"\nversion = \"0.1.0\"\n",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	report, err := Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Suggestions) != 2 {
		t.Fatalf("got %#v, want Go and Rust suggestions", report.Suggestions)
	}
	if !reflect.DeepEqual(report.Suggestions[0].Args, []string{"build", "./..."}) {
		t.Fatalf("unexpected Go suggestion: %#v", report.Suggestions[0])
	}
	if !reflect.DeepEqual(report.Suggestions[1].Args, []string{"build", "--release"}) {
		t.Fatalf("unexpected Rust suggestion: %#v", report.Suggestions[1])
	}
}

func TestInspectDoesNotSuggestBuildForNestedManifest(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "service")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "go.mod"), []byte("module example.com/service\n"), 0644); err != nil {
		t.Fatal(err)
	}

	report, err := Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Suggestions) != 0 {
		t.Fatalf("nested project should not receive root build suggestion: %#v", report.Suggestions)
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
