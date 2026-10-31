package build

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/flushwhy/fe/internal/config"
)

// PlannedTarget is a validated command ready for execution.
type PlannedTarget struct {
	Platform string
	Command  string
	Args     []string
	Output   string
}

// Plan selects configured targets and resolves their commands without executing anything.
func Plan(cfg config.BuildConfig, opts Options) ([]PlannedTarget, error) {
	if len(cfg.Targets) == 0 {
		return nil, errors.New("no build targets defined in .fe.yaml and no automatic build suggestion is available")
	}

	targets := cfg.Targets
	if opts.Platform != "" {
		targets = FilterByPlatform(targets, opts.Platform)
		if len(targets) == 0 {
			return nil, fmt.Errorf("no target found for platform %q", opts.Platform)
		}
	}
	return resolveTargets(targets)
}

// PlanWithSuggestions uses detected targets only when no explicit targets exist.
// Automatically detected targets are host-only and never satisfy --platform.
func PlanWithSuggestions(cfg config.BuildConfig, opts Options, suggestions []config.BuildTarget) ([]PlannedTarget, error) {
	if len(cfg.Targets) > 0 {
		return Plan(cfg, opts)
	}
	if len(suggestions) == 0 {
		return Plan(cfg, opts)
	}
	if opts.Platform != "" {
		return nil, errors.New("automatic build suggestions target the current host only; remove --platform or configure explicit build.targets in .fe.yaml")
	}

	targets := append([]config.BuildTarget(nil), suggestions...)
	for i := range targets {
		targets[i].Platform = "host"
	}
	return resolveTargets(targets)
}

func resolveTargets(targets []config.BuildTarget) ([]PlannedTarget, error) {
	plan := make([]PlannedTarget, 0, len(targets))
	for _, target := range targets {
		command := strings.TrimSpace(target.Cmd)
		args := append([]string(nil), target.Args...)

		// A nil Args slice identifies legacy configs where Cmd contains argv.
		// Structured configs keep Cmd intact, including executable paths with spaces.
		if target.Args == nil {
			parts := strings.Fields(command)
			if len(parts) == 0 {
				return nil, fmt.Errorf("build target %q has no command", target.Platform)
			}
			command, args = parts[0], parts[1:]
		} else if command == "" {
			return nil, fmt.Errorf("build target %q has no command", target.Platform)
		}

		plan = append(plan, PlannedTarget{
			Platform: target.Platform,
			Command:  command,
			Args:     args,
			Output:   target.Output,
		})
	}
	return plan, nil
}

// CheckTools checks every executable before any target is started.
func CheckTools(plan []PlannedTarget) error {
	checked := make(map[string]struct{}, len(plan))
	for _, target := range plan {
		if _, ok := checked[target.Command]; ok {
			continue
		}
		checked[target.Command] = struct{}{}
		if _, err := exec.LookPath(target.Command); err != nil {
			return fmt.Errorf("build tool %q required for %s was not found in PATH; install it or configure an available command in .fe.yaml", target.Command, target.Platform)
		}
	}
	return nil
}

// RunPlannedTarget executes a target that has already passed planning.
func RunPlannedTarget(target PlannedTarget, runner Runner) error {
	if err := runner(target.Command, target.Args...); err != nil {
		return fmt.Errorf("build failed for %s: %w", target.Platform, err)
	}
	return nil
}
