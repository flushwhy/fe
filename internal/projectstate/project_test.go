package projectstate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncAndCleanPreserveUserGitignore(t *testing.T) {
	t.Setenv("FE_HOME", t.TempDir())
	root := t.TempDir()
	original := "build/\nuser-cache/\n"
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte(original), 0o644); err != nil { t.Fatal(err) }
	if err := Sync(root); err != nil { t.Fatal(err) }
	if err := Sync(root); err != nil { t.Fatalf("second sync: %v", err) }
	data, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil { t.Fatal(err) }
	if strings.Count(string(data), ignoreStart) != 1 || !strings.Contains(string(data), original) {
		t.Fatalf("sync did not preserve existing ignore rules or duplicated its block: %q", data)
	}
	for name := range uninstallers {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil { t.Fatalf("missing %s: %v", name, err) }
	}
	result, err := Clean(root, false, false)
	if err != nil { t.Fatal(err) }
	if len(result.Preserved) != 0 { t.Fatalf("unexpected preserved files: %v", result.Preserved) }
	data, err = os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil { t.Fatal(err) }
	if string(data) != original { t.Fatalf("cleanup changed user ignore rules: %q", data) }
	for name := range uninstallers {
		if _, err := os.Stat(filepath.Join(root, name)); !os.IsNotExist(err) { t.Fatalf("%s still exists, err=%v", name, err) }
	}
}

func TestSyncCreatesAndCleanRemovesGitignore(t *testing.T) {
	t.Setenv("FE_HOME", t.TempDir())
	root := t.TempDir()
	if err := Sync(root); err != nil { t.Fatal(err) }
	if _, err := os.Stat(filepath.Join(root, ".gitignore")); err != nil { t.Fatal(err) }
	if _, err := Clean(root, false, false); err != nil { t.Fatal(err) }
	if _, err := os.Stat(filepath.Join(root, ".gitignore")); !os.IsNotExist(err) { t.Fatalf("Fe-created .gitignore was not removed: %v", err) }
}

func TestCleanPreservesModifiedUninstallerAndRecord(t *testing.T) {
	t.Setenv("FE_HOME", t.TempDir())
	root := t.TempDir()
	if err := Sync(root); err != nil { t.Fatal(err) }
	path := filepath.Join(root, "fe-uninstall.sh")
	if err := os.WriteFile(path, []byte("# user modification\n"), 0o755); err != nil { t.Fatal(err) }
	result, err := Clean(root, false, false)
	if err != nil { t.Fatal(err) }
	if len(result.Preserved) != 1 { t.Fatalf("expected modified script to be preserved, got %v", result.Preserved) }
	if _, err := os.Stat(path); err != nil { t.Fatalf("modified script was removed: %v", err) }
	if _, err := loadRecord(root); err != nil { t.Fatalf("record should remain for manual cleanup: %v", err) }
}

func TestCleanDryRunDoesNotChangeProject(t *testing.T) {
	t.Setenv("FE_HOME", t.TempDir())
	root := t.TempDir()
	if err := Sync(root); err != nil { t.Fatal(err) }
	result, err := Clean(root, true, false)
	if err != nil { t.Fatal(err) }
	if len(result.Removed) == 0 { t.Fatal("dry run did not report any cleanup") }
	if _, err := os.Stat(filepath.Join(root, "fe-uninstall.sh")); err != nil { t.Fatal("dry run removed the uninstall script") }
	if _, err := loadRecord(root); err != nil { t.Fatalf("dry run removed the record: %v", err) }
}

func TestSyncRejectsGitignoreSymlink(t *testing.T) {
	t.Setenv("FE_HOME", t.TempDir())
	root := t.TempDir()
	target := filepath.Join(t.TempDir(), "ignore")
	if err := os.WriteFile(target, []byte("keep/\\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, ".gitignore")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := Sync(root); err == nil {
		t.Fatal("sync followed a .gitignore symlink")
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "keep/\\n" {
		t.Fatalf("symlink target changed: %q", data)
	}
}

func TestSyncDoesNotClaimExistingUninstaller(t *testing.T) {
	t.Setenv("FE_HOME", t.TempDir())
	root := t.TempDir()
	path := filepath.Join(root, "fe-uninstall.sh")
	if err := os.WriteFile(path, []byte(uninstallers["fe-uninstall.sh"]), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Sync(root); err == nil {
		t.Fatal("sync claimed an existing uninstaller without an ownership record")
	}
}

func TestSyncRejectsGitignoreSymlink(t *testing.T) {
	t.Setenv("FE_HOME", t.TempDir())
	root := t.TempDir()
	target := filepath.Join(t.TempDir(), "ignore")
	if err := os.WriteFile(target, []byte("keep/\\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, ".gitignore")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := Sync(root); err == nil {
		t.Fatal("sync followed a .gitignore symlink")
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "keep/\\n" {
		t.Fatalf("symlink target changed: %q", data)
	}
}

func TestSyncDoesNotClaimExistingUninstaller(t *testing.T) {
	t.Setenv("FE_HOME", t.TempDir())
	root := t.TempDir()
	path := filepath.Join(root, "fe-uninstall.sh")
	if err := os.WriteFile(path, []byte(uninstallers["fe-uninstall.sh"]), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Sync(root); err == nil {
		t.Fatal("sync claimed an existing uninstaller without an ownership record")
	}
}
