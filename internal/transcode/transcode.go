package transcode

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/flushwhy/fe/internal/config"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// Runner wraps an ffmpeg call so transcodes can be tested without ffmpeg.
type Runner func(inputFile, outputFile string, args map[string]interface{}) error

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
	if runner == nil {
		return errors.New("transcode runner is required")
	}
	if cfg.OutputFile == "" {
		return errors.New("outputFile is required")
	}

	if err := runner(cfg.InputFile, cfg.OutputFile, buildArgs(cfg)); err != nil {
		return fmt.Errorf("transcoding %s → %s: %w", cfg.InputFile, cfg.OutputFile, err)
	}

	fmt.Printf("✓ Transcoded %s → %s\n", cfg.InputFile, cfg.OutputFile)
	return nil
}

// buildArgs selects audio-only or game-friendly video defaults from the output
// extension. Explicit settings always win; legacy codec/bitrate fields remain
// supported as the codec/bitrate override for the selected media type.
func buildArgs(cfg config.TranscodeConfig) map[string]interface{} {
	args := make(map[string]interface{})
	video := isVideoOutput(cfg.OutputFile)

	if video {
		videoCodec := first(cfg.VideoCodec, cfg.Codec, "libx264")
		args["c:v"] = videoCodec
		args["preset"] = first(cfg.Preset, "medium")
		args["crf"] = first(cfg.CRF, "20")
		args["pix_fmt"] = first(cfg.PixelFormat, "yuv420p")
		args["c:a"] = first(cfg.AudioCodec, "aac")
		args["b:a"] = first(cfg.AudioBitrate, "192k")
		if cfg.Bitrate != "" {
			args["b:v"] = cfg.Bitrate
		}
		switch strings.ToLower(filepath.Ext(cfg.OutputFile)) {
		case ".mp4", ".m4v", ".mov":
			args["movflags"] = "+faststart"
		case ".webm":
			if cfg.VideoCodec == "" && cfg.Codec == "" {
				args["c:v"] = "libvpx-vp9"
				args["crf"] = first(cfg.CRF, "30")
				args["b:v"] = "0"
			}
			if cfg.AudioCodec == "" {
				args["c:a"] = "libopus"
			}
		}
		if cfg.AudioChannels != "" {
			args["ac"] = cfg.AudioChannels
		}
		if cfg.VideoFrameRate != "" {
			args["r"] = cfg.VideoFrameRate
		}
		if cfg.VideoResolution != "" {
			args["s"] = cfg.VideoResolution
		}
	} else {
		if codec := first(cfg.AudioCodec, cfg.Codec); codec != "" {
			args["c:a"] = codec
		}
		if bitrate := first(cfg.AudioBitrate, cfg.Bitrate); bitrate != "" {
			args["b:a"] = bitrate
		}
		if cfg.AudioChannels != "" {
			args["ac"] = cfg.AudioChannels
		}
	}

	if cfg.StartTime != "" {
		args["ss"] = cfg.StartTime
	}
	if cfg.EndTime != "" {
		args["t"] = cfg.EndTime
	}
	return args
}

func isVideoOutput(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp4", ".m4v", ".mov", ".mkv", ".webm", ".avi", ".ogv":
		return true
	default:
		return false
	}
}

func first(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
