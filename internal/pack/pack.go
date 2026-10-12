package pack

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flushwhy/fe/internal/config"
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

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".png" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		f, err := os.Open(path)
		if err != nil {
			fmt.Printf("  warning: could not open %s: %v\n", path, err)
			continue
		}

		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			fmt.Printf("  warning: could not decode %s: %v\n", path, err)
			continue
		}

		images = append(images, img)
	}

	return images, nil
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
