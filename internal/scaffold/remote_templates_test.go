package scaffold

import (
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
