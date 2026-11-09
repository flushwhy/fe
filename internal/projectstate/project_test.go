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
	nl := string(rune(10))
	original := "build/" + nl + "user-cache/" + nl
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Sync(root); err != nil {
		t.Fatal(err)
	}
	if err := Sync(root); err != nil {
		t.Fatalf("second sync: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(data), ignoreStart) != 1 || !strings.Contains(string(data), original) {
		t.Fatalf("sync changed existing ignore rules or duplicated its block: %q", data)
	}
	for name := range uninstallers {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
	result, err := Clean(root, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Preserved) != 0 {
		t.Fatalf("unexpected preserved files: %v", result.Preserved)
	}
	data, err = os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != original {
		t.Fatalf("cleanup changed user ignore rules: %q", data)
	}
	for name := range uninstallers {
		if _, err := os.Stat(filepath.Join(root, name)); !os.IsNotExist(err) {
			t.Fatalf("%s still exists, err=%v", name, err)
		}
	}
}

func TestSyncCreatesAndCleanRemovesGitignore(t *testing.T) {
	t.Setenv("FE_HOME", t.TempDir())
	root := t.TempDir()
	if err := Sync(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".gitignore")); err != nil {
		t.Fatal(err)
	}
	if _, err := Clean(root, false, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".gitignore")); !os.IsNotExist(err) {
		t.Fatalf("Fe-created .gitignore was not removed: %v", err)
	}
}

func TestCleanPreservesModifiedUninstallerAndRecord(t *testing.T) {
	t.Setenv("FE_HOME", t.TempDir())
	root := t.TempDir()
	if err := Sync(root); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "fe-uninstall.sh")
	if err := os.WriteFile(path, []byte("# user modification" + string(rune(10))), 0o755); err != nil {
		t.Fatal(err)
	}
	result, err := Clean(root, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Preserved) != 1 {
		t.Fatalf("expected modified script to be preserved, got %v", result.Preserved)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("modified script was removed: %v", err)
	}
	if _, err := loadRecord(root); err != nil {
		t.Fatalf("record should remain for manual cleanup: %v", err)
	}
}

func TestCleanDryRunDoesNotChangeProject(t *testing.T) {
	t.Setenv("FE_HOME", t.TempDir())
	root := t.TempDir()
	if err := Sync(root); err != nil {
		t.Fatal(err)
	}
	result, err := Clean(root, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Removed) == 0 {
		t.Fatal("dry run did not report any cleanup")
	}
	if _, err := os.Stat(filepath.Join(root, "fe-uninstall.sh")); err != nil {
		t.Fatal("dry run removed the uninstall script")
	}
	if _, err := loadRecord(root); err != nil {
		t.Fatalf("dry run removed the record: %v", err)
	}
}

func TestSyncRejectsGitignoreSymlink(t *testing.T) {
	t.Setenv("FE_HOME", t.TempDir())
	root := t.TempDir()
	target := filepath.Join(t.TempDir(), "ignore")
	content := "keep/" + string(rune(10))
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
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
	if string(data) != content {
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
	if _, err := os.Stat(filepath.Join(root, ".gitignore")); !os.IsNotExist(err) {
		t.Fatalf("sync changed .gitignore before rejecting the existing script: %v", err)
	}
}

func TestCleanAllRemovesRecordForDeletedProject(t *testing.T) {
	t.Setenv("FE_HOME", t.TempDir())
	root := t.TempDir()
	if err := Sync(root); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(root); err != nil {
		t.Fatal(err)
	}
	results, err := CleanAll(false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || len(results[0].Removed) != 1 {
		t.Fatalf("expected one stale project record to be removed, got %#v", results)
	}
}

func TestCleanAllRejectsUnknownManagedPaths(t *testing.T) {
	t.Setenv("FE_HOME", t.TempDir())
	root := t.TempDir()
	if err := Sync(root); err != nil {
		t.Fatal(err)
	}
	record, err := loadRecord(root)
	if err != nil {
		t.Fatal(err)
	}
	record.ManagedFiles["../../important.txt"] = "not-a-real-hash"
	if err := saveRecord(record); err != nil {
		t.Fatal(err)
	}
	if _, err := CleanAll(false); err == nil {
		t.Fatal("cleanup accepted an unrecognized managed path")
	}
	for name := range uninstallers {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Fatalf("cleanup removed a file before validating the record: %v", err)
		}
	}
}
