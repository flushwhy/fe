// Package archivepath validates portable archive entry names.
package archivepath

import (
	"fmt"
	"path"
	"strings"
	"unicode"
)

// Validate rejects paths that could escape an archive root or be interpreted
// differently by Windows extraction tools.
func Validate(name string) error {
	if name == "" || strings.Contains(name, "\\") || path.IsAbs(name) || path.Clean(name) != name || name == "." || name == ".." || strings.HasPrefix(name, "../") {
		return fmt.Errorf("invalid or non-canonical archive path")
	}
	for _, part := range strings.Split(name, "/") {
		if part == "" || part == "." || part == ".." || strings.TrimRight(part, " .") != part {
			return fmt.Errorf("invalid archive path component %q", part)
		}
		for _, r := range part {
			if r < 32 || strings.ContainsRune(`<>:"|?*`, r) {
				return fmt.Errorf("archive path component %q is not portable", part)
			}
		}
		base := strings.ToUpper(strings.SplitN(part, ".", 2)[0])
		switch base {
		case "CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
			return fmt.Errorf("reserved Windows device name %q", part)
		}
		for _, r := range part {
			if unicode.IsControl(r) {
				return fmt.Errorf("archive path component %q contains a control character", part)
			}
		}
	}
	return nil
}
