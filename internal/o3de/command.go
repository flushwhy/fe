package o3de

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
)

// CLIInvocation returns the executable and arguments needed to run an O3DE CLI
// script on the current platform. Windows batch files require cmd.exe, so their
// arguments are quoted as one command string and values that cmd.exe would
// expand or reinterpret are rejected.
func CLIInvocation(goos, cliPath string, args []string) (string, []string, error) {
	ext := strings.ToLower(filepath.Ext(cliPath))
	if goos != "windows" || (ext != ".bat" && ext != ".cmd") {
		return cliPath, args, nil
	}

	values := make([]string, 0, len(args)+1)
	values = append(values, cliPath)
	values = append(values, args...)
	for _, value := range values {
		if strings.Contains(value, "%") || strings.ContainsRune(value, '"' ) || strings.IndexFunc(value, unicode.IsControl) >= 0 {
			return "", nil, fmt.Errorf("cannot safely pass %q to the O3DE Windows batch script: percent signs, quotes, and control characters are not supported", value)
		}
	}

	quoted := make([]string, len(values))
	for i, value := range values {
		quoted[i] = `"` + value + `"`
	}
	command := strings.Join(quoted, " ")

	// /s strips only the outer pair of quotes. /d disables cmd AutoRun
	// commands, and /v:off prevents delayed expansion of exclamation marks.
	command = `"` + command + `"`
	return "cmd.exe", []string{"/d", "/v:off", "/s", "/c", command}, nil
}
