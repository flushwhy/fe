package transcode

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/flushwhy/fe/internal/config"
)

func TestRun(t *testing.T) {
	t.Run("requires input, output, and runner", func(t *testing.T) {
		if err := Run(config.TranscodeConfig{}, nil); err == nil || !strings.Contains(err.Error(), "inputFile") {
			t.Fatalf("expected inputFile error, got %v", err)
		}
		if err := Run(config.TranscodeConfig{InputFile: "in.wav"}, nil); err == nil || !strings.Contains(err.Error(), "outputFile") {
			t.Fatalf("expected outputFile error, got %v", err)
		}
		if err := Run(config.TranscodeConfig{InputFile: "in.wav", OutputFile: "out.ogg"}, nil); err == nil || !strings.Contains(err.Error(), "runner") {
			t.Fatalf("expected runner error, got %v", err)
		}
	})

	t.Run("passes audio settings", func(t *testing.T) {
		cfg := config.TranscodeConfig{InputFile: "in.wav", OutputFile: "out.ogg", Codec: "libvorbis", Bitrate: "128k", AudioChannels: "2"}
		want := map[string]interface{}{"c:a": "libvorbis", "b:a": "128k", "ac": "2"}
		var got map[string]interface{}
		err := Run(cfg, func(_, _ string, args map[string]interface{}) error { got = args; return nil })
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("args mismatch: got %v, want %v", got, want)
		}
	})

	t.Run("uses game-friendly MP4 defaults", func(t *testing.T) {
		got := buildArgs(config.TranscodeConfig{OutputFile: "trailer.mp4"})
		want := map[string]interface{}{
			"c:v": "libx264", "preset": "medium", "crf": "20", "pix_fmt": "yuv420p",
			"c:a": "aac", "b:a": "192k", "movflags": "+faststart",
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("args mismatch: got %v, want %v", got, want)
		}
	})

	t.Run("supports webm defaults", func(t *testing.T) {
		got := buildArgs(config.TranscodeConfig{OutputFile: "capture.webm"})
		want := map[string]interface{}{
			"c:v": "libvpx-vp9", "preset": "medium", "crf": "30", "pix_fmt": "yuv420p",
			"c:a": "libopus", "b:a": "192k", "b:v": "0",
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("args mismatch: got %v, want %v", got, want)
		}
	})

	t.Run("explicit video options override defaults", func(t *testing.T) {
		cfg := config.TranscodeConfig{
			OutputFile: "capture.mp4", VideoCodec: "libx265", AudioCodec: "libopus",
			Preset: "slow", CRF: "24", PixelFormat: "yuv444p", AudioBitrate: "256k",
			Bitrate: "4M", VideoFrameRate: "60", VideoResolution: "1920x1080",
		}
		want := map[string]interface{}{
			"c:v": "libx265", "preset": "slow", "crf": "24", "pix_fmt": "yuv444p",
			"c:a": "libopus", "b:a": "256k", "b:v": "4M", "movflags": "+faststart",
			"r": "60", "s": "1920x1080",
		}
		if got := buildArgs(cfg); !reflect.DeepEqual(got, want) {
			t.Fatalf("args mismatch: got %v, want %v", got, want)
		}
	})

	t.Run("preserves legacy video codec override and trims", func(t *testing.T) {
		cfg := config.TranscodeConfig{InputFile: "in.mov", OutputFile: "out.mp4", Codec: "libx265", StartTime: "00:00:10", EndTime: "00:00:20"}
		want := map[string]interface{}{
			"c:v": "libx265", "preset": "medium", "crf": "20", "pix_fmt": "yuv420p",
			"c:a": "aac", "b:a": "192k", "movflags": "+faststart",
			"ss": "00:00:10", "t": "00:00:20",
		}
		if got := buildArgs(cfg); !reflect.DeepEqual(got, want) {
			t.Fatalf("args mismatch: got %v, want %v", got, want)
		}
	})

	t.Run("wraps runner errors", func(t *testing.T) {
		cfg := config.TranscodeConfig{InputFile: "in.wav", OutputFile: "out.ogg"}
		err := Run(cfg, func(_, _ string, _ map[string]interface{}) error { return errors.New("ffmpeg died") })
		if err == nil || !strings.Contains(err.Error(), "ffmpeg died") {
			t.Fatalf("expected wrapped error, got %v", err)
		}
	})
}
