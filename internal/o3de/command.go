package o3de

import (
	"path/filepath"
	"strings"
)

// CLIInvocation returns the executable and arguments needed to run an O3DE CLI
// script on the current platform.
func CLIInvocation(goos, cliPath string, args []string) (string, []string) {
	ext := strings.ToLower(filepath.Ext(cliPath))
	if goos == "windows" && (ext == ".bat" || ext == ".cmd") {
		commandArgs := make([]string, 0, len(args)+4)
		commandArgs = append(commandArgs, "/d", "/c", "call", cliPath)
		commandArgs = append(commandArgs, args...)
		return "cmd.exe", commandArgs
	}
	return cliPath, args
}
