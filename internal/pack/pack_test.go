package pack

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/flushwhy/fe/internal/config"
)

func writeTestPNG(t *testing.T, path string, width, height int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	img.Set(0, 0, color.White)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRunNamedSheets(t *testing.T) {
	root := t.TempDir()
	writeTestPNG(t, filepath.Join(root, "sprites", "player", "idle.png"), 2, 3)
	writeTestPNG(t, filepath.Join(root, "sprites", "player", "run.png"), 4, 3)
	writeTestPNG(t, filepath.Join(root, "sprites", "npcs", "villager.png"), 5, 2)

	cfg := config.PackConfig{
		Input:  filepath.Join(root, "sprites"),
		Output: filepath.Join(root, "assets", "spritesheet.png"),
		Sheets: map[string]config.PackSheet{
			"player": {Input: filepath.Join(root, "sprites", "player"), Output: filepath.Join(root, "assets", "player.png")},
			"npcs":   {Input: filepath.Join(root, "sprites", "npcs"), Output: filepath.Join(root, "assets", "npcs.png")},
		},
	}
	if err := Run(cfg); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Open outputs relative to the temporary root instead of passing fully
	// constructed filesystem paths to os.Open.
	outputRoot, err := os.OpenRoot(root)
	if err != nil {
		t.Fatalf("opening output root: %v", err)
	}
	defer outputRoot.Close()

	for _, tc := range []struct {
		name string
		w, h int
	}{
		{"assets/player.png", 6, 3},
		{"assets/npcs.png", 5, 2},
	} {
		f, err := outputRoot.Open(tc.name)
		if err != nil {
			t.Fatalf("opening output %s: %v", tc.name, err)
		}
		img, err := png.Decode(f)
		f.Close()
		if err != nil {
			t.Fatalf("decoding output %s: %v", tc.name, err)
		}
		if got := img.Bounds().Dx(); got != tc.w {
			t.Errorf("%s width = %d, want %d", tc.name, got, tc.w)
		}
		if got := img.Bounds().Dy(); got != tc.h {
			t.Errorf("%s height = %d, want %d", tc.name, got, tc.h)
		}
	}
}

func TestRunNamedSheetsRejectsUnsafeName(t *testing.T) {
	err := Run(config.PackConfig{
		Sheets: map[string]config.PackSheet{"../escape": {}},
	})
	if err == nil {
		t.Fatal("Run() error = nil, want invalid sheet name error")
	}
}


type boundsOnlyImage struct{ bounds image.Rectangle }

func (img boundsOnlyImage) ColorModel() color.Model { return color.RGBAModel }
func (img boundsOnlyImage) Bounds() image.Rectangle { return img.bounds }
func (img boundsOnlyImage) At(x, y int) color.Color { return color.RGBA{} }

func TestValidateImageSizeEnforcesDimensionAndAggregateBudgets(t *testing.T) {
	if _, err := validateImageSize(maxImageDimension+1, 1, 0); err == nil {
		t.Fatal("accepted image wider than the dimension limit")
	}
	if _, err := validateImageSize(4096, 4096, maxSheetPixels-1); err == nil {
		t.Fatal("accepted image exceeding the aggregate decoded-pixel budget")
	}
	if pixels, err := validateImageSize(64, 32, 0); err != nil || pixels != 2048 {
		t.Fatalf("validateImageSize() = %d, %v; want 2048, nil", pixels, err)
	}
}

func TestValidateCanvasEnforcesOutputPixelBudget(t *testing.T) {
	images := []image.Image{
		boundsOnlyImage{bounds: image.Rect(0, 0, 8192, 2048)},
		boundsOnlyImage{bounds: image.Rect(0, 0, 8192, 2048)},
	}
	if err := validateCanvas(images); err == nil {
		t.Fatal("accepted a canvas exceeding the output-pixel budget")
	}
}


func TestLoadPNGsRejectsOversizedDimensionsBeforeDecode(t *testing.T) {
	root := t.TempDir()
	writeTestPNG(t, filepath.Join(root, "oversized.png"), maxImageDimension+1, 1)
	if _, err := loadPNGs(root); err == nil {
		t.Fatal("loadPNGs accepted an image wider than the configured dimension limit")
	}
}


func TestLoadPNGsSkipsSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation may require elevated privileges on Windows")
	}
	root := t.TempDir()
	target := filepath.Join(t.TempDir(), "outside.png")
	writeTestPNG(t, target, 2, 2)
	if err := os.Symlink(target, filepath.Join(root, "linked.png")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	images, err := loadPNGs(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 0 {
		t.Fatalf("loadPNGs loaded %d symlinked images, want 0", len(images))
	}
}


func TestValidatePNGFileSizeEnforcesAggregateBudget(t *testing.T) {
	if err := validatePNGFileSize(maxPNGBytes+1, 0); err == nil {
		t.Fatal("accepted an encoded PNG larger than the per-file limit")
	}
	if err := validatePNGFileSize(maxPNGBytes, maxPNGTotalBytes-maxPNGBytes+1); err == nil {
		t.Fatal("accepted encoded PNG input exceeding the aggregate limit")
	}
	if err := validatePNGFileSize(1024, 0); err != nil {
		t.Fatalf("validatePNGFileSize() rejected a small PNG: %v", err)
	}
}
