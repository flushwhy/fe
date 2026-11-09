package cmd

import (
	"errors"
	"strings"
	"testing"
)

func TestO3DEBuildFailureViewIncludesSummaryAndOutput(t *testing.T) {
	model := newO3DEBuildFailureModel(
		"build/linux",
		errors.New("exit status 1"),
		"src/main.cpp:4: fatal error: missing.hpp: No such file or directory\n",
	)

	view := model.View()
	for _, want := range []string{
		"O3DE build failed",
		"build/linux",
		"What to check",
		"header or file is missing",
		"Build output",
		"missing.hpp",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("failure view does not contain %q:\n%s", want, view)
		}
	}
}

func TestO3DEBuildFailureViewHandlesEmptyOutput(t *testing.T) {
	model := newO3DEBuildFailureModel("build", errors.New("exit status 1"), "")
	if !strings.Contains(model.View(), "CMake did not produce any output.") {
		t.Fatal("expected placeholder for empty build output")
	}
}
