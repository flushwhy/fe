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

	file, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("create package %q: %w", output, err)
	}
	defer file.Close()

	switch format {
	case "zip":
		writer := zip.NewWriter(file)
		walkErr := walkFiles(root, output, func(path string, info os.FileInfo) error {
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
		return errors.Join(walkErr, writer.Close())
	case "tar.gz":
		compressed := gzip.NewWriter(file)
		writer := tar.NewWriter(compressed)
		walkErr := walkFiles(root, output, func(path string, info os.FileInfo) error {
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			header.Name = filepath.ToSlash(path)
			if err := writer.WriteHeader(header); err != nil {
				return err
			}
			return copyFile(writer, filepath.Join(root, filepath.FromSlash(path)))
		})
		return errors.Join(walkErr, writer.Close(), compressed.Close())
	}
	return nil
}

func walkFiles(root, output string, visit func(string, os.FileInfo) error) error {
	count := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == root || path == output {
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
