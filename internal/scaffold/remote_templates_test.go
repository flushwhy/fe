package scaffold

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveTemplateRejectsUnsafeIDs(t *testing.T) {
	for _, id := range []string{
		"",
		"../escape",
		"cpp/raylib",
		"CPP-Raylib-Xmake",
		".hidden",
		"trailing-",
	} {
		t.Run(id, func(t *testing.T) {
			_, err := ResolveTemplate(id, false)
			if err == nil {
				t.Fatalf("ResolveTemplate(%q) error = nil, want validation error", id)
			}
			if !strings.Contains(err.Error(), "invalid template ID") {
				t.Fatalf("ResolveTemplate(%q) error = %v, want invalid ID error", id, err)
			}
		})
	}
}

func TestDecodeRepoTreeRejectsTrailingJSON(t *testing.T) {
	_, err := decodeRepoTree(strings.NewReader(`{"tree":[]} {}`))
	if err == nil {
		t.Fatal("decodeRepoTree accepted trailing JSON")
	}
}

func TestDecodeRepoTreeRejectsOversizedResponse(t *testing.T) {
	data := bytes.Repeat([]byte(" "), (5<<20)+1)
	_, err := decodeRepoTree(bytes.NewReader(data))
	if err == nil {
		t.Fatal("decodeRepoTree accepted response larger than 5 MiB")
	}
}

func TestDecodeRepoTreeAcceptsValidDocument(t *testing.T) {
	tree, err := decodeRepoTree(strings.NewReader(`{"sha":"abc","tree":[],"truncated":false}`))
	if err != nil {
		t.Fatalf("decodeRepoTree: %v", err)
	}
	if tree.SHA != "abc" || tree.Trunc || len(tree.Tree) != 0 {
		t.Fatalf("unexpected tree: %+v", tree)
	}
}

func TestCachedTemplateIsSafeRejectsSymlinks(t *testing.T) {
	cache := t.TempDir()
	target := t.TempDir()

	if err := os.Symlink(target, filepath.Join(cache, "linked")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cache, "linked.ready"), []byte("ready"), 0600); err != nil {
		t.Fatal(err)
	}
	if cachedTemplateIsSafe(cache, "linked") {
		t.Fatal("accepted symlinked template directory")
	}

	if err := os.Mkdir(filepath.Join(cache, "regular"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(target, "marker"), filepath.Join(cache, "regular.ready")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if cachedTemplateIsSafe(cache, "regular") {
		t.Fatal("accepted symlinked cache marker")
	}
}

func TestCachedTemplateIsSafeAcceptsRegularCache(t *testing.T) {
	cache := t.TempDir()
	if err := os.Mkdir(filepath.Join(cache, "template"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cache, "template.ready"), []byte("ready"), 0600); err != nil {
		t.Fatal(err)
	}
	if !cachedTemplateIsSafe(cache, "template") {
		t.Fatal("rejected regular cached template")
	}
}

func TestDecodeTemplateCatalog(t *testing.T) {
	data := []byte(`{"version":1,"templates":[{"id":"godot-2d","name":"Godot 2D","description":"A starter project.","category":"Godot","tags":["godot"]},{"id":"c-make","name":"C / Make","description":"A C starter.","category":"C","tags":["c"]}]}`)
	catalog, err := decodeTemplateCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Templates) != 2 || catalog.Templates[0].ID != "c-make" {
		t.Fatalf("catalog was not sorted: %#v", catalog.Templates)
	}
}

func TestDecodeTemplateCatalogRejectsDuplicateIDs(t *testing.T) {
	data := []byte(`{"version":1,"templates":[{"id":"go-basic","name":"Go","description":"Go starter.","category":"Go"},{"id":"go-basic","name":"Go duplicate","description":"Duplicate.","category":"Go"}]}`)
	if _, err := decodeTemplateCatalog(data); err == nil {
		t.Fatal("expected duplicate template ID to be rejected")
	}
}

func TestCachedTemplateCatalogSupportsOfflineListing(t *testing.T) {
	cache := t.TempDir()
	data := []byte(`{"version":1,"templates":[{"id":"go-basic","name":"Go basic","description":"A Go starter.","category":"Go"}]}`)
	if err := os.WriteFile(filepath.Join(cache, "catalog.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	templates, err := cachedTemplateCatalog(cache, os.ErrDeadlineExceeded)
	if err != nil {
		t.Fatal(err)
	}
	if len(templates) != 1 || templates[0].ID != "go-basic" {
		t.Fatalf("unexpected offline catalog: %#v", templates)
	}
}

func TestReplaceTemplateCachePreservesPreviousCacheOnFailure(t *testing.T) {
	cache := t.TempDir()
	dest := filepath.Join(cache, "go-basic")
	if err := os.Mkdir(dest, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "main.go"), []byte("package main"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cache, "go-basic.ready"), []byte("ready"), 0600); err != nil {
		t.Fatal(err)
	}
	err := replaceTemplateCache(cache, "go-basic", filepath.Join(cache, "missing-temp"), dest)
	if err == nil {
		t.Fatal("expected replacement to fail")
	}
	if !cachedTemplateIsSafe(cache, "go-basic") {
		t.Fatal("failed refresh invalidated the previous cache")
	}
	if _, err := os.Stat(filepath.Join(dest, "main.go")); err != nil {
		t.Fatalf("previous cache was not restored: %v", err)
	}
}
