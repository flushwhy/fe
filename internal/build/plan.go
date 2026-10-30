package build

import (
	"fmt"
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

// Plan selects targets and resolves their commands without executing anything.
func Plan(cfg config.BuildConfig, opts Options) ([]PlannedTarget, error) {
	if len(cfg.Targets) == 0 {
		return nil, fmt.Errorf("no build targets defined in .fe.yaml")
	}

	targets := cfg.Targets
	if opts.Platform != "" {
		targets = FilterByPlatform(targets, opts.Platform)
		if len(targets) == 0 {
			return nil, fmt.Errorf("no target found for platform %q", opts.Platform)
		}
	}

	plan := make([]PlannedTarget, 0, len(targets))
	for _, target := range targets {
		parts := strings.Fields(target.Cmd)
		if len(parts) == 0 {
			return nil, fmt.Errorf("build target %q has no command", target.Platform)
		}

		args := append([]string{}, parts[1:]...)
		args = append(args, target.Args...)
		plan = append(plan, PlannedTarget{
			Platform: target.Platform,
			Command:  parts[0],
			Args:     args,
			Output:   target.Output,
		})
	}
	return plan, nil
}

// RunPlannedTarget executes a target that has already passed planning.
func RunPlannedTarget(target PlannedTarget, runner Runner) error {
	if err := runner(target.Command, target.Args...); err != nil {
		return fmt.Errorf("build failed for %s: %w", target.Platform, err)
	}
	return nil
}
