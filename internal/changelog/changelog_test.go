package changelog

import (
	"strings"
	"testing"
)

func TestFormatGroupsConventionalCommitsAndMarksBreakingChanges(t *testing.T) {
	entries := parse([]byte("feat(cli): add update command\x1f\x1e" +
		"fix: handle missing tags\x1f\x1e" +
		"feat!: remove legacy config\x1fBREAKING CHANGE: migrate to the new config key\x1e"))
	got := Format(entries, "v1.0.0", "HEAD")
	for _, want := range []string{"## Changes (v1.0.0...HEAD)", "### Features", "cli: add update command", "### Bug Fixes", "**BREAKING:** remove legacy config", "migrate to the new config key"} {
		if !strings.Contains(got, want) {
			t.Fatalf("Format() missing %q in:\n%s", want, got)
		}
	}
}
