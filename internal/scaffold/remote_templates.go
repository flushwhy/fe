package scaffold

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const templateRepoAPI = "https://api.github.com/repos/flushwhy/fe-templates"
const templateCatalogURL = "https://raw.githubusercontent.com/flushwhy/fe-templates/main/catalog.json"
const maxTemplateFileBytes = 10 << 20
const maxTemplateFiles = 500
const maxCatalogBytes = 1 << 20

var templateIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type repoTree struct {
	SHA   string     `json:"sha"`
	Tree  []treeItem `json:"tree"`
	Trunc bool       `json:"truncated"`
}

type treeItem struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type TemplateInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
}

type templateCatalog struct {
	Version   int            `json:"version"`
	Templates []TemplateInfo `json:"templates"`
}

func templateCacheRoot() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("locating user cache: %w", err)
	}
	return filepath.Join(base, "fe", "templates"), nil
}

// ListRemoteTemplates returns catalog metadata. The last successful catalog is
// retained locally so browsing templates also works without a network connection.
func ListRemoteTemplates() ([]TemplateInfo, error) {
	cacheRoot, err := templateCacheRoot()
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, templateCatalogURL, nil)
	if err != nil {
		return cachedTemplateCatalog(cacheRoot, err)
	}
	req.Header.Set("Accept", "application/json")
	resp, requestErr := client.Do(req)
	if requestErr == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			data, readErr := io.ReadAll(io.LimitReader(resp.Body, maxCatalogBytes+1))
			if readErr == nil && len(data) <= maxCatalogBytes {
				catalog, parseErr := decodeTemplateCatalog(data)
				if parseErr == nil {
					if err := saveTemplateCatalog(cacheRoot, data); err != nil {
						return catalog.Templates, fmt.Errorf("loaded template catalog but could not cache it: %w", err)
					}
					return catalog.Templates, nil
				}
				requestErr = parseErr
			} else if readErr != nil {
				requestErr = readErr
			} else {
				requestErr = fmt.Errorf("template catalog exceeds 1 MiB limit")
			}
		} else {
			requestErr = fmt.Errorf("GitHub returned %s", resp.Status)
		}
	}
	return cachedTemplateCatalog(cacheRoot, requestErr)
}

func decodeTemplateCatalog(data []byte) (templateCatalog, error) {
	var catalog templateCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return catalog, fmt.Errorf("decoding template catalog: %w", err)
	}
	if catalog.Version != 1 || len(catalog.Templates) == 0 {
		return catalog, fmt.Errorf("unsupported or empty template catalog (version %d)", catalog.Version)
	}
	seen := make(map[string]bool, len(catalog.Templates))
	for _, item := range catalog.Templates {
		if !templateIDPattern.MatchString(item.ID) || strings.TrimSpace(item.Name) == "" || strings.TrimSpace(item.Description) == "" || strings.TrimSpace(item.Category) == "" {
			return catalog, fmt.Errorf("invalid metadata for template %q", item.ID)
		}
		if seen[item.ID] {
			return catalog, fmt.Errorf("duplicate template ID %q in catalog", item.ID)
		}
		seen[item.ID] = true
	}
	sort.Slice(catalog.Templates, func(i, j int) bool {
		if catalog.Templates[i].Category == catalog.Templates[j].Category {
			return catalog.Templates[i].Name < catalog.Templates[j].Name
		}
		return catalog.Templates[i].Category < catalog.Templates[j].Category
	})
	return catalog, nil
}

func cachedTemplateCatalog(cacheRoot string, cause error) ([]TemplateInfo, error) {
	data, err := os.ReadFile(filepath.Join(cacheRoot, "catalog.json"))
	if err == nil {
		catalog, parseErr := decodeTemplateCatalog(data)
		if parseErr == nil {
			return catalog.Templates, nil
		}
		err = parseErr
	}
	if cause != nil {
		return nil, fmt.Errorf("could not load template catalog (%v); no usable offline catalog is cached: %w", cause, err)
	}
	return nil, fmt.Errorf("no usable offline template catalog is cached: %w", err)
}

func saveTemplateCatalog(cacheRoot string, data []byte) error {
	if err := os.MkdirAll(cacheRoot, 0750); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(cacheRoot, ".catalog-")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	dest := filepath.Join(cacheRoot, "catalog.json")
	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Rename(name, dest)
}

// ResolveTemplate returns a cached template, downloading it on first use.
// The ID is always a single safe directory name, never an arbitrary URL or path.
func ResolveTemplate(id string, refresh bool) (string, error) {
	if !templateIDPattern.MatchString(id) {
		return "", fmt.Errorf("invalid template ID %q (use lowercase words separated by hyphens)", id)
	}
	cacheRoot, err := templateCacheRoot()
	if err != nil {
		return "", err
	}
	dest := filepath.Join(cacheRoot, id)
	if !refresh && cachedTemplateIsSafe(cacheRoot, id) {
		return dest, nil
	}
	return downloadTemplate(id, cacheRoot, dest)
}

// cachedTemplateIsSafe rejects symlinks so a poisoned cache entry cannot redirect
// template reads outside the cache directory.
func cachedTemplateIsSafe(cacheRoot, id string) bool {
	marker, err := os.Lstat(filepath.Join(cacheRoot, id+".ready"))
	if err != nil || !marker.Mode().IsRegular() {
		return false
	}
	dir, err := os.Lstat(filepath.Join(cacheRoot, id))
	return err == nil && dir.IsDir() && dir.Mode()&os.ModeSymlink == 0
}

func downloadTemplate(id, cacheRoot, dest string) (string, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequest(http.MethodGet, templateRepoAPI+"/git/trees/main?recursive=1", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching template catalog: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetching template catalog: GitHub returned %s", resp.Status)
	}
	tree, err := decodeRepoTree(resp.Body)
	if err != nil {
		return "", fmt.Errorf("decoding template catalog: %w", err)
	}
	if tree.Trunc {
		return "", fmt.Errorf("template repository tree was truncated")
	}
	prefix := id + "/"
	var files []string
	for _, item := range tree.Tree {
		if item.Type == "blob" && strings.HasPrefix(item.Path, prefix) {
			rel := strings.TrimPrefix(item.Path, prefix)
			if rel == "" || path.IsAbs(rel) || path.Clean(rel) != rel || rel == ".." || strings.HasPrefix(rel, "../") || strings.Contains(rel, "\\") {
				return "", fmt.Errorf("unsafe path in template repository: %q", item.Path)
			}
			files = append(files, rel)
		}
	}
	if len(files) == 0 {
		return "", fmt.Errorf("template %q was not found in https://github.com/flushwhy/fe-templates", id)
	}
	if len(files) > maxTemplateFiles {
		return "", fmt.Errorf("template %q has too many files (%d)", id, len(files))
	}
	sort.Strings(files)

	if err := os.MkdirAll(cacheRoot, 0750); err != nil {
		return "", fmt.Errorf("creating template cache: %w", err)
	}
	tmp, err := os.MkdirTemp(cacheRoot, "."+id+"-")
	if err != nil {
		return "", fmt.Errorf("creating temporary template cache: %w", err)
	}
	defer os.RemoveAll(tmp)

	var total int64
	for _, rel := range files {
		fileURL := "https://raw.githubusercontent.com/flushwhy/fe-templates/main/" + id + "/" + rel
		fileReq, err := http.NewRequest(http.MethodGet, fileURL, nil)
		if err != nil {
			return "", err
		}
		fileResp, err := client.Do(fileReq)
		if err != nil {
			return "", fmt.Errorf("downloading template file %s: %w", rel, err)
		}
		if fileResp.StatusCode != http.StatusOK {
			fileResp.Body.Close()
			return "", fmt.Errorf("downloading template file %s: GitHub returned %s", rel, fileResp.Status)
		}
		data, readErr := io.ReadAll(io.LimitReader(fileResp.Body, maxTemplateFileBytes+1))
		fileResp.Body.Close()
		if readErr != nil {
			return "", fmt.Errorf("reading template file %s: %w", rel, readErr)
		}
		if len(data) > maxTemplateFileBytes {
			return "", fmt.Errorf("template file %s exceeds the 10 MiB limit", rel)
		}
		total += int64(len(data))
		if total > 50<<20 {
			return "", fmt.Errorf("template %q exceeds the 50 MiB total size limit", id)
		}
		target := filepath.Join(tmp, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(target), 0750); err != nil {
			return "", fmt.Errorf("creating cache directory for %s: %w", rel, err)
		}
		if err := os.WriteFile(target, data, 0640); err != nil {
			return "", fmt.Errorf("caching template file %s: %w", rel, err)
		}
	}
	if err := replaceTemplateCache(cacheRoot, id, tmp, dest); err != nil {
		return "", err
	}
	return dest, nil
}

// replaceTemplateCache keeps the last usable cache until the replacement is complete.
func replaceTemplateCache(cacheRoot, id, tmp, dest string) error {
	marker := filepath.Join(cacheRoot, id+".ready")
	markerBackup := marker + ".old"
	dirBackup := dest + ".old"
	_ = os.Remove(markerBackup)
	_ = os.RemoveAll(dirBackup)

	markerTmp, err := os.CreateTemp(cacheRoot, "."+id+"-ready-")
	if err != nil {
		return fmt.Errorf("creating template cache marker: %w", err)
	}
	markerTmpName := markerTmp.Name()
	if _, err := markerTmp.WriteString("flushwhy/fe-templates main\n"); err != nil {
		markerTmp.Close()
		os.Remove(markerTmpName)
		return fmt.Errorf("writing template cache marker: %w", err)
	}
	if err := markerTmp.Close(); err != nil {
		os.Remove(markerTmpName)
		return fmt.Errorf("closing template cache marker: %w", err)
	}
	defer os.Remove(markerTmpName)

	hadMarker, hadDir := false, false
	if _, err := os.Lstat(marker); err == nil {
		if err := os.Rename(marker, markerBackup); err != nil {
			return fmt.Errorf("preserving template cache marker: %w", err)
		}
		hadMarker = true
	}
	if _, err := os.Lstat(dest); err == nil {
		if err := os.Rename(dest, dirBackup); err != nil {
			if hadMarker { _ = os.Rename(markerBackup, marker) }
			return fmt.Errorf("preserving existing template cache: %w", err)
		}
		hadDir = true
	}
	restore := func() {
		_ = os.RemoveAll(dest)
		if hadDir { _ = os.Rename(dirBackup, dest) }
		_ = os.Remove(marker)
		if hadMarker { _ = os.Rename(markerBackup, marker) }
	}
	if err := os.Rename(tmp, dest); err != nil {
		restore()
		return fmt.Errorf("saving template cache: %w", err)
	}
	if err := os.Rename(markerTmpName, marker); err != nil {
		restore()
		return fmt.Errorf("publishing template cache marker: %w", err)
	}
	if hadDir { _ = os.RemoveAll(dirBackup) }
	if hadMarker { _ = os.Remove(markerBackup) }
	return nil
}

func decodeRepoTree(body io.Reader) (repoTree, error) {
	var tree repoTree
	data, err := io.ReadAll(io.LimitReader(body, (5<<20)+1))
	if err != nil {
		return tree, err
	}
	if len(data) > 5<<20 {
		return tree, fmt.Errorf("template repository catalog exceeds 5 MiB limit")
	}
	if err := json.Unmarshal(data, &tree); err != nil {
		return tree, err
	}
	return tree, nil
}
