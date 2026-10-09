// Package odinpkg downloads and installs Odin source repositories.
package odinpkg

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const maxArchiveSize = 50 << 20

var aliases = map[string]string{
	"ltdk": "https://github.com/flushwhy/ltdk.odin",
}

// Install resolves an alias or GitHub repository URL and installs it under destDir.
func Install(spec, destDir string) (string, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return "", fmt.Errorf("package name or repository URL is required")
	}
	if u, ok := aliases[strings.ToLower(spec)]; ok {
		spec = u
	}
	owner, repo, err := parseGitHubURL(spec)
	if err != nil {
		return "", err
	}
	name := strings.TrimSuffix(repo, ".odin")
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\\") {
		return "", fmt.Errorf("invalid package name derived from repository %q", repo)
	}
	if destDir == "" {
		destDir = "mylibs"
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("creating package directory: %w", err)
	}
	dest, err := filepath.Abs(filepath.Join(destDir, name))
	if err != nil {
		return "", err
	}
	if _, err := os.Lstat(dest); err == nil {
		return "", fmt.Errorf("destination %q already exists; refusing to overwrite", dest)
	} else if !os.IsNotExist(err) {
		return "", err
	}

	archiveURL := "https://api.github.com/repos/" + owner + "/" + repo + "/zipball"
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodGet, archiveURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("downloading %s: %w", spec, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub returned %s for %s", resp.Status, spec)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxArchiveSize+1))
	if err != nil {
		return "", fmt.Errorf("reading repository archive: %w", err)
	}
	if len(data) > maxArchiveSize {
		return "", fmt.Errorf("repository archive exceeds %d MiB limit", maxArchiveSize>>20)
	}
	zr, err := zip.NewReader(strings.NewReader(string(data)), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("reading repository archive: %w", err)
	}

	stage, err := os.MkdirTemp(destDir, "."+name+"-staging-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(stage)
	odinFound := false
	for _, f := range zr.File {
		clean := path.Clean(f.Name)
		parts := strings.Split(clean, "/")
		if len(parts) < 2 || clean == "." || strings.HasPrefix(clean, "../") || path.IsAbs(clean) {
			continue
		}
		rel := filepath.FromSlash(strings.Join(parts[1:], "/"))
		if rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		target := filepath.Join(stage, rel)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil { return "", err }
			continue
		}
		if strings.EqualFold(filepath.Ext(rel), ".odin") { odinFound = true }
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil { return "", err }
		in, err := f.Open()
		if err != nil { return "", err }
		out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil { in.Close(); return "", err }
		_, copyErr := io.Copy(out, io.LimitReader(in, maxArchiveSize+1))
		closeInErr, closeOutErr := in.Close(), out.Close()
		if copyErr != nil { return "", copyErr }
		if closeInErr != nil { return "", closeInErr }
		if closeOutErr != nil { return "", closeOutErr }
	}
	if !odinFound {
		return "", fmt.Errorf("repository %s contains no .odin files", spec)
	}
	if err := os.Rename(stage, dest); err != nil {
		return "", fmt.Errorf("installing package: %w", err)
	}
	return dest, nil
}

func parseGitHubURL(raw string) (owner, repo string, err error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || !strings.EqualFold(u.Hostname(), "github.com") {
		return "", "", fmt.Errorf("unsupported repository URL %q; use https://github.com/owner/repo", raw)
	}
	if u.User != nil || u.Port() != "" {
		return "", "", fmt.Errorf("invalid GitHub repository URL %q", raw)
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected a repository URL like https://github.com/owner/repo")
	}
	owner, repo = parts[0], strings.TrimSuffix(parts[1], ".git")
	if owner == "" || repo == "" || owner == "." || owner == ".." || repo == "." || repo == ".." {
		return "", "", fmt.Errorf("invalid GitHub repository URL %q", raw)
	}
	return owner, repo, nil
}
