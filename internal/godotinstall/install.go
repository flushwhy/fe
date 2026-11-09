package godotinstall

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/flushwhy/fe/internal/projectstate"
)

const (
	releaseAPI = "https://api.github.com/repos/godotengine/godot-builds/releases/tags/"
	maxArchive = 250 << 20
	maxExtract = 500 << 20
)

type release struct {
	Assets []struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func Install(version string) (string, string, error) {
	tag, major, err := releaseTag(version)
	if err != nil {
		return "", "", err
	}
	assetName, platform, err := assetFor(tag, major, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return "", "", err
	}

	cache, err := projectstate.ToolCacheDir()
	if err != nil {
		return "", "", err
	}
	toolName := "godot:" + tag
	target := filepath.Join(cache, "godot", tag, platform)
	if executable, err := existingInstall(target, toolName, runtime.GOOS); err == nil {
		return executable, target, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", "", err
	}

	client := &http.Client{Timeout: 2 * time.Minute}
	req, err := http.NewRequest(http.MethodGet, releaseAPI+url.PathEscape(tag), nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Fe")
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("look up Godot %s release: %w", tag, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("look up Godot %s release: GitHub returned %s", tag, resp.Status)
	}
	var metadata release
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&metadata); err != nil {
		return "", "", fmt.Errorf("decode Godot release metadata: %w", err)
	}
	var asset *struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
		URL  string `json:"browser_download_url"`
	}
	for i := range metadata.Assets {
		if metadata.Assets[i].Name == assetName {
			asset = &metadata.Assets[i]
			break
		}
	}
	if asset == nil {
		return "", "", fmt.Errorf("Godot %s does not publish a %s build for %s/%s", tag, assetName, runtime.GOOS, runtime.GOARCH)
	}
	if asset.Size <= 0 || asset.Size > maxArchive {
		return "", "", fmt.Errorf("Godot archive size %d is outside the allowed limit", asset.Size)
	}
	downloadURL, err := url.Parse(asset.URL)
	if err != nil || downloadURL.Scheme != "https" || downloadURL.Host != "github.com" {
		return "", "", errors.New("Godot release returned an invalid download URL")
	}

	parent := filepath.Dir(target)
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return "", "", fmt.Errorf("create Fe tool cache: %w", err)
	}
	staging, err := os.MkdirTemp(parent, ".install-*")
	if err != nil {
		return "", "", err
	}
	defer os.RemoveAll(staging)

	download, err := client.Get(asset.URL)
	if err != nil {
		return "", "", fmt.Errorf("download Godot %s: %w", tag, err)
	}
	defer download.Body.Close()
	if download.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("download Godot %s: %s", tag, download.Status)
	}
	archivePath := filepath.Join(staging, "godot.zip")
	file, err := os.OpenFile(archivePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", "", err
	}
	written, copyErr := io.Copy(file, io.LimitReader(download.Body, maxArchive+1))
	closeErr := file.Close()
	if copyErr != nil {
		return "", "", copyErr
	}
	if closeErr != nil {
		return "", "", closeErr
	}
	if written != asset.Size || written > maxArchive {
		return "", "", fmt.Errorf("Godot download size mismatch: received %d bytes, expected %d", written, asset.Size)
	}

	extractDir := filepath.Join(staging, "engine")
	if err := extractArchive(archivePath, extractDir); err != nil {
		return "", "", fmt.Errorf("extract Godot %s: %w", tag, err)
	}
	executable, err := findExecutable(extractDir, runtime.GOOS)
	if err != nil {
		return "", "", err
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(executable, 0o755); err != nil {
			return "", "", fmt.Errorf("make Godot executable: %w", err)
		}
	}
	if err := os.WriteFile(filepath.Join(extractDir, ".fe-managed"), []byte(toolName+"
"), 0o600); err != nil {
		return "", "", err
	}
	if _, err := os.Lstat(target); err == nil {
		return "", "", fmt.Errorf("tool cache path already exists but is not a valid Fe installation: %s", target)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", "", err
	}
	if err := os.Rename(extractDir, target); err != nil {
		return "", "", fmt.Errorf("finish Godot installation: %w", err)
	}
	executable, err = findExecutable(target, runtime.GOOS)
	if err != nil {
		return "", "", err
	}
	if err := writeLauncher(target, executable, runtime.GOOS); err != nil {
		return "", "", err
	}
	return executable, target, nil
}

func releaseTag(version string) (string, int, error) {
	version = strings.TrimSpace(strings.TrimSuffix(version, "-stable"))
	parts := strings.Split(version, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return "", 0, fmt.Errorf("invalid Godot version %q; use a stable version such as 4.3 or 3.5.3", version)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil || (major != 3 && major != 4) {
		return "", 0, fmt.Errorf("unsupported Godot major version in %q; only Godot 3 and 4 stable builds are supported", version)
	}
	for _, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return "", 0, fmt.Errorf("invalid Godot version %q", version)
		}
	}
	return version + "-stable", major, nil
}

func assetFor(tag string, major int, goos, goarch string) (string, string, error) {
	version := strings.TrimSuffix(tag, "-stable")
	prefix := "Godot_v" + version + "-stable_"
	switch {
	case goos == "windows" && goarch == "amd64":
		if major == 3 {
			return prefix + "win64.exe.zip", "windows-x86_64", nil
		}
		return prefix + "win64.exe.zip", "windows-x86_64", nil
	case goos == "windows" && goarch == "arm64" && major == 4:
		return prefix + "windows_arm64.exe.zip", "windows-arm64", nil
	case goos == "darwin" && (goarch == "amd64" || goarch == "arm64"):
		if major == 3 {
			return prefix + "osx.universal.zip", "macos-universal", nil
		}
		return prefix + "macos.universal.zip", "macos-universal", nil
	case goos == "linux" && goarch == "amd64":
		if major == 3 {
			return prefix + "x11.64.zip", "linux-x86_64", nil
		}
		return prefix + "linux.x86_64.zip", "linux-x86_64", nil
	case goos == "linux" && goarch == "arm64" && major == 4:
		return prefix + "linux.arm64.zip", "linux-arm64", nil
	default:
		return "", "", fmt.Errorf("local Godot installation is not supported on %s/%s", goos, goarch)
	}
}

func existingInstall(path, name, goos string) (string, error) {
	marker, err := os.ReadFile(filepath.Join(path, ".fe-managed"))
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(string(marker)) != name {
		return "", fmt.Errorf("tool cache path %s has an invalid ownership marker", path)
	}
	executable, err := findExecutable(path, goos)
	if err != nil {
		return "", err
	}
	if err := writeLauncher(path, executable, goos); err != nil {
		return "", err
	}
	return executable, nil
}

func writeLauncher(root, executable, goos string) error {
	bin := filepath.Join(root, "bin")
	if err := os.MkdirAll(bin, 0o700); err != nil {
		return err
	}
	relative, err := filepath.Rel(bin, executable)
	if err != nil {
		return err
	}
	var name, content string
	var mode os.FileMode = 0o700
	if goos == "windows" {
		name = "godot.cmd"
		relative = strings.ReplaceAll(relative, "/", "\\")
		content = "@echo off\r\n\"%~dp0" + relative + "\" %*\r\n"
		mode = 0o600
	} else {
		name = "godot"
		content = "#!/bin/sh\nexec \"$(dirname \"$0\")\"/" + shellQuote(filepath.ToSlash(relative)) + " \"$@\"\n"
	}
	if err := os.WriteFile(filepath.Join(bin, name), []byte(content), mode); err != nil {
		return fmt.Errorf("create project Godot launcher: %w", err)
	}
	return nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func extractArchive(archivePath, destination string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()
	if len(reader.File) == 0 || len(reader.File) > 5000 {
		return errors.New("archive contains an invalid number of files")
	}
	if err := os.MkdirAll(destination, 0o700); err != nil {
		return err
	}
	var total uint64
	for _, entry := range reader.File {
		name := filepath.Clean(filepath.FromSlash(entry.Name))
		if filepath.IsAbs(name) || name == "." || name == ".." || strings.HasPrefix(name, ".."+string(filepath.Separator)) {
			return fmt.Errorf("archive contains unsafe path %q", entry.Name)
		}
		mode := entry.Mode()
		if mode&os.ModeSymlink != 0 || (mode.Type() != 0 && !mode.IsRegular() && !mode.IsDir()) {
			return fmt.Errorf("archive contains unsupported file type: %s", entry.Name)
		}
		target := filepath.Join(destination, name)
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o700); err != nil {
				return err
			}
			continue
		}
		total += entry.UncompressedSize64
		if entry.UncompressedSize64 > maxExtract || total > maxExtract {
			return errors.New("archive expands beyond the allowed size")
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		source, err := entry.Open()
		if err != nil {
			return err
		}
		output, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			source.Close()
			return err
		}
		_, copyErr := io.Copy(output, io.LimitReader(source, int64(entry.UncompressedSize64)+1))
		sourceErr := source.Close()
		outputErr := output.Close()
		if copyErr != nil {
			return copyErr
		}
		if sourceErr != nil {
			return sourceErr
		}
		if outputErr != nil {
			return outputErr
		}
	}
	return nil
}

func findExecutable(root, goos string) (string, error) {
	var matches []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		name := entry.Name()
		match := false
		switch goos {
		case "windows":
			match = strings.HasPrefix(name, "Godot") && strings.HasSuffix(strings.ToLower(name), ".exe")
		case "darwin":
			match = name == "Godot" && strings.Contains(path, ".app"+string(filepath.Separator)+"Contents"+string(filepath.Separator)+"MacOS")
		default:
			match = strings.HasPrefix(name, "Godot_v")
		}
		if match {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("expected one Godot executable in %s, found %d", root, len(matches))
	}
	return matches[0], nil
}
