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

	"github.com/spf13/cobra"
)

// pngjoinerCmd represents the pngjoiner command
var pngjoinerCmd = &cobra.Command{
	Use:   "pngjoiner",
	Short: "This takes multiple PNGs and combines them into a single PNG.",
	Long: `This takes multiple PNGs and combines them into a single PNG.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pngjoiner called")
	},
}

func init() {
	rootCmd.AddCommand(pngjoinerCmd)

    pngjoinerCmd.Flags().String("output", "you_forgot_to_specify_an_output_file.png", "output file")
    pngjoinerCmd.Flags().String("input", "", "input file(s) or directory")
    pngjoinerCmd.Flags().String("rows", "5", "rows")
    pngjoinerCmd.Flags().String("cols", "5", "cols")
}
