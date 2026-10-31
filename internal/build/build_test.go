package build

import (
	"errors"
	"strings"
	"testing"

	"github.com/flushwhy/fe/internal/config"
)

func TestRun(t *testing.T) {
	targets := []config.BuildTarget{
		{Platform: "linux-x64", Cmd: "gcc -o builds/linux-x64/game src/main.c"},
		{Platform: "windows-x64", Cmd: "x86_64-w64-mingw32-gcc -o builds/windows-x64/game.exe src/main.c"},
	}

	t.Run("fails with no targets", func(t *testing.T) {
		err := Run(config.BuildConfig{}, Options{}, nil)
		if err == nil || !strings.Contains(err.Error(), "no build targets") {
			t.Errorf("expected no targets error, got: %v", err)
		}
	})

	t.Run("builds all targets", func(t *testing.T) {
		var ran []string
		runner := func(cmd string) error {
			ran = append(ran, cmd)
			return nil
		}

		err := Run(config.BuildConfig{Targets: targets}, Options{}, runner)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ran) != 2 {
			t.Errorf("expected 2 commands run, got %d", len(ran))
		}
	})

	t.Run("filters by platform", func(t *testing.T) {
		var ran []string
		runner := func(cmd string) error {
			ran = append(ran, cmd)
			return nil
		}

		err := Run(config.BuildConfig{Targets: targets}, Options{Platform: "linux-x64"}, runner)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ran) != 1 || !strings.Contains(ran[0], "gcc") {
			t.Errorf("expected 1 linux command, got %v", ran)
		}
	})

	t.Run("fails on unknown platform", func(t *testing.T) {
		err := Run(config.BuildConfig{Targets: targets}, Options{Platform: "dos-x16"}, nil)
		if err == nil || !strings.Contains(err.Error(), "dos-x16") {
			t.Errorf("expected platform not found error, got: %v", err)
		}
	})

	t.Run("propagates runner error", func(t *testing.T) {
		runner := func(cmd string) error {
			return errors.New("compiler not found")
		}

		err := Run(config.BuildConfig{Targets: targets[:1]}, Options{}, runner)
		if err == nil || !strings.Contains(err.Error(), "compiler not found") {
			t.Errorf("expected runner error to propagate, got: %v", err)
		}
	})
}
