package gumroad

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGumroadHelperProcess(_ *testing.T) {
	if os.Getenv("GO_WANT_GUMROAD_HELPER") != "1" {
		return
	}
	fmt.Fprint(os.Stdout, os.Getenv("GUMROAD_HELPER_OUTPUT"))
	os.Exit(0)
}

func testExecutor(t *testing.T, responses []string, calls *[][]string) CommandExecutor {
	t.Helper()
	index := 0
	return func(name string, args ...string) *exec.Cmd {
		*calls = append(*calls, append([]string{name}, args...))
		response := ""
		if index < len(responses) {
			response = responses[index]
		}
		index++
		cmd := exec.Command(os.Args[0], "-test.run=TestGumroadHelperProcess")
		cmd.Env = append(os.Environ(), "GO_WANT_GUMROAD_HELPER=1", "GUMROAD_HELPER_OUTPUT="+response)
		return cmd
	}
}

func writeTestManifest(t *testing.T, version string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "linux-x64"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "linux-x64", "game.zip"), []byte("artifact"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := fmt.Sprintf(`{"version":%q,"artifacts":[{"path":"linux-x64/game.zip"}]}`, version)
	if err := os.WriteFile(filepath.Join(dir, "artifact-manifest.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestPublishStableProduct(t *testing.T) {
	dir := writeTestManifest(t, "1.2.3")
	var calls [][]string
	responses := []string{
		`{"preserved":[]}`,
		`{}`,
		`{"preserved":[{"id":"uploaded-file"}]}`,
	}
	err := Publish("1.2.3", dir, map[string]string{"1.2.3": "product-123"}, testExecutor(t, responses, &calls), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 3 {
		t.Fatalf("expected preflight, upload, and verification calls; got %d", len(calls))
	}
	if calls[0][1] != "products" || calls[0][2] != "update" || calls[0][3] != "product-123" {
		t.Fatalf("unexpected Gumroad command: %v", calls[0])
	}
	if !containsArgs(calls[0], "--dry-run") || containsArgs(calls[1], "--dry-run") {
		t.Fatalf("expected first and third calls to be dry runs and second to upload: %v", calls)
	}
}

func TestPublishRefusesNonEmptyStableProduct(t *testing.T) {
	dir := writeTestManifest(t, "1.2.3")
	var calls [][]string
	err := Publish("1.2.3", dir, map[string]string{"1.2.3": "product-123"}, testExecutor(t, []string{`{"preserved":[{"id":"old-file"}]}`}, &calls), nil)
	if err == nil || !strings.Contains(err.Error(), "already contains") {
		t.Fatalf("expected non-empty product to be rejected, got %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("expected upload to be blocked after preflight, got %d command calls", len(calls))
	}
}

func TestValidateReleaseRejectsManifestVersionMismatch(t *testing.T) {
	dir := writeTestManifest(t, "1.2.2")
	_, _, err := ValidateRelease("1.2.3", dir, map[string]string{"1.2.3": "product-123"})
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("expected manifest version mismatch, got %v", err)
	}
}

func containsArgs(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}
