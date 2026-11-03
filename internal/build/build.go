package build

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/flushwhy/fe/internal/config"
)

type Runner func(command string) error

// DefaultRunner splits configured commands into argv and never invokes a shell.
// Shell operators, expansions, and redirections are intentionally not interpreted.
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
	Target   string
	Platform string
}

func Run(cfg config.BuildConfig, opts Options, runner Runner) error {
	if len(cfg.Targets) == 0 {
		return fmt.Errorf("no build targets defined in .fe.yaml")
	}
	if opts.Target != "" && opts.Platform != "" {
		return fmt.Errorf("choose either --target or --platform, not both")
	}

	targets := cfg.Targets
	if opts.Target != "" {
		targets = FilterByName(targets, opts.Target)
		if len(targets) == 0 {
			return fmt.Errorf("no build target named %q", opts.Target)
		}
		if len(targets) > 1 {
			return fmt.Errorf("build target name %q is not unique", opts.Target)
		}
	}
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
	label := t.Name
	if label == "" {
		label = t.Platform
	}
	if err := runner(t.Cmd); err != nil {
		return fmt.Errorf("build failed for %s: %w", label, err)
	}
	return nil
}

// FilterByName returns targets with an exact, case-insensitive name match.
func FilterByName(targets []config.BuildTarget, name string) []config.BuildTarget {
	var out []config.BuildTarget
	for _, t := range targets {
		if t.Name != "" && strings.EqualFold(t.Name, name) {
			out = append(out, t)
		}
	}
	return out
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
