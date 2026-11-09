package projectstate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ignoreStart = "# >>> Fe managed >>>"
	ignoreEnd   = "# <<< Fe managed <<<"
)

var lineBreak = string(rune(10))

type Record struct {
	Root             string            `json:"root"`
	CreatedAt        time.Time         `json:"created_at"`
	ManagedFiles     map[string]string `json:"managed_files"`
	GitignoreCreated bool              `json:"gitignore_created,omitempty"`
}

type Cleanup struct {
	Removed   []string
	Preserved []string
}

func Sync(root string) error {
	root, err := canonicalRoot(root)
	if err != nil {
		return err
	}
	record, err := loadRecord(root)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if errors.Is(err, os.ErrNotExist) {
		record = Record{Root: root, CreatedAt: time.Now().UTC(), ManagedFiles: map[string]string{}}
	}
	if record.ManagedFiles == nil {
		record.ManagedFiles = map[string]string{}
	}

	gitignore := filepath.Join(root, ".gitignore")
	info, statErr := os.Lstat(gitignore)
	if statErr == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to modify symlink %s", gitignore)
	}
	if errors.Is(statErr, os.ErrNotExist) {
		record.GitignoreCreated = true
	} else if statErr != nil {
		return fmt.Errorf("inspect .gitignore: %w", statErr)
	}
	if err := ensureGitignore(gitignore); err != nil {
		return err
	}

	for name, content := range uninstallers {
		path := filepath.Join(root, name)
		if err := writeManagedFile(path, content, record.ManagedFiles[name]); err != nil {
			return err
		}
		hash := sha256.Sum256([]byte(content))
		record.ManagedFiles[name] = hex.EncodeToString(hash[:])
	}
	return saveRecord(record)
}

func Clean(root string, dryRun, keepUninstallers bool) (Cleanup, error) {
	root, err := canonicalRoot(root)
	if err != nil {
		return Cleanup{}, err
	}
	record, err := loadRecord(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Cleanup{}, fmt.Errorf("no Fe project record found for %s; nothing is safe to remove", root)
		}
		return Cleanup{}, err
	}
	return cleanRecord(record, dryRun, keepUninstallers)
}

func CleanAll(dryRun bool) ([]Cleanup, error) {
	dir, err := recordsDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var results []Cleanup
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return results, err
		}
		var record Record
		if err := json.Unmarshal(data, &record); err != nil {
			return results, fmt.Errorf("read project record %s: %w", entry.Name(), err)
		}
		info, err := os.Lstat(record.Root)
		if errors.Is(err, os.ErrNotExist) {
			if !dryRun {
				if err := os.Remove(path); err != nil {
					return results, err
				}
			}
			results = append(results, Cleanup{Removed: []string{"stale project record: " + record.Root}})
			continue
		}
		if err != nil {
			return results, fmt.Errorf("inspect project %s: %w", record.Root, err)
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return results, fmt.Errorf("registered project path %s is no longer a real directory", record.Root)
		}
		result, err := cleanRecord(record, dryRun, false)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	if !dryRun {
		if err := removeEmptyStateDirs(); err != nil {
			return results, err
		}
	}
	return results, nil
}

func cleanRecord(record Record, dryRun, keepUninstallers bool) (Cleanup, error) {
	result := Cleanup{}
	for name, expectedHash := range record.ManagedFiles {
		if keepUninstallers {
			continue
		}
		path := filepath.Join(record.Root, name)
		data, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return result, fmt.Errorf("read managed file %s: %w", path, err)
		}
		hash := sha256.Sum256(data)
		if hex.EncodeToString(hash[:]) != expectedHash {
			result.Preserved = append(result.Preserved, path+" (modified since Fe created it)")
			continue
		}
		result.Removed = append(result.Removed, path)
		if !dryRun {
			if err := os.Remove(path); err != nil {
				return result, fmt.Errorf("remove managed file %s: %w", path, err)
			}
		}
	}

	gitignore := filepath.Join(record.Root, ".gitignore")
	info, statErr := os.Lstat(gitignore)
	if statErr == nil && info.Mode()&os.ModeSymlink != 0 {
		return result, fmt.Errorf("refusing to modify symlink %s", gitignore)
	}
	if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return result, fmt.Errorf("inspect .gitignore: %w", statErr)
	}
	data, err := os.ReadFile(gitignore)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return result, fmt.Errorf("read .gitignore: %w", err)
	}
	if err == nil {
		cleaned, changed := removeIgnoreBlock(string(data))
		if !changed && (strings.Contains(string(data), ignoreStart) || strings.Contains(string(data), ignoreEnd)) {
			return result, fmt.Errorf("incomplete Fe-managed block in %s; refusing cleanup", gitignore)
		}
		if changed {
			if record.GitignoreCreated && strings.TrimSpace(cleaned) == "" {
				result.Removed = append(result.Removed, gitignore)
				if !dryRun {
					if err := os.Remove(gitignore); err != nil {
						return result, fmt.Errorf("remove Fe-created .gitignore: %w", err)
					}
				}
			} else {
				result.Removed = append(result.Removed, "Fe entries from "+gitignore)
				if !dryRun {
					if err := os.WriteFile(gitignore, []byte(cleaned), 0o644); err != nil {
						return result, fmt.Errorf("restore .gitignore: %w", err)
					}
				}
			}
		}
	}

	if len(result.Preserved) == 0 && !dryRun {
		path, err := recordPath(record.Root)
		if err != nil {
			return result, err
		}
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return result, fmt.Errorf("remove project record: %w", err)
		}
		if err := removeEmptyStateDirs(); err != nil {
			return result, err
		}
	}
	return result, nil
}

func ensureGitignore(path string) error {
	info, err := os.Lstat(path)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to modify symlink %s", path)
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect .gitignore: %w", err)
	}
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read .gitignore: %w", err)
	}
	content := string(data)
	if strings.Contains(content, ignoreStart) || strings.Contains(content, ignoreEnd) {
		if strings.Contains(content, ignoreStart) && strings.Contains(content, ignoreEnd) {
			return nil
		}
		return fmt.Errorf("incomplete Fe-managed block in %s; refusing to modify it", path)
	}
	if content != "" && !strings.HasSuffix(content, lineBreak) {
		content += lineBreak
	}
	content += ignoreStart + lineBreak + ".fe/" + lineBreak + ignoreEnd + lineBreak
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("update .gitignore: %w", err)
	}
	return nil
}

func removeIgnoreBlock(content string) (string, bool) {
	start := strings.Index(content, ignoreStart)
	if start < 0 {
		return content, false
	}
	end := strings.Index(content[start:], ignoreEnd)
	if end < 0 {
		return content, false
	}
	end += start + len(ignoreEnd)
	if end < len(content) && content[end] == 10 {
		end++
	}
	return content[:start] + content[end:], true
}

func writeManagedFile(path, content, expectedHash string) error {
	info, err := os.Lstat(path)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing to replace symlink %s", path)
		}
		existing, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		hash := sha256.Sum256(existing)
		if expectedHash != "" && hex.EncodeToString(hash[:]) == expectedHash && string(existing) == content {
			return nil
		}
		return fmt.Errorf("%s already exists and is not Fe-managed; refusing to overwrite it", path)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	return nil
}

func canonicalRoot(root string) (string, error) {
	if root == "" {
		root = "."
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve project directory: %w", err)
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("resolve project directory %s: %w", abs, err)
	}
	info, err := os.Stat(real)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("project path %s is not a directory", real)
	}
	return real, nil
}

func loadRecord(root string) (Record, error) {
	path, err := recordPath(root)
	if err != nil {
		return Record{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Record{}, err
	}
	var record Record
	if err := json.Unmarshal(data, &record); err != nil {
		return Record{}, fmt.Errorf("decode project record: %w", err)
	}
	if record.Root != root {
		return Record{}, fmt.Errorf("project record path mismatch for %s", root)
	}
	return record, nil
}

func saveRecord(record Record) error {
	path, err := recordPath(record.Root)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create Fe state directory: %w", err)
	}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, 10)
	tmp, err := os.CreateTemp(filepath.Dir(path), ".project-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func recordPath(root string) (string, error) {
	dir, err := recordsDir()
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(root))
	return filepath.Join(dir, hex.EncodeToString(hash[:])+".json"), nil
}

func removeEmptyStateDirs() error {
	dir, err := recordsDir()
	if err != nil {
		return err
	}
	for _, path := range []string{dir, filepath.Dir(dir), filepath.Dir(filepath.Dir(dir))} {
		entries, err := os.ReadDir(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
	}
	return nil
}

func recordsDir() (string, error) {
	home := os.Getenv("FE_HOME")
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
		home = filepath.Join(home, ".fe")
	}
	return filepath.Join(home, "state", "projects"), nil
}

var uninstallers = map[string]string{
	"fe-uninstall.sh": "#!/bin/sh\nset -eu\nproject_dir=$(CDPATH= cd -- \"$(dirname -- \"$0\")\" && pwd)\nif ! command -v fe >/dev/null 2>&1; then\n  echo \"Fe must be installed and available on PATH to clean this project.\" >&2\n  exit 127\nfi\nfe clean \"$project_dir\" --keep-uninstallers\nrm -f \"$project_dir/fe-uninstall.sh\" \"$project_dir/fe-uninstall.bat\"\n",
	"fe-uninstall.bat": "@echo off\nwhere fe >nul 2>nul\nif errorlevel 1 (\n  echo Fe must be installed and available on PATH to clean this project.\n  exit /b 127\n)\nfor %%I in (\"%~dp0.\") do set \"FE_PROJECT=%%~fI\"\nfe clean \"%FE_PROJECT%\" --keep-uninstallers\nif errorlevel 1 exit /b %ERRORLEVEL%\ndel \"%~f0\"\ndel \"%FE_PROJECT%\\fe-uninstall.sh\"\nexit /b 0\n",
}
