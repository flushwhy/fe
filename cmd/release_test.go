package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flushwhy/fe/internal/config"
)

func TestValidateReleaseBuildDir(t *testing.T) {
	t.Run("missing directory", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "builds")
		err := validateReleaseBuildDir(dir)
		if err == nil {
			t.Fatal("expected missing build directory to fail")
		}
		if !strings.Contains(err.Error(), "configure build.targets") {
			t.Fatalf("expected actionable configuration guidance, got: %v", err)
		}
	})

	t.Run("directory without platform folders", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "README.txt"), []byte("not a build"), 0o600); err != nil {
			t.Fatal(err)
		}
		err := validateReleaseBuildDir(dir)
		if err == nil || !strings.Contains(err.Error(), "contains no platform folders") {
			t.Fatalf("expected missing platform folders error, got: %v", err)
		}
	})

	t.Run("prebuilt platform folder", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "builds")
		if err := os.MkdirAll(filepath.Join(dir, "linux-x64"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := validateReleaseBuildDir(dir); err != nil {
			t.Fatalf("expected prebuilt output to be accepted, got: %v", err)
		}
	})
}

func TestReleaseSpriteInputsAvailable(t *testing.T) {
	t.Run("missing input is optional", func(t *testing.T) {
		cfg := config.PackConfig{Input: filepath.Join(t.TempDir(), "missing")}
		if releaseSpriteInputsAvailable(cfg) {
			t.Fatal("expected missing sprite input to be unavailable")
		}
	})

	t.Run("existing input directory", func(t *testing.T) {
		cfg := config.PackConfig{Input: t.TempDir()}
		if !releaseSpriteInputsAvailable(cfg) {
			t.Fatal("expected existing sprite input directory to be available")
		}
	})

	t.Run("configured named sheet input", func(t *testing.T) {
		input := t.TempDir()
		cfg := config.PackConfig{
			Input: filepath.Join(t.TempDir(), "sprites"),
			Sheets: map[string]config.PackSheet{
				"player": {Input: input},
			},
		}
		if !releaseSpriteInputsAvailable(cfg) {
			t.Fatal("expected configured named sheet input to be available")
		}
	})
}
