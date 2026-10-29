package scaffold

import (
	"strings"
	"bytes"
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
