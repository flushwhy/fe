package butler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Logger is satisfied by log.Logger and any test double.
type Logger interface {
	Printf(format string, v ...interface{})
}

// CommandExecutor lets tests inject a fake exec.Command.
type CommandExecutor func(name string, arg ...string) *exec.Cmd

// PlatformMap maps normalised folder names to the itch.io channel suffix.
// Convention: builds/windows-x64/ → channel "windows-x64"
// Also accepts legacy names like "win", "osx" etc.
var platformMap = map[string]string{
	"windows": "windows",
	"win":     "windows",
	"linux":   "linux",
	"macos":   "macos",
	"mac":     "macos",
	"osx":     "macos",
}

var archMap = map[string]string{
	"x64":   "x64",
	"x32":   "x32",
	"32":    "x32",
	"64":    "x64",
	"arm64": "arm64",
	"arm32": "arm32",
	// hyphenated variants used by the convention-based builds/ layout
	"win-x64":  "win-x64",
	"win-x32":  "win-x32",
	"linux-x64": "linux-x64",
	"linux-x32": "linux-x32",
	"linux32":  "linux-x32",
}

// Push reads directory and pushes each platform/arch subfolder to itch.io.
// It supports two layouts:
//
//  1. Convention layout (builds/ root produced by `fe build`):
//     builds/windows-x64/   → channel windows-x64
//     builds/linux-x64/     → channel linux-x64
//
//  2. Legacy two-level layout:
//     builds/windows/x64/   → channel windowsx64
func Push(username, game, directory, userversion string, executor CommandExecutor, logger Logger) error {
	logger.Printf("Pushing to %s/%s from %s", username, game, directory)

	if userversion != "" {
		logger.Printf("Userversion: %s", userversion)
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("could not read directory %s: %w", directory, err)
	}

	pushed := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		nameLower := strings.ToLower(name)

		// Convention layout: single folder named "platform-arch" e.g. "windows-x64"
		if channel, ok := resolveConventionChannel(nameLower); ok {
			fullPath := filepath.Join(directory, name)
			if err := runPush(executor, logger, fullPath, username, game, channel, userversion); err != nil {
				return err
			}
			pushed++
			continue
		}

		// Legacy layout: platform folder containing arch subfolders
		if _, ok := platformMap[nameLower]; ok {
			subEntries, err := os.ReadDir(filepath.Join(directory, name))
			if err != nil {
				logger.Printf("Could not read subdirectory %s: %v", name, err)
				continue
			}

			for _, sub := range subEntries {
				if !sub.IsDir() {
					continue
				}
				subLower := strings.ToLower(sub.Name())
				arch, ok := archMap[subLower]
				if !ok {
					logger.Printf("Skipping %s/%s: unrecognised architecture", name, sub.Name())
					continue
				}

				channel := platformMap[nameLower] + "-" + arch
				fullPath := filepath.Join(directory, name, sub.Name())
				if err := runPush(executor, logger, fullPath, username, game, channel, userversion); err != nil {
					return err
				}
				pushed++
			}
			continue
		}

		logger.Printf("Skipping %s: not a recognised platform folder", name)
	}

	if pushed == 0 {
		return fmt.Errorf("no valid platform folders found in %s", directory)
	}

	logger.Printf("Done — pushed %d channel(s)", pushed)
	return nil
}

// resolveConventionChannel handles the flat "platform-arch" folder convention.
func resolveConventionChannel(name string) (string, bool) {
	// e.g. "windows-x64", "linux-arm64", "macos-x64"
	parts := strings.SplitN(name, "-", 2)
	if len(parts) != 2 {
		return "", false
	}

	platform, ok := platformMap[parts[0]]
	if !ok {
		return "", false
	}

	arch, ok := archMap[parts[1]]
	if !ok {
		return "", false
	}

	return platform + "-" + arch, true
}

func runPush(executor CommandExecutor, logger Logger, path, username, game, channel, userversion string) error {
	pushArgs := []string{
		"push",
		path,
		fmt.Sprintf("%s/%s:%s", username, game, channel),
	}

	if userversion != "" {
		pushArgs = append(pushArgs, "--userversion", userversion)
	}

	cmd := executor("butler", pushArgs...)
	logger.Printf("Running: %s", cmd.String())

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("butler push failed for channel %s: %w\n%s", channel, err, string(out))
	}

	logger.Printf("✓ Pushed channel: %s", channel)
	return nil
}
