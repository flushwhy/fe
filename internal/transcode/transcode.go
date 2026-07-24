package transcode

import (
	"errors"
	"fmt"

	"github.com/flushwhy/fe/internal/config"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// Runner is a function that wraps an ffmpeg call.
// Swapped out in tests to avoid requiring a real ffmpeg binary.
type Runner func(inputFile, outputFile string, args map[string]interface{}) error

// DefaultRunner calls ffmpeg-go for real.
var DefaultRunner Runner = func(inputFile, outputFile string, args map[string]interface{}) error {
	return ffmpeg.Input(inputFile).
		Output(outputFile, args).
		OverWriteOutput().
		ErrorToStdOut().
		Run()
}

// Run executes a transcode using the given runner and config.
func Run(cfg config.TranscodeConfig, runner Runner) error {
	if cfg.InputFile == "" {
		return errors.New("inputFile is required")
	}

	args := buildArgs(cfg)

	if err := runner(cfg.InputFile, cfg.OutputFile, args); err != nil {
		return fmt.Errorf("transcoding %s → %s: %w", cfg.InputFile, cfg.OutputFile, err)
	}

	fmt.Printf("✓ Transcoded %s → %s\n", cfg.InputFile, cfg.OutputFile)
	return nil
}

// buildArgs converts the typed config into the map ffmpeg-go expects.
// Only non-empty values are included so ffmpeg uses its own defaults.
func buildArgs(cfg config.TranscodeConfig) map[string]interface{} {
	raw := map[string]string{
		"c:v": cfg.Codec,
		"b:v": cfg.Bitrate,
		"ac":  cfg.AudioChannels,
		"r":   cfg.VideoFrameRate,
		"s":   cfg.VideoResolution,
		"ss":  cfg.StartTime,
		"t":   cfg.EndTime,
	}

	out := make(map[string]interface{}, len(raw))
	for k, v := range raw {
		if v != "" {
			out[k] = v
		}
	}
	return out
}
