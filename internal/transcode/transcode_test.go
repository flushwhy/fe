package transcode

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/flushwhy/fe/internal/config"
)

func TestRun(t *testing.T) {
	t.Run("fails when inputFile is empty", func(t *testing.T) {
		err := Run(config.TranscodeConfig{}, nil)
		if err == nil || !strings.Contains(err.Error(), "inputFile is required") {
			t.Errorf("expected inputFile error, got: %v", err)
		}
	})

	t.Run("calls runner with correct args (basic)", func(t *testing.T) {
		cfg := config.TranscodeConfig{
			InputFile:  "in.wav",
			OutputFile: "out.ogg",
		}

		var gotInput, gotOutput string
		var gotArgs map[string]interface{}

		runner := func(in, out string, args map[string]interface{}) error {
			gotInput, gotOutput, gotArgs = in, out, args
			return nil
		}

		if err := Run(cfg, runner); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if gotInput != "in.wav" {
			t.Errorf("input: got %s, want in.wav", gotInput)
		}
		if gotOutput != "out.ogg" {
			t.Errorf("output: got %s, want out.ogg", gotOutput)
		}
		if len(gotArgs) != 0 {
			t.Errorf("expected no extra args, got %v", gotArgs)
		}
	})

	t.Run("calls runner with all options set", func(t *testing.T) {
		cfg := config.TranscodeConfig{
			InputFile:       "in.mov",
			OutputFile:      "out.mp4",
			Codec:           "libx264",
			Bitrate:         "2M",
			AudioChannels:   "2",
			VideoFrameRate:  "60",
			VideoResolution: "1920x1080",
			StartTime:       "00:00:10",
			EndTime:         "00:00:20",
		}

		want := map[string]interface{}{
			"c:v": "libx264",
			"b:v": "2M",
			"ac":  "2",
			"r":   "60",
			"s":   "1920x1080",
			"ss":  "00:00:10",
			"t":   "00:00:20",
		}

		runner := func(in, out string, args map[string]interface{}) error {
			if !reflect.DeepEqual(args, want) {
				t.Errorf("args mismatch:\ngot  %v\nwant %v", args, want)
			}
			return nil
		}

		if err := Run(cfg, runner); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("wraps runner error", func(t *testing.T) {
		cfg := config.TranscodeConfig{InputFile: "in.wav", OutputFile: "out.ogg"}

		runner := func(_, _ string, _ map[string]interface{}) error {
			return errors.New("ffmpeg died")
		}

		err := Run(cfg, runner)
		if err == nil || !strings.Contains(err.Error(), "ffmpeg died") {
			t.Errorf("expected wrapped error, got: %v", err)
		}
	})
}
