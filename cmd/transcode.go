package cmd

import (
	"fmt"
	"time"

	"github.com/flushwhy/fe/internal/config"
	"github.com/flushwhy/fe/internal/transcode"
	"github.com/flushwhy/fe/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var transcodeCmd = &cobra.Command{
	Use:   "transcode",
	Short: "Convert audio/video via ffmpeg.",
	Long: `Transcodes a file using ffmpeg. All options are optional — omitted
flags fall through to ffmpeg defaults.

Examples:
  fe transcode --inputFile music.wav --outputFile music.ogg --codec libvorbis
  fe transcode --inputFile video.mov --outputFile video.mp4 --codec libx264 --bitrate 2M`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
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
	transcodeCmd.Flags().String("inputFile", "", "input file path (required)")
	transcodeCmd.Flags().String("outputFile", "output.ogg", "output file path")
	transcodeCmd.Flags().String("codec", "", "codec (e.g. libvorbis, libx264)")
	transcodeCmd.Flags().String("bitrate", "", "bitrate (e.g. 128k, 2M)")
	transcodeCmd.Flags().String("audioChannels", "", "number of audio channels")
	transcodeCmd.Flags().String("videoFrameRate", "", "video frame rate")
	transcodeCmd.Flags().String("videoResolution", "", "video resolution (e.g. 1920x1080)")
	transcodeCmd.Flags().String("startTime", "", "start time offset")
	transcodeCmd.Flags().String("endTime", "", "end/duration")
	viper.BindPFlag("transcode.inputFile", transcodeCmd.Flags().Lookup("inputFile"))
	viper.BindPFlag("transcode.outputFile", transcodeCmd.Flags().Lookup("outputFile"))
	viper.BindPFlag("transcode.codec", transcodeCmd.Flags().Lookup("codec"))
	viper.BindPFlag("transcode.bitrate", transcodeCmd.Flags().Lookup("bitrate"))
	viper.BindPFlag("transcode.audioChannels", transcodeCmd.Flags().Lookup("audioChannels"))
	viper.BindPFlag("transcode.videoFrameRate", transcodeCmd.Flags().Lookup("videoFrameRate"))
	viper.BindPFlag("transcode.videoResolution", transcodeCmd.Flags().Lookup("videoResolution"))
	viper.BindPFlag("transcode.startTime", transcodeCmd.Flags().Lookup("startTime"))
	viper.BindPFlag("transcode.endTime", transcodeCmd.Flags().Lookup("endTime"))
}
