package cmd

import (
	"strings"
	"testing"
)

func TestProjectStateCommandsAreDisabledInCI(t *testing.T) {
	t.Setenv("CI", "true")
	if err := syncCmd.RunE(syncCmd, nil); err == nil || !strings.Contains(err.Error(), "disabled in CI") {
		t.Fatalf("expected sync CI guard, got %v", err)
	}
	if err := cleanCmd.RunE(cleanCmd, nil); err == nil || !strings.Contains(err.Error(), "disabled in CI") {
		t.Fatalf("expected clean CI guard, got %v", err)
	}
}
