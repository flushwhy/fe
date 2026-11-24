package ui

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestWriteErrorIncludesMessageCauseAndHint(t *testing.T) {
	var output bytes.Buffer
	err := &CommandError{
		Message: "could not create project",
		Cause:   errors.New("permission denied"),
		Hint:    "choose a writable directory",
	}

	WriteError(&output, err)
	got := output.String()
	for _, want := range []string{
		"Fe could not complete that command",
		"could not create project",
		"permission denied",
		"Try: choose a writable directory",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("WriteError() = %q, want it to contain %q", got, want)
		}
	}
}

func TestWriteErrorHandlesPlainError(t *testing.T) {
	var output bytes.Buffer
	WriteError(&output, errors.New("something failed"))
	if !strings.Contains(output.String(), "something failed") {
		t.Fatalf("WriteError() = %q, want original error", output.String())
	}
}

func TestWriteErrorIncludesActionableDetails(t *testing.T) {
	var output bytes.Buffer
	err := &CommandError{
		Message: "build configuration is invalid",
		Details: []string{
			"build.targets[0].cmd is required",
			"build.targets[1].platform is required",
		},
		Hint: "run fe validate after correcting .fe.yaml",
	}

	WriteError(&output, err)
	got := output.String()
	for _, want := range []string{
		"build configuration is invalid",
		"- build.targets[0].cmd is required",
		"- build.targets[1].platform is required",
		"Try: run fe validate after correcting .fe.yaml",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("WriteError() = %q, want it to contain %q", got, want)
		}
	}
}
