package build

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/flushwhy/fe/internal/config"
)

type Runner func(command string) error

var DefaultRunner Runner = func(command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type Options struct {
	Platform string
}

func Run(cfg config.BuildConfig, opts Options, runner Runner) error {
	if len(cfg.Targets) == 0 {
		return fmt.Errorf("no build targets defined in .fe.yaml")
	}
	targets := cfg.Targets
	if opts.Platform != "" {
		targets = FilterByPlatform(cfg.Targets, opts.Platform)
		if len(targets) == 0 {
			return fmt.Errorf("no target found for platform %q", opts.Platform)
		}
	}
	fmt.Printf("Building %d target(s)...\n", len(targets))
	for _, t := range targets {
		if err := RunTarget(t, runner); err != nil {
			return err
		}
	}
	fmt.Printf("✓ Build complete\n")
	return nil
}

// RunTarget builds a single target. Exported so cmd/build.go can call it per-spinner.
func RunTarget(t config.BuildTarget, runner Runner) error {
	if err := runner(t.Cmd); err != nil {
		return fmt.Errorf("build failed for %s: %w", t.Platform, err)
	}
	return nil
}

// FilterByPlatform returns only targets matching the given platform name.
func FilterByPlatform(targets []config.BuildTarget, platform string) []config.BuildTarget {
	var out []config.BuildTarget
	for _, t := range targets {
		if strings.EqualFold(t.Platform, platform) {
			out = append(out, t)
		}
	}
	return out
}
