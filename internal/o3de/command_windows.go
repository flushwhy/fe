//go:build windows

package o3de

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func newPlatformCLICommand(ctx context.Context, cliPath string, args []string) (*exec.Cmd, error) {
	ext := strings.ToLower(filepath.Ext(cliPath))
	if ext != ".bat" && ext != ".cmd" {
		return exec.CommandContext(ctx, cliPath, args...), nil
	}

	executable, invocationArgs, err := CLIInvocation("windows", cliPath, args)
	if err != nil {
		return nil, err
	}
	if executable != "cmd.exe" || len(invocationArgs) != 5 {
		return nil, fmt.Errorf("invalid Windows batch invocation for %q", cliPath)
	}

	cmd := exec.CommandContext(ctx, executable)
	// cmd.exe has a custom parser and does not use the normal Windows argv
	// quoting rules. Pass the complete command line to CreateProcess directly.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: "cmd.exe " + strings.Join(invocationArgs, " "),
	}
	return cmd, nil
}
