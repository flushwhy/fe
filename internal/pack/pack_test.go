package pack

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
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
