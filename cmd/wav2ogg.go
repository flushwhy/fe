/*
Copyright © 2024 Ryan Flush <roflush@pm.me>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/go-audio/wav"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"mccoy.space/g/ogg"
)

// wav2oggCmd represents the wav2ogg command
var wav2oggCmd = &cobra.Command{
	Use:   "wav2ogg",
	Short: "takes a wav file and converts it to an ogg file ",
	Long: `Converts a wav file to an ogg file. In the future, this will also be able to convert a folder full of wav files to ogg.
	--input <input file>.wav --output <output file>.ogg`,

	Run: func(cmd *cobra.Command, args []string) {
        fmt.Println(viper.GetString("inputFile"))
		if err := convertWAVToOGG(viper.GetString("inputFile"), viper.GetString("outputFile")); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	audioCmd.AddCommand(wav2oggCmd)

	wav2oggCmd.Flags().String("inputFile", "", "input file")
	wav2oggCmd.Flags().String("outputFile", "You_forgot_to_specify_an_output_file.ogg", "output file")
}

func convertWAVToOGG(inputFile, outputFile string) error {
	var sampleRate int
	input, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer input.Close()

	// Parse WAV header to get sample rate
	wavDecoder := wav.NewDecoder(input)
	if wavDecoder == nil {
		return fmt.Errorf("Failed to create WAV decoder")
	}
	if wavDecoder.WavAudioFormat == 0 {
		return fmt.Errorf("No audio format found in the WAV file")
	}
	sampleRate = int(wavDecoder.SampleRate)

	// Create output file
	output, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer output.Close()

	// Initialize OGG encoder with sample rate
	encoder := ogg.NewEncoder(uint32(sampleRate), output) // 2 channels for stereo

	// Read input data and write to OGG encoder
	buffer := make([]byte, 4096)
	for {
		n, err := input.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		if err := encoder.Encode(int64(n), [][]byte{buffer[:n]}); err != nil {
			return err
		}
	}
	return nil
}
