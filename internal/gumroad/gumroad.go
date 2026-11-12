package gumroad

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type CommandExecutor func(name string, args ...string) *exec.Cmd

type Logger interface {
	Printf(format string, v ...interface{})
}

type artifactManifest struct {
	Version   string           `json:"version,omitempty"`
	Artifacts []artifactRecord `json:"artifacts"`
}

type artifactRecord struct {
	Path string `json:"path"`
}

type dryRunResult struct {
	Preserved []struct {
		ID string `json:"id"`
	} `json:"preserved"`
}

// Publish uploads a stable release to its explicitly configured, version-specific
// product. It refuses to append to a product that already contains files.
func Publish(version, directory string, products map[string]string, executor CommandExecutor, logger Logger) error {
	productID, files, args, err := preflight(version, directory, products, executor)
	if err != nil {
		return err
	}
	if executor == nil {
		executor = exec.Command
	}
	if logger == nil {
		logger = discardLogger{}
	}

	logger.Printf("Uploading %d artifact(s) to Gumroad product %s", len(files), productID)
	liveArgs := append(append([]string{}, args...), "--json", "--no-input")
	if _, err := run(executor, liveArgs...); err != nil {
		return fmt.Errorf("Gumroad upload for product %s returned an error; its state may be partial, inspect the product before retrying: %w", productID, err)
	}

	checkArgs := append(append([]string{}, args...), "--dry-run", "--json", "--no-input")
	output, err := run(executor, checkArgs...)
	if err != nil {
		return fmt.Errorf("Gumroad upload command completed, but product %s could not be verified: %w", productID, err)
	}
	var after dryRunResult
	if err := json.Unmarshal(output, &after); err != nil {
		return fmt.Errorf("Gumroad upload command completed, but the verification response was unreadable: %w", err)
	}
	if len(after.Preserved) != len(files) {
		return fmt.Errorf("Gumroad upload verification failed for product %s: expected %d file(s), found %d; inspect the product before retrying", productID, len(files), len(after.Preserved))
	}

	logger.Printf("Verified %d artifact(s) on Gumroad product %s", len(files), productID)
	return nil
}

// CheckProduct verifies that the selected stable product has no existing downloadable files.
// It performs only the Gumroad CLI's non-mutating update dry-run.
func CheckProduct(version, directory string, products map[string]string, executor CommandExecutor) error {
	_, _, _, err := preflight(version, directory, products, executor)
	return err
}

func preflight(version, directory string, products map[string]string, executor CommandExecutor) (string, []string, []string, error) {
	version = strings.TrimSpace(version)
	if version == "" {
		return "", nil, nil, fmt.Errorf("Gumroad stable release requires a version")
	}
	configuredProductID := products[version]
	productID := strings.TrimSpace(configuredProductID)
	if productID == "" {
		return "", nil, nil, fmt.Errorf("no Gumroad stable product configured for version %q (set gumroad.stable.%s in .fe.yaml)", version, version)
	}
	if configuredProductID != productID || strings.ContainsAny(productID, " \t\r\n") {
		return "", nil, nil, fmt.Errorf("Gumroad product ID for version %q is invalid", version)
	}
	files, err := releaseFiles(directory, version)
	if err != nil {
		return "", nil, nil, err
	}
	args := []string{"products", "update", productID}
	for _, path := range files {
		args = append(args, "--file", path)
	}
	if executor == nil {
		executor = exec.Command
	}
	checkArgs := append(append([]string{}, args...), "--dry-run", "--json", "--no-input")
	output, err := run(executor, checkArgs...)
	if err != nil {
		return "", nil, nil, fmt.Errorf("could not inspect Gumroad product %s before upload: %w", productID, err)
	}
	var before dryRunResult
	if err := json.Unmarshal(output, &before); err != nil {
		return "", nil, nil, fmt.Errorf("Gumroad CLI returned an unreadable product preflight response: %w", err)
	}
	if len(before.Preserved) != 0 {
		return "", nil, nil, fmt.Errorf("Gumroad product %s already contains %d file(s); refusing to append a stable release (use a new product ID with no existing files for version %s)", productID, len(before.Preserved), version)
	}
	return productID, files, args, nil
}

// ValidateRelease checks the local manifest and resolves the exact stable product
// without invoking Gumroad. It is used by dry-run planning.
func ValidateRelease(version, directory string, products map[string]string) (string, []string, error) {
	version = strings.TrimSpace(version)
	if version == "" {
		return "", nil, fmt.Errorf("Gumroad stable release requires a version")
	}
	configuredProductID := products[version]
	productID := strings.TrimSpace(configuredProductID)
	if productID == "" {
		return "", nil, fmt.Errorf("no Gumroad stable product configured for version %q (set gumroad.stable.%s in .fe.yaml)", version, version)
	}
	if configuredProductID != productID || strings.ContainsAny(productID, " \t\r\n") {
		return "", nil, fmt.Errorf("Gumroad product ID for version %q is invalid", version)
	}
	files, err := releaseFiles(directory, version)
	if err != nil {
		return "", nil, err
	}
	return productID, files, nil
}

func releaseFiles(directory, version string) ([]string, error) {
	manifestPath := filepath.Join(directory, "artifact-manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read release artifact manifest %q: %w", manifestPath, err)
	}
	var manifest artifactManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse release artifact manifest %q: %w", manifestPath, err)
	}
	if manifest.Version != version {
		return nil, fmt.Errorf("artifact manifest version %q does not match requested stable version %q", manifest.Version, version)
	}
	root, err := filepath.Abs(directory)
	if err != nil {
		return nil, fmt.Errorf("resolve artifact directory: %w", err)
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return nil, fmt.Errorf("resolve artifact directory symlinks: %w", err)
	}
	var files []string
	for _, artifact := range manifest.Artifacts {
		rel := filepath.FromSlash(artifact.Path)
		if rel == "" || filepath.IsAbs(rel) || rel == "." {
			return nil, fmt.Errorf("artifact manifest contains invalid path %q", artifact.Path)
		}
		clean := filepath.Clean(rel)
		if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("artifact manifest path %q escapes the artifact directory", artifact.Path)
		}
		path := filepath.Join(root, clean)
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			return nil, fmt.Errorf("resolve artifact %q: %w", artifact.Path, err)
		}
		relResolved, err := filepath.Rel(root, resolved)
		if err != nil || relResolved == ".." || strings.HasPrefix(relResolved, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("artifact %q resolves outside the artifact directory", artifact.Path)
		}
		info, err := os.Stat(resolved)
		if err != nil {
			return nil, fmt.Errorf("inspect artifact %q: %w", artifact.Path, err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("artifact %q is not a regular file", artifact.Path)
		}
		files = append(files, resolved)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("artifact manifest contains no files to upload")
	}
	return files, nil
}

func run(executor CommandExecutor, args ...string) ([]byte, error) {
	cmd := executor("gumroad", args...)
	if cmd == nil {
		return nil, fmt.Errorf("Gumroad command executor returned no command")
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("gumroad %s failed: %w\n%s", strings.Join(args[:min(len(args), 3)], " "), err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

type discardLogger struct{}

func (discardLogger) Printf(string, ...interface{}) {}
