package build

import (
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
