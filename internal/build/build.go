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

// Run plans every selected target before executing the first one.
func Run(cfg config.BuildConfig, opts Options, runner Runner) error {
	plan, err := Plan(cfg, opts)
	if err != nil {
		return err
	}
	fmt.Printf("Building %d target(s)...\n", len(plan))
	for _, target := range plan {
		if err := RunPlannedTarget(target, runner); err != nil {
			return err
		}
	}
	fmt.Printf("✓ Build complete\n")
	return nil
}

// RunTarget is retained for callers that execute a single configured target.
func RunTarget(target config.BuildTarget, runner Runner) error {
	plan, err := Plan(config.BuildConfig{Targets: []config.BuildTarget{target}}, Options{})
	if err != nil {
		return err
	}
	return RunPlannedTarget(plan[0], runner)
}

// FilterByPlatform returns only targets matching the given platform name.
func FilterByPlatform(targets []config.BuildTarget, platform string) []config.BuildTarget {
	var out []config.BuildTarget
	for _, target := range targets {
		if strings.EqualFold(target.Platform, platform) {
			out = append(out, target)
		}
	}
	return out
}
