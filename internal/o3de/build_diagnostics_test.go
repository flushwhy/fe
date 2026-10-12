package o3de

import (
	"strings"
	"testing"
)

func TestSummarizeBuildFailureFindsCompilerAndLinkerClues(t *testing.T) {
	output := "src/main.cpp:4: fatal error: missing.hpp: No such file or directory\nundefined reference to `engine_init'\n"
	got := strings.Join(SummarizeBuildFailure(output), "\n")
	if !strings.Contains(got, "header or file is missing") {
		t.Fatalf("expected missing-file guidance, got %q", got)
	}
	if !strings.Contains(got, "Linking failed") {
		t.Fatalf("expected linker guidance, got %q", got)
	}
}

func TestSummarizeBuildFailureHasFallback(t *testing.T) {
	got := SummarizeBuildFailure("process exited with code 1")
	if len(got) == 0 || !strings.Contains(got[0], "CMake build failed") {
		t.Fatalf("expected fallback summary, got %#v", got)
	}
}
