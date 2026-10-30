package build

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/flushwhy/fe/internal/config"
)

// Runner executes a program with its arguments without invoking a shell.
type Runner func(command string, args ...string) error

var DefaultRunner Runner = func(command string, args ...string) error {
	cmd := exec.Command(command, args...)
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

// RunTarget supports legacy command strings while preserving arguments exactly
// when they are supplied separately in .fe.yaml.
func RunTarget(t config.BuildTarget, runner Runner) error {
	parts := strings.Fields(t.Cmd)
	if len(parts) == 0 {
		return fmt.Errorf("build target %q has no command", t.Platform)
	}
	args := append(parts[1:], t.Args...)
	if err := runner(parts[0], args...); err != nil {
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
