package cmd

import (
	"os"
	"path/filepath"
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

func TestSyncProjectRegistersOrdinaryProjectWithoutInstallingTools(t *testing.T) {
	for _, key := range []string{"CI", "GITHUB_ACTIONS", "JENKINS_URL", "BUILDKITE", "BUILD_ID", "TF_BUILD"} {
		t.Setenv(key, "")
	}
	t.Setenv("FE_HOME", t.TempDir())
	project := t.TempDir()
	got, err := syncProject(project)
	if err != nil {
		t.Fatal(err)
	}
	if got == "" {
		t.Fatal("syncProject returned an empty project path")
	}
	for _, name := range []string{"fe-uninstall.sh", "fe-uninstall.bat", ".gitignore"} {
		if _, err := os.Stat(filepath.Join(project, name)); err != nil {
			t.Fatalf("sync did not create %s: %v", name, err)
		}
	}
}
