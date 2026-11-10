package cmd

import (
	"fmt"
	"time"

	"github.com/flushwhy/fe/internal/transcode"
	"github.com/flushwhy/fe/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var transcodeCmd = &cobra.Command{
	Use:   "transcode",
	Short: "Convert audio/video via ffmpeg.",
	Long: `Transcodes audio or video using ffmpeg. Video output gets game-friendly
H.264/AAC MP4 defaults (or VP9/Opus for WebM); explicit flags override them.

Examples:
  fe transcode --inputFile music.wav --outputFile music.ogg --codec libvorbis
  fe transcode --inputFile capture.mov --outputFile trailer.mp4
  fe transcode --inputFile capture.mp4 --outputFile web-preview.webm --crf 30`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadProjectConfig()
		if err != nil {
			return err
		}
		if errs := cfg.Validate("transcode"); len(errs) > 0 {
			for _, e := range errs {
				ui.Fail("config: %s", e)
			}
			return fmt.Errorf("missing required config")
		}

		ui.Title("Transcoding  %s → %s", cfg.Transcode.InputFile, cfg.Transcode.OutputFile)
		start := time.Now()

		err = ui.RunSpinner(
			fmt.Sprintf("Running ffmpeg (%s)", cfg.Transcode.Codec),
			func() error {
				return transcode.Run(cfg.Transcode, transcode.DefaultRunner)
			},
		)
		if err != nil {
			return err
		}

		ui.Elapsed(start)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(transcodeCmd)
	flags := transcodeCmd.Flags()
	flags.String("inputFile", "", "input file path (required)")
	flags.String("outputFile", "output.ogg", "output file path")
	flags.String("codec", "", "legacy codec override for the selected media type")
	flags.String("bitrate", "", "legacy bitrate override for the selected media type")
	flags.String("videoCodec", "", "video codec (defaults to H.264 for video output)")
	flags.String("audioCodec", "", "audio codec (defaults to AAC for video output)")
	flags.String("preset", "", "video encoder preset (default: medium)")
	flags.String("crf", "", "video quality, lower is higher quality (default: 20 for MP4)")
	flags.String("pixelFormat", "", "video pixel format (default: yuv420p)")
	flags.String("audioBitrate", "", "audio bitrate (default: 192k for video output)")
	flags.String("audioChannels", "", "number of audio channels")
	flags.String("videoFrameRate", "", "video frame rate (default: preserve source)")
	flags.String("videoResolution", "", "video resolution (e.g. 1920x1080)")
	flags.String("startTime", "", "start time offset")
	flags.String("endTime", "", "end/duration")
	for _, name := range []string{
		"inputFile", "outputFile", "codec", "bitrate", "videoCodec", "audioCodec",
		"preset", "crf", "pixelFormat", "audioBitrate", "audioChannels",
		"videoFrameRate", "videoResolution", "startTime", "endTime",
	} {
		_ = viper.BindPFlag("transcode."+name, flags.Lookup(name))
	}
}
