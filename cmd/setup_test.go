package cmd

import (
	"bytes"
	"strings"
	"testing"

	projectsetup "github.com/flushwhy/fe/internal/setup"
)

func TestPrintSetupResults(t *testing.T) {
	var out bytes.Buffer
	printSetupResults(&out, "./demo", []projectsetup.Tool{
		{Name: "Go", Required: true, Reason: "build the project", Found: "/usr/bin/go"},
		{Name: "Blender", Required: false, Reason: "build the extension"},
		{Name: "CMake", Required: true, Reason: "configure the project"},
	})

	output := out.String()
	for _, want := range []string{
		"Directory: ./demo",
		"Go",
		"FOUND",
		"/usr/bin/go",
		"Blender",
		"optional",
		"Optional; install only if you use this workflow.",
		"CMake",
		"MISSING",
		"1 required tool(s) missing",
		"Next: install the missing tools",
		"no files were changed",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("setup output missing %q:\n%s", want, output)
		}
	}
}

func TestPrintSetupResultsNoRequirements(t *testing.T) {
	var out bytes.Buffer
	printSetupResults(&out, ".", nil)

	output := out.String()
	if !strings.Contains(output, "No tool requirements detected.") {
		t.Fatalf("expected empty setup result, got:\n%s", output)
	}
	if !strings.Contains(output, "fe inspect") {
		t.Fatalf("expected inspect guidance, got:\n%s", output)
	}
}
