package pack

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"github.com/flushwhy/fe/internal/config"
)

// Run packs all PNG files in cfg.Input into a single horizontal spritesheet
// written to cfg.Output.
func Run(cfg config.PackConfig) error {
	if cfg.Input == "" {
		return fmt.Errorf("input directory is required")
	}
	if cfg.Output == "" {
		return fmt.Errorf("output path is required")
	}

	// Ensure output directory exists.
	if err := os.MkdirAll(filepath.Dir(cfg.Output), os.ModePerm); err != nil {
		return fmt.Errorf("could not create output directory: %w", err)
	}

	images, err := loadPNGs(cfg.Input)
	if err != nil {
		return err
	}
	if len(images) == 0 {
		return fmt.Errorf("no PNG files found in %s", cfg.Input)
	}

	canvas := stitch(images)

	if err := writePNG(cfg.Output, canvas); err != nil {
		return err
	}

	fmt.Printf("✓ Packed %d images → %s\n", len(images), cfg.Output)
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

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create output file %s: %w", path, err)
	}
	defer f.Close()

	if err := png.Encode(f, canvas); err != nil {
		return fmt.Errorf("could not encode PNG: %w", err)
	}

	return nil
}
