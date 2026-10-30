package build

import (
	"errors"
	"reflect"
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
		runner := func(cmd string, args ...string) error {
			ran = append(ran, append([]string{cmd}, args...)...)
			return nil
		}
		err := Run(config.BuildConfig{Targets: targets}, Options{}, runner)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ran) != 8 {
			t.Errorf("expected 8 argv elements across 2 commands, got %d: %v", len(ran), ran)
		}
	})

	t.Run("filters by platform", func(t *testing.T) {
		var ran []string
		runner := func(cmd string, args ...string) error {
			ran = append(ran, cmd)
			return nil
		}
		err := Run(config.BuildConfig{Targets: targets}, Options{Platform: "linux-x64"}, runner)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ran) != 1 || ran[0] != "gcc" {
			t.Errorf("expected one linux command, got %v", ran)
		}
	})

	t.Run("fails on unknown platform", func(t *testing.T) {
		err := Run(config.BuildConfig{Targets: targets}, Options{Platform: "dos-x16"}, nil)
		if err == nil || !strings.Contains(err.Error(), "dos-x16") {
			t.Errorf("expected platform not found error, got: %v", err)
		}
	})

	t.Run("propagates runner error", func(t *testing.T) {
		runner := func(cmd string, args ...string) error {
			return errors.New("compiler not found")
		}
		err := Run(config.BuildConfig{Targets: targets[:1]}, Options{}, runner)
		if err == nil || !strings.Contains(err.Error(), "compiler not found") {
			t.Errorf("expected runner error to propagate, got: %v", err)
		}
	})
}

func TestRunTargetExplicitArgs(t *testing.T) {
	target := config.BuildTarget{
		Platform: "linux-x64",
		Cmd:      "go",
		Args:     []string{"build", "-ldflags=-X main.version=1.2.0", "./..."},
	}
	var gotCommand string
	var gotArgs []string
	runner := func(command string, args ...string) error {
		gotCommand = command
		gotArgs = args
		return nil
	}

	if err := RunTarget(target, runner); err != nil {
		t.Fatal(err)
	}
	if gotCommand != "go" {
		t.Fatalf("command = %q, want go", gotCommand)
	}
	want := []string{"build", "-ldflags=-X main.version=1.2.0", "./..."}
	if !reflect.DeepEqual(gotArgs, want) {
		t.Fatalf("args = %#v, want %#v", gotArgs, want)
	}
}

func TestRunTargetRejectsEmptyCommand(t *testing.T) {
	err := RunTarget(config.BuildTarget{Platform: "linux-x64"}, nil)
	if err == nil || !strings.Contains(err.Error(), "has no command") {
		t.Fatalf("expected empty command error, got %v", err)
	}
}
