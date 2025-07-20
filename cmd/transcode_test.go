package cmd

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestTranscodeCommandLogic(t *testing.T) {

	originalRunner := ffmpegRunner
	defer func() { ffmpegRunner = originalRunner }()

	t.Run("fails when no input file is provided", func(t *testing.T) {
		viper.Reset()

		viper.Set("transcode.outputFile", "test.ogg")

		err := runTranscode(nil, []string{})

		if err == nil {
			t.Fatal("expected an error, but got nil")
		}
		if !strings.Contains(err.Error(), "--inputFile must be set") {
			t.Errorf("expected error about missing input file, but got: %v", err)
		}
	})

	t.Run("transcodes with basic options", func(t *testing.T) {
		viper.Reset()

		ffmpegRunner = func(inputFile, outputFile string, args map[string]interface{}) error {
			if inputFile != "input.wav" {
				t.Errorf("expected input file 'input.wav', got '%s'", inputFile)
			}
			if outputFile != "output.ogg" {
				t.Errorf("expected output file 'output.ogg', got '%s'", outputFile)
			}
			if len(args) != 0 {
				t.Errorf("expected no extra args, but got %v", args)
			}
			return nil
		}

		viper.Set("transcode.inputFile", "input.wav")
		viper.Set("transcode.outputFile", "output.ogg")

		err := runTranscode(nil, []string{})
		if err != nil {
			t.Errorf("expected no error, but got: %v", err)
		}
	})

	t.Run("transcodes with all options", func(t *testing.T) {
		viper.Reset()

		expectedArgs := map[string]interface{}{
			"-c:v": "libx264",
			"-b:v": "2M",
			"-ac":  "1",
			"-r":   "60",
			"-s":   "1920x1080",
			"-ss":  "00:00:10",
			"-t":   "00:00:20",
		}

		ffmpegRunner = func(inputFile, outputFile string, args map[string]interface{}) error {
			if inputFile != "input.mov" {
				t.Errorf("expected input file 'input.mov', got '%s'", inputFile)
			}
			if outputFile != "output.mp4" {
				t.Errorf("expected output file 'output.mp4', got '%s'", outputFile)
			}
			if !reflect.DeepEqual(args, expectedArgs) {
				t.Errorf("unexpected ffmpeg args map:\ngot:  %v\nwant: %v", args, expectedArgs)
			}
			return nil
		}

		viper.Set("transcode.inputFile", "input.mov")
		viper.Set("transcode.outputFile", "output.mp4")
		viper.Set("transcode.codec", "libx264")
		viper.Set("transcode.bitrate", "2M")
		viper.Set("transcode.audioChannels", "1")
		viper.Set("transcode.videoFrameRate", "60")
		viper.Set("transcode.videoResolution", "1920x1080")
		viper.Set("transcode.startTime", "00:00:10")
		viper.Set("transcode.endTime", "00:00:20")

		err := runTranscode(nil, []string{})
		if err != nil {
			t.Errorf("expected no error, but got: %v", err)
		}
	})
}
