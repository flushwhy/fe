package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/flushwhy/fe/internal/config"
	"github.com/spf13/cobra"
)

type artifactManifest struct {
	Version     string           `json:"version,omitempty"`
	Commit      string           `json:"commit,omitempty"`
	GeneratedAt time.Time        `json:"generated_at"`
	Artifacts   []artifactRecord `json:"artifacts"`
}

type artifactRecord struct {
	Platform string `json:"platform"`
	Path     string `json:"path"`
	Size     int64  `json:"size_bytes"`
	SHA256   string `json:"sha256"`
}

var manifestCmd = &cobra.Command{
	Use:   "manifest",
	Short: "Write a JSON inventory of build artifacts with SHA-256 checksums.",
	Long:  "Scans a build output directory and records each file's relative path, platform folder, size, and SHA-256 checksum.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = cfg.Butler.Directory
		}
		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			output = filepath.Join(dir, "artifact-manifest.json")
		}
		version, _ := cmd.Flags().GetString("version")
		if version == "" {
			version = cfg.Butler.Userversion
		}
		commit := os.Getenv("GITHUB_SHA")
		if commit == "" {
			if result, err := exec.Command("git", "rev-parse", "HEAD").Output(); err == nil {
				commit = strings.TrimSpace(string(result))
			}
		}
		return writeArtifactManifest(dir, output, version, commit)
	},
}

func init() {
	rootCmd.AddCommand(manifestCmd)
	manifestCmd.Flags().String("dir", "", "artifact directory (defaults to butler.directory)")
	manifestCmd.Flags().String("output", "", "manifest path (defaults to <dir>/artifact-manifest.json)")
	manifestCmd.Flags().String("version", "", "artifact version (defaults to butler.userversion)")
}

func writeArtifactManifest(root, output, version, commit string) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve artifact directory: %w", err)
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return fmt.Errorf("resolve artifact directory symlinks: %w", err)
	}
	output, err = filepath.Abs(output)
	if err != nil {
		return fmt.Errorf("resolve manifest path: %w", err)
	}
	if outputInfo, statErr := os.Lstat(output); statErr == nil && outputInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("manifest output %q must not be a symlink", output)
	} else if statErr != nil && !os.IsNotExist(statErr) {
		return fmt.Errorf("inspect manifest output %q: %w", output, statErr)
	}
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("inspect artifact directory %q: %w", root, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("artifact path %q is not a directory", root)
	}

	manifest := artifactManifest{Version: version, Commit: commit, GeneratedAt: time.Now().UTC()}
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		return addArtifact(root, output, path, entry, &manifest)
	})
	if err != nil {
		return fmt.Errorf("scan artifacts: %w", err)
	}
	sort.Slice(manifest.Artifacts, func(i, j int) bool {
		return manifest.Artifacts[i].Path < manifest.Artifacts[j].Path
	})
	if len(manifest.Artifacts) == 0 {
		return fmt.Errorf("no artifact files found in %q", root)
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return fmt.Errorf("create manifest directory: %w", err)
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := writeManifestAtomically(output, data); err != nil {
		return fmt.Errorf("write manifest %q: %w", output, err)
	}
	fmt.Printf("Wrote %d artifact(s) to %s\n", len(manifest.Artifacts), output)
	return nil
}


func addArtifact(root, output, path string, entry os.DirEntry, manifest *artifactManifest) error {
	if path == output || entry.Type()&os.ModeSymlink != 0 {
		return nil
	}
	info, err := entry.Info()
	if err != nil {
		return fmt.Errorf("inspect artifact %q: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open artifact %q: %w", path, err)
	}
	hash := sha256.New()
	size, copyErr := io.Copy(hash, file)
	closeErr := file.Close()
	if copyErr != nil {
		return fmt.Errorf("hash artifact %q: %w", path, copyErr)
	}
	if closeErr != nil {
		return closeErr
	}

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return err
	}
	relativePath := filepath.ToSlash(rel)
	platform := "root"
	if parts := strings.Split(relativePath, "/"); len(parts) > 1 {
		platform = parts[0]
	}
	manifest.Artifacts = append(manifest.Artifacts, artifactRecord{
		Platform: platform,
		Path:     relativePath,
		Size:     size,
		SHA256:   hex.EncodeToString(hash.Sum(nil)),
	})
	return nil
}

 // writeManifestAtomically avoids truncating an existing file before the complete
 // manifest is ready. Rename replaces the destination entry rather than following
 // a symlink that appears after the initial validation.
 func writeManifestAtomically(output string, data []byte) error {
 	dir := filepath.Dir(output)
 	tmp, err := os.CreateTemp(dir, ".artifact-manifest-*")
 	if err != nil {
 		return err
 	}
 	tmpName := tmp.Name()
 	defer os.Remove(tmpName)

 	if err := tmp.Chmod(0o644); err != nil {
 		_ = tmp.Close()
 		return err
 	}
 	if _, err := tmp.Write(data); err != nil {
 		_ = tmp.Close()
 		return err
 	}
 	if err := tmp.Sync(); err != nil {
 		_ = tmp.Close()
 		return err
 	}
 	if err := tmp.Close(); err != nil {
 		return err
 	}
 	return os.Rename(tmpName, output)
 }
