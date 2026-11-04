package pack

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flushwhy/fe/internal/config"
)

const (
	maxPNGBytes       = 32 << 20
	maxPNGTotalBytes  = int64(64 << 20)
	maxSheetPixels    = int64(16 << 20)
	maxImageDimension = 16384
	maxSheetImages    = 4096
)

// Run packs PNG files into either one conventional spritesheet or a set of
// named sheets configured in cfg.Sheets.
func Run(cfg config.PackConfig) error {
	if len(cfg.Sheets) > 0 {
		if cfg.Input == "" {
			cfg.Input = "assets/sprites"
		}
		if cfg.Output == "" {
			cfg.Output = "assets/spritesheet.png"
		}
		names := make([]string, 0, len(cfg.Sheets))
		for name := range cfg.Sheets {
			if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\\") {
				return fmt.Errorf("invalid pack sheet name %q", name)
			}
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			sheet := cfg.Sheets[name]
			if sheet.Input == "" {
				sheet.Input = filepath.Join(cfg.Input, name)
			}
			if sheet.Output == "" {
				sheet.Output = filepath.Join(filepath.Dir(cfg.Output), name+".png")
			}
			if err := packOne(sheet.Input, sheet.Output); err != nil {
				return fmt.Errorf("packing sheet %q: %w", name, err)
			}
		}
		return nil
	}
	if cfg.Input == "" {
		return fmt.Errorf("input directory is required")
	}
	if cfg.Output == "" {
		return fmt.Errorf("output path is required")
	}
	return packOne(cfg.Input, cfg.Output)
}

func packOne(input, output string) error {
	if input == "" {
		return fmt.Errorf("input directory is required")
	}
	if output == "" {
		return fmt.Errorf("output path is required")
	}
	if err := os.MkdirAll(filepath.Dir(output), 0750); err != nil {
		return fmt.Errorf("could not create output directory: %w", err)
	}
	images, err := loadPNGs(input)
	if err != nil {
		return err
	}
	if len(images) == 0 {
		return fmt.Errorf("no PNG files found in %s", input)
	}
	if err := validateCanvas(images); err != nil {
		return err
	}
	canvas := stitch(images)
	if err := writePNG(output, canvas); err != nil {
		return err
	}
	fmt.Printf("✓ Packed %d images → %s\n", len(images), output)
	return nil
}

// loadPNGs reads every .png in dir and returns the decoded images.
func loadPNGs(dir string) ([]image.Image, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not read input directory %s: %w", dir, err)
	}

	var images []image.Image
	var totalPixels, totalEncodedBytes int64
	pngCount := 0

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".png" || entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		pngCount++
		if pngCount > maxSheetImages {
			return nil, fmt.Errorf("sprite sheet contains more than %d PNG files", maxSheetImages)
		}

		path := filepath.Join(dir, entry.Name())
		f, err := os.Open(path)
		if err != nil {
			fmt.Printf("  warning: could not open %s: %v\n", path, err)
			continue
		}
		info, err := f.Stat()
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("inspect PNG %s: %w", path, err)
		}
		if !info.Mode().IsRegular() {
			f.Close()
			continue
		}
		if err := validatePNGFileSize(info.Size(), totalEncodedBytes); err != nil {
			f.Close()
			return nil, fmt.Errorf("PNG %s: %w", path, err)
		}
		totalEncodedBytes += info.Size()

		cfg, err := png.DecodeConfig(f)
		if err != nil {
			f.Close()
			fmt.Printf("  warning: could not read PNG dimensions %s: %v\n", path, err)
			continue
		}
		pixels, err := validateImageSize(cfg.Width, cfg.Height, totalPixels)
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("PNG %s: %w", path, err)
		}
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			f.Close()
			return nil, fmt.Errorf("rewind PNG %s: %w", path, err)
		}
		img, err := png.Decode(f)
		closeErr := f.Close()
		if err != nil {
			fmt.Printf("  warning: could not decode %s: %v\n", path, err)
			continue
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close PNG %s: %w", path, closeErr)
		}
		totalPixels += pixels
		images = append(images, img)
	}
	return images, nil
}

func validatePNGFileSize(size, totalBytes int64) error {
	if size < 0 || size > maxPNGBytes {
		return fmt.Errorf("encoded size exceeds the %d MiB per-file limit", maxPNGBytes>>20)
	}
	if size > maxPNGTotalBytes-totalBytes {
		return fmt.Errorf("spritesheet exceeds the %d MiB total encoded-size limit", maxPNGTotalBytes>>20)
	}
	return nil
}

func validateImageSize(width, height int, totalPixels int64) (int64, error) {
	if width <= 0 || height <= 0 || width > maxImageDimension || height > maxImageDimension {
		return 0, fmt.Errorf("dimensions %dx%d exceed the %dx%d limit", width, height, maxImageDimension, maxImageDimension)
	}
	pixels := int64(width) * int64(height)
	if pixels > maxSheetPixels-totalPixels {
		return 0, fmt.Errorf("spritesheet exceeds the %d million decoded-pixel limit", maxSheetPixels>>20)
	}
	return pixels, nil
}

func validateCanvas(images []image.Image) error {
	var totalWidth, maxHeight int64
	for _, img := range images {
		bounds := img.Bounds()
		totalWidth += int64(bounds.Dx())
		if int64(bounds.Dy()) > maxHeight {
			maxHeight = int64(bounds.Dy())
		}
		if totalWidth > maxImageDimension || maxHeight > maxImageDimension {
			return fmt.Errorf("spritesheet dimensions exceed the %dx%d limit", maxImageDimension, maxImageDimension)
		}
	}
	if totalWidth == 0 || maxHeight == 0 || totalWidth*maxHeight > maxSheetPixels {
		return fmt.Errorf("spritesheet exceeds the %d million output-pixel limit", maxSheetPixels>>20)
	}
	return nil
}

// stitch combines images horizontally into a single RGBA canvas.
func stitch(images []image.Image) *image.RGBA {
	totalW, maxH := 0, 0
	for _, img := range images {
		b := img.Bounds()
		totalW += b.Dx()
		if b.Dy() > maxH {
			maxH = b.Dy()
		}
	}

	canvas := image.NewRGBA(image.Rect(0, 0, totalW, maxH))
	x := 0
	for _, img := range images {
		b := img.Bounds()
		dst := image.Rect(x, 0, x+b.Dx(), b.Dy())
		draw.Draw(canvas, dst, img, image.Point{}, draw.Src)
		x += b.Dx()
	}

	return canvas
}

// writePNG encodes canvas as a PNG file at path, replacing any existing file.
func writePNG(path string, canvas *image.RGBA) error {
	// Remove existing file first so we never serve a partial write.
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not remove existing output file: %w", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return fmt.Errorf("could not create output file %s: %w", path, err)
	}
	defer f.Close()

	if err := png.Encode(f, canvas); err != nil {
		return fmt.Errorf("could not encode PNG: %w", err)
	}

	return nil
}
