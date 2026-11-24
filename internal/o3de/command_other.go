//go:build !windows

package o3de

import (
	"context"
	"os/exec"
)

func newPlatformCLICommand(ctx context.Context, cliPath string, args []string) (*exec.Cmd, error) {
	return exec.CommandContext(ctx, cliPath, args...), nil
}
