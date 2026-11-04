package archivepkg

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/flushwhy/fe/internal/archivepath"
)

func Create(root, output, format string) error {
	if format != "zip" && format != "tar.gz" {
		return fmt.Errorf("unsupported archive format %q (use zip or tar.gz)", format)
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve package directory: %w", err)
	}
	output, err = filepath.Abs(output)
	if err != nil {
		return fmt.Errorf("resolve package output: %w", err)
	}
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("read package directory %q: %w", root, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("package source %q is not a directory", root)
	}
	if info, err := os.Lstat(output); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("package output %q is a symlink", output)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("package output %q is not a regular file", output)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect package output %q: %w", output, err)
	}

	file, err := os.CreateTemp(filepath.Dir(output), ".fe-package-*")
	if err != nil {
		return fmt.Errorf("create temporary package in %q: %w", filepath.Dir(output), err)
	}
	temp := file.Name()
	defer os.Remove(temp)

	var writeErr error
	switch format {
	case "zip":
		writer := zip.NewWriter(file)
		walkErr := walkFiles(root, output, temp, func(path string, info os.FileInfo) error {
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			header.Name = path
			header.Method = zip.Deflate
			target, err := writer.CreateHeader(header)
			if err != nil {
				return err
			}
			return copyFile(target, filepath.Join(root, filepath.FromSlash(path)))
		})
		writeErr = errors.Join(walkErr, writer.Close())
	case "tar.gz":
		compressed := gzip.NewWriter(file)
		writer := tar.NewWriter(compressed)
		walkErr := walkFiles(root, output, temp, func(path string, info os.FileInfo) error {
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			header.Name = path
			if err := writer.WriteHeader(header); err != nil {
				return err
			}
			return copyFile(writer, filepath.Join(root, filepath.FromSlash(path)))
		})
		writeErr = errors.Join(walkErr, writer.Close(), compressed.Close())
	}
	if err := file.Close(); err != nil {
		writeErr = errors.Join(writeErr, err)
	}
	if writeErr != nil {
		return writeErr
	}
	if err := os.Chmod(temp, 0644); err != nil {
		return fmt.Errorf("set package permissions: %w", err)
	}
	// Recheck immediately before replacement. Removing the destination unlinks
	// a symlink itself; it never opens or truncates the symlink target.
	if info, err := os.Lstat(output); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("package output %q became a symlink", output)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("package output %q is not a regular file", output)
		}
		if err := os.Remove(output); err != nil {
			return fmt.Errorf("replace package %q: %w", output, err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect package output %q: %w", output, err)
	}
	if err := os.Rename(temp, output); err != nil {
		return fmt.Errorf("publish package %q: %w", output, err)
	}
	return nil
}

func walkFiles(root, output, temp string, visit func(string, os.FileInfo) error) error {
	count := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == root || path == output || path == temp {
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(relative)
		if err := archivepath.Validate(name); err != nil {
			return fmt.Errorf("unsafe archive entry %q: %w", name, err)
		}
		if err := visit(name, info); err != nil {
			return err
		}
		count++
		return nil
	})
	if err != nil {
		return fmt.Errorf("add files to package: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("package directory %q contains no regular files", root)
	}
	return nil
}

func copyFile(dst io.Writer, path string) error {
	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer src.Close()
	_, err = io.Copy(dst, src)
	return err
}

func FormatExtension(format string) string {
	if strings.EqualFold(format, "tar.gz") {
		return ".tar.gz"
	}
	return "." + format
}
