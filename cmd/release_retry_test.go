package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flushwhy/fe/internal/config"
)

func TestReleaseJournalSaveAndLoadLast(t *testing.T) {
	old, err := os.Getwd()
	if err != nil { t.Fatal(err) }
	if err := os.Chdir(t.TempDir()); err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = os.Chdir(old) })

	journal := newReleaseJournal("1.2.3", releaseGitState{Commit: "abc123", Branch: "main"}, "rolling", "builds", "zip", "dist/game.zip")
	journal.BuildTargets["target-a"] = true
	journal.ButlerChannels["linux-x64"] = true
	if err := journal.save(); err != nil { t.Fatal(err) }

	loaded, err := loadLastReleaseJournal()
	if err != nil { t.Fatal(err) }
	if loaded.Version != journal.Version || loaded.Commit != journal.Commit || !loaded.BuildTargets["target-a"] || !loaded.ButlerChannels["linux-x64"] {
		t.Fatalf("journal did not round-trip: %+v", loaded)
	}
}

func TestValidateRetryJournalPinsReleaseIdentity(t *testing.T) {
	journal := newReleaseJournal("1.2.3", releaseGitState{Commit: "abc123"}, "", "builds", "", "")
	git := releaseGitState{Commit: "abc123"}
	if err := validateRetryJournal(journal, "1.2.3", git, "", "builds", "", ""); err != nil {
		t.Fatalf("expected matching release identity: %v", err)
	}
	git.Commit = "def456"
	if err := validateRetryJournal(journal, "1.2.3", git, "", "builds", "", ""); err == nil {
		t.Fatal("expected a different commit to be rejected")
	}
}

func TestLoadLastReleaseJournalChoosesMostRecentlyUpdated(t *testing.T) {
	old, err := os.Getwd()
	if err != nil { t.Fatal(err) }
	if err := os.Chdir(t.TempDir()); err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = os.Chdir(old) })

	first := newReleaseJournal("1.0.0", releaseGitState{}, "", "builds", "", "")
	first.UpdatedAt = time.Now().Add(-time.Hour)
	if err := first.save(); err != nil { t.Fatal(err) }
	second := newReleaseJournal("1.1.0", releaseGitState{}, "", "builds", "", "")
	if err := second.save(); err != nil { t.Fatal(err) }

	loaded, err := loadLastReleaseJournal()
	if err != nil { t.Fatal(err) }
	if loaded.Version != "1.1.0" { t.Fatalf("got %s, want latest 1.1.0", loaded.Version) }
}

func TestEnsureNoExistingReleaseJournalBlocksDuplicateVersion(t *testing.T) {
	old, err := os.Getwd()
	if err != nil { t.Fatal(err) }
	if err := os.Chdir(t.TempDir()); err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = os.Chdir(old) })

	journal := newReleaseJournal("1.2.3", releaseGitState{Commit: "abc123"}, "", "builds", "", "")
	if err := journal.save(); err != nil { t.Fatal(err) }
	if err := ensureNoExistingReleaseJournal("1.2.3"); err == nil {
		t.Fatal("expected an existing incomplete release to block a duplicate attempt")
	}
}

func TestReleaseSettingsHashChangesWithBuildConfiguration(t *testing.T) {
	cfg := &config.Config{}
	first, err := releaseSettingsHash(cfg, false, false, false, "", "", "", "")
	if err != nil { t.Fatal(err) }
	cfg.Build.Targets = []config.BuildTarget{{Name: "linux", Platform: "linux-x64", Cmd: "make"}}
	second, err := releaseSettingsHash(cfg, false, false, false, "", "", "", "")
	if err != nil { t.Fatal(err) }
	if first == second { t.Fatal("expected build config change to alter release settings hash") }
}


func TestReleasePackageIntegrityDetectsChangedArchive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "game.zip")
	if err := os.WriteFile(path, []byte("original archive"), 0o600); err != nil {
		t.Fatal(err)
	}
	size, checksum, err := releasePackageIntegrity(path)
	if err != nil {
		t.Fatal(err)
	}
	journal := &releaseJournal{PackageSize: size, PackageSHA256: checksum}
	if err := verifyReleasePackage(journal, path); err != nil {
		t.Fatalf("unchanged archive should verify: %v", err)
	}
	if err := os.WriteFile(path, []byte("modified archive"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := verifyReleasePackage(journal, path); err == nil {
		t.Fatal("expected modified archive to be rejected")
	}
}

func TestVerifyReleasePackageRejectsMissingChecksumAndNonRegularFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "game.zip")
	if err := os.WriteFile(path, []byte("archive"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := verifyReleasePackage(&releaseJournal{}, path); err == nil {
		t.Fatal("expected missing checksum to be rejected")
	}
	if err := verifyReleasePackage(&releaseJournal{PackageSHA256: "bad"}, root); err == nil {
		t.Fatal("expected non-regular package path to be rejected")
	}
}
