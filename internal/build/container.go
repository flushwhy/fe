package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flushwhy/fe/internal/config"
)

// RunContainer runs a configured build command in a disposable container.
// The project directory is mounted at /src, so build artifacts remain on the host.
func RunContainer(cfg config.ContainerBuild, runtimeName string) error {
	if strings.TrimSpace(cfg.Image) == "" {
		return fmt.Errorf("build.container.image is required in .fe.yaml")
	}
	if strings.TrimSpace(cfg.Command) == "" {
		return fmt.Errorf("build.container.command is required in .fe.yaml")
	}

	runtimeName = strings.ToLower(strings.TrimSpace(runtimeName))
	if runtimeName == "" {
		runtimeName = "docker"
	}
	if runtimeName != "docker" && runtimeName != "podman" {
		return fmt.Errorf("unsupported container runtime %q (choose docker or podman)", runtimeName)
	}
	if _, err := exec.LookPath(runtimeName); err != nil {
		return fmt.Errorf("%s is required for container builds; install it and ensure it is on PATH", runtimeName)
	}

	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting project directory: %w", err)
	}
	projectDir, err = filepath.Abs(projectDir)
	if err != nil {
		return fmt.Errorf("resolving project directory: %w", err)
	}

	fmt.Printf("Building in %s container (%s)\n", runtimeName, cfg.Image)
	args := []string{
		"run", "--rm",
		"-v", projectDir + ":/src",
		"-w", "/src",
		cfg.Image,
		"sh", "-lc", cfg.Command,
	}
	cmd := exec.Command(runtimeName, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("container build failed: %w", err)
	}
	fmt.Println("Container build complete")
	return nil
}
