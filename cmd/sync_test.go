package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flushwhy/fe/internal/projectstate"
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

func TestSyncProjectReusesMatchingGlobalGodot(t *testing.T) {
	if os.PathSeparator == '\\' {
		t.Skip("uses a Unix executable fixture")
	}
	for _, key := range []string{"CI", "GITHUB_ACTIONS", "JENKINS_URL", "BUILDKITE", "BUILD_ID", "TF_BUILD"} {
		t.Setenv(key, "")
	}
	t.Setenv("FE_HOME", t.TempDir())
	bin := t.TempDir()
	godot := filepath.Join(bin, "godot")
	if err := os.WriteFile(godot, []byte("#!/bin/sh\necho 4.3.stable.official\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "project.godot"), []byte(`config/features=PackedStringArray("4.3", "GL Compatibility")`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := syncProject(project); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := projectstate.ToolPath(project, "godot:4.3-stable"); err != nil || ok {
		t.Fatalf("sync downloaded a local engine despite a matching global one: ok=%v err=%v", ok, err)
	}
}

func TestSyncProjectRootDoesNotClaimParentProject(t *testing.T) {
	for _, key := range []string{"CI", "GITHUB_ACTIONS", "JENKINS_URL", "BUILDKITE", "BUILD_ID", "TF_BUILD"} {
		t.Setenv(key, "")
	}
	t.Setenv("FE_HOME", t.TempDir())
	parent := t.TempDir()
	if err := os.WriteFile(filepath.Join(parent, "project.godot"), []byte(`config/features=PackedStringArray("4.3", "GL Compatibility")`), 0o644); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(parent, "new-project")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := syncProjectRoot(child); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(child, "fe-uninstall.sh")); err != nil {
		t.Fatalf("new project was not synced: %v", err)
	}
	if _, err := os.Stat(filepath.Join(parent, "fe-uninstall.sh")); !os.IsNotExist(err) {
		t.Fatalf("sync claimed the parent project instead: %v", err)
	}
}

func TestConfirmClean(t *testing.T) {
	for _, tc := range []struct {
		answer string
		want   bool
	}{
		{"y\n", true},
		{"YES\n", true},
		{"n\n", false},
		{"\n", false},
	} {
		var output bytes.Buffer
		got, err := confirmClean(bufio.NewReader(strings.NewReader(tc.answer)), &output, "project")
		if err != nil {
			t.Fatalf("confirmClean(%q): %v", tc.answer, err)
		}
		if got != tc.want {
			t.Errorf("confirmClean(%q) = %v, want %v", tc.answer, got, tc.want)
		}
		if !strings.Contains(output.String(), "Remove Fe-managed files from project?") {
			t.Errorf("confirmation prompt missing from %q", output.String())
		}
	}
}
