package cmd

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/flushwhy/fe/internal/scaffold"
)

func TestPromptInitLanguageAcceptsNumberAndName(t *testing.T) {
	for _, input := range []string{"2\n", string(scaffold.SupportedLangs[1]) + "\n"} {
		var output bytes.Buffer
		got, err := promptInitLanguage(bufio.NewReader(strings.NewReader(input)), &output)
		if err != nil {
			t.Fatalf("promptInitLanguage(%q): %v", input, err)
		}
		if got != scaffold.SupportedLangs[1] {
			t.Fatalf("promptInitLanguage(%q) = %q, want %q", input, got, scaffold.SupportedLangs[1])
		}
	}
}

func TestPromptProjectNameUsesDefault(t *testing.T) {
	got, err := promptProjectName(bufio.NewReader(strings.NewReader("\n")), &bytes.Buffer{}, "my-project")
	if err != nil {
		t.Fatal(err)
	}
	if got != "my-project" {
		t.Fatalf("promptProjectName() = %q, want my-project", got)
	}
}

func TestPromptProjectNameUsesEnteredValue(t *testing.T) {
	got, err := promptProjectName(bufio.NewReader(strings.NewReader("custom-name\n")), &bytes.Buffer{}, "my-project")
	if err != nil {
		t.Fatal(err)
	}
	if got != "custom-name" {
		t.Fatalf("promptProjectName() = %q, want custom-name", got)
	}
}
