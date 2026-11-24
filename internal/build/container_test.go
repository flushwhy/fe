package build

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/flushwhy/fe/internal/config"
)

func TestRunContainerRequiresImageAndCommand(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.ContainerBuild
		want string
	}{
		{name: "image", cfg: config.ContainerBuild{Command: "odin build ."}, want: "build.container.image is required"},
		{name: "command", cfg: config.ContainerBuild{Image: "flushy/odin:nightly"}, want: "build.container.command is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RunContainer(tt.cfg, "docker")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("RunContainer() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestRunContainerRejectsUnknownRuntime(t *testing.T) {
	err := RunContainer(config.ContainerBuild{Image: "example:latest", Command: "true"}, "nerdctl")
	if err == nil || !strings.Contains(err.Error(), "choose docker or podman") {
		t.Fatalf("RunContainer() error = %v, want unsupported runtime error", err)
	}
}

func TestRunContainerReportsRuntimeExitCode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a POSIX test executable")
	}
	dir := t.TempDir()
	docker := filepath.Join(dir, "docker")
	if err := os.WriteFile(docker, []byte("#!/bin/sh\nexit 19\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	err := RunContainer(config.ContainerBuild{Image: "example:latest", Command: "true"}, "docker")
	if err == nil {
		t.Fatal("expected container runtime failure")
	}
	if !strings.Contains(err.Error(), "docker run exited with code 19") {
		t.Fatalf("RunContainer() error = %q, want runtime name and exit code", err)
	}
}
