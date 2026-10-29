package scaffold

import (
	"encoding/json"
	"errors"
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
const maxTemplateFileBytes = 10 << 20
const maxTemplateFiles = 500

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

func templateCacheRoot() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("locating user cache: %w", err)
	}
	return filepath.Join(base, "fe", "templates"), nil
}

// ListRemoteTemplates returns the root-level template directories in fe-templates.
func ListRemoteTemplates() ([]string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, templateRepoAPI+"/git/trees/main?recursive=1", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("listing templates: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("listing templates: GitHub returned %s", resp.Status)
	}
	tree, err := decodeRepoTree(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decoding template list: %w", err)
	}
	if tree.Trunc {
		return nil, fmt.Errorf("template repository tree was truncated")
	}
	found := map[string]bool{}
	for _, item := range tree.Tree {
		parts := strings.Split(item.Path, "/")
		if len(parts) > 1 && templateIDPattern.MatchString(parts[0]) {
			found[parts[0]] = true
		}
	}
	ids := make([]string, 0, len(found))
	for id := range found {
		ids = append(ids, id)
	}
	sortStrings(ids)
	return ids, nil
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
	if !refresh {
		if info, err := os.Stat(filepath.Join(cacheRoot, id+".ready")); err == nil && !info.IsDir() {
			if dirInfo, dirErr := os.Stat(dest); dirErr == nil && dirInfo.IsDir() {
				return dest, nil
			}
		}
	}
	return downloadTemplate(id, cacheRoot, dest)
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
	_ = os.Remove(filepath.Join(cacheRoot, id+".ready"))
	if err := os.RemoveAll(dest); err != nil {
		return "", fmt.Errorf("replacing cached template: %w", err)
	}
	if err := os.Rename(tmp, dest); err != nil {
		return "", fmt.Errorf("saving template cache: %w", err)
	}
	if err := os.WriteFile(filepath.Join(cacheRoot, id+".ready"), []byte("flushwhy/fe-templates main\\n"), 0600); err != nil {
		return "", fmt.Errorf("writing template cache marker: %w", err)
	}
	return dest, nil
}

func sortStrings(values []string) { sort.Strings(values) }


// decodeRepoTree accepts exactly one complete JSON document and enforces the
// response limit, including trailing data.
func decodeRepoTree(body io.Reader) (repoTree, error) {
	var tree repoTree
	decoder := json.NewDecoder(io.LimitReader(body, (5<<20)+1))
	if err := decoder.Decode(&tree); err != nil {
		return repoTree{}, err
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return repoTree{}, fmt.Errorf("unexpected trailing JSON value")
		}
		return repoTree{}, fmt.Errorf("invalid trailing response data: %w", err)
	}
	return tree, nil
}
