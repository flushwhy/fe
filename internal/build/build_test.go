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

	t.Run("selects a named target", func(t *testing.T) {
		named := []config.BuildTarget{
			{Name: "linux-debug", Platform: "linux-x64", Cmd: "make debug"},
			{Name: "linux-release", Platform: "linux-x64", Cmd: "make release"},
		}
		var ran []string
		err := Run(config.BuildConfig{Targets: named}, Options{Target: "LINUX-RELEASE"}, func(cmd string) error {
			ran = append(ran, cmd)
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ran) != 1 || ran[0] != "make release" {
			t.Fatalf("expected only release target, got %v", ran)
		}
	})

	t.Run("fails on unknown target", func(t *testing.T) {
		err := Run(config.BuildConfig{Targets: targets}, Options{Target: "missing"}, nil)
		if err == nil || !strings.Contains(err.Error(), "missing") {
			t.Fatalf("expected missing target error, got %v", err)
		}
	})

	t.Run("rejects duplicate target names", func(t *testing.T) {
		duplicate := []config.BuildTarget{
			{Name: "release", Platform: "linux-x64", Cmd: "make linux"},
			{Name: "release", Platform: "windows-x64", Cmd: "make windows"},
		}
		err := Run(config.BuildConfig{Targets: duplicate}, Options{Target: "release"}, nil)
		if err == nil || !strings.Contains(err.Error(), "not unique") {
			t.Fatalf("expected duplicate target error, got %v", err)
		}
	})

	t.Run("rejects target and platform together", func(t *testing.T) {
		err := Run(config.BuildConfig{Targets: targets}, Options{Target: "release", Platform: "linux-x64"}, nil)
		if err == nil || !strings.Contains(err.Error(), "either --target or --platform") {
			t.Fatalf("expected mutually exclusive selector error, got %v", err)
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
