package version

import "strings"

// Version and Commit are set by release build scripts.
var (
	Version = "dev"
	Commit  string
)

func Current() string {
	return strings.TrimPrefix(strings.TrimPrefix(Version, "fe-v"), "v")
}
