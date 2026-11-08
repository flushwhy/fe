package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type releaseJournal struct {
	Version        string          `json:"version"`
	Commit         string          `json:"commit"`
	Branch         string          `json:"branch,omitempty"`
	ChannelSuffix  string          `json:"channel_suffix,omitempty"`
	ButlerDirectory string         `json:"butler_directory"`
	PackageFormat  string          `json:"package_format,omitempty"`
	PackageOutput  string          `json:"package_output,omitempty"`
	BuildTargets   map[string]bool `json:"build_targets,omitempty"`
	Stages         map[string]bool `json:"stages,omitempty"`
	ButlerChannels map[string]bool `json:"butler_channels,omitempty"`
	UpdatedAt      time.Time       `json:"updated_at"`
	Complete       bool            `json:"complete,omitempty"`
}

func releaseJournalDir() string { return filepath.Join(".fe", "releases") }

func releaseJournalPath(version string) string {
	sum := sha256.Sum256([]byte(version))
	return filepath.Join(releaseJournalDir(), hex.EncodeToString(sum[:8])+".json")
}

func newReleaseJournal(version string, git releaseGitState, channelSuffix, directory, format, output string) *releaseJournal {
	return &releaseJournal{
		Version: version, Commit: git.Commit, Branch: git.Branch,
		ChannelSuffix: channelSuffix, ButlerDirectory: directory,
		PackageFormat: format, PackageOutput: output,
		BuildTargets: map[string]bool{}, Stages: map[string]bool{},
		ButlerChannels: map[string]bool{}, UpdatedAt: time.Now().UTC(),
	}
}

func (j *releaseJournal) save() error {
	if j == nil { return nil }
	if err := os.MkdirAll(releaseJournalDir(), 0o700); err != nil {
		return fmt.Errorf("create release journal directory: %w", err)
	}
	j.UpdatedAt = time.Now().UTC()
	data, err := json.MarshalIndent(j, "", "  ")
	if err != nil { return fmt.Errorf("encode release journal: %w", err) }
	data = append(data, '\n')
	tmp, err := os.CreateTemp(releaseJournalDir(), ".release-*.tmp")
	if err != nil { return fmt.Errorf("create release journal: %w", err) }
	name := tmp.Name()
	defer os.Remove(name)
	if err := tmp.Chmod(0o600); err != nil { _ = tmp.Close(); return err }
	if _, err := tmp.Write(data); err != nil { _ = tmp.Close(); return err }
	if err := tmp.Sync(); err != nil { _ = tmp.Close(); return err }
	if err := tmp.Close(); err != nil { return err }
	if err := os.Rename(name, releaseJournalPath(j.Version)); err != nil {
		return fmt.Errorf("save release journal: %w", err)
	}
	return nil
}

func loadLastReleaseJournal() (*releaseJournal, error) {
	entries, err := os.ReadDir(releaseJournalDir())
	if err != nil {
		if os.IsNotExist(err) { return nil, fmt.Errorf("no previous release journal found; run a release first") }
		return nil, fmt.Errorf("read release journals: %w", err)
	}
	var latest *releaseJournal
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") { continue }
		data, err := os.ReadFile(filepath.Join(releaseJournalDir(), entry.Name()))
		if err != nil { return nil, fmt.Errorf("read release journal %q: %w", entry.Name(), err) }
		var candidate releaseJournal
		if err := json.Unmarshal(data, &candidate); err != nil {
			return nil, fmt.Errorf("invalid release journal %q: %w", entry.Name(), err)
		}
		if latest == nil || candidate.UpdatedAt.After(latest.UpdatedAt) { latest = &candidate }
	}
	if latest == nil { return nil, fmt.Errorf("no previous release journal found; run a release first") }
	return latest, nil
}

func validateRetryJournal(j *releaseJournal, version string, git releaseGitState, channelSuffix, directory, format, output string) error {
	if j == nil { return fmt.Errorf("no previous release journal found") }
	if version == "" || version != j.Version {
		return fmt.Errorf("retry version mismatch: previous release used %q; pass --userversion %s", j.Version, j.Version)
	}
	if git.Commit != j.Commit {
		return fmt.Errorf("retry commit mismatch: previous release used %s, current checkout is %s; check out the original commit before retrying", j.Commit, git.Commit)
	}
	if channelSuffix != j.ChannelSuffix || directory != j.ButlerDirectory || format != j.PackageFormat || output != j.PackageOutput {
		return fmt.Errorf("release settings changed since the previous attempt; restore the original channel, output directory, and package settings before retrying")
	}
	return nil
}

func releaseBuildTargetKey(name, platform, command string) string {
	sum := sha256.Sum256([]byte(name + "\x00" + platform + "\x00" + command))
	return hex.EncodeToString(sum[:])
}
