package cmd

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/flushwhy/fe/internal/tooling"
)

func TestPromptToolAcceptsNumberAndName(t *testing.T) {
	for _, input := range []string{"1\n", string(tooling.SupportedTools[0]) + "\n"} {
		var output bytes.Buffer
		got, err := promptTool(bufio.NewReader(strings.NewReader(input)), &output)
		if err != nil {
			t.Fatalf("promptTool(%q): %v", input, err)
		}
		if got != tooling.SupportedTools[0] {
			t.Fatalf("promptTool(%q) = %q, want %q", input, got, tooling.SupportedTools[0])
		}
	}
}
