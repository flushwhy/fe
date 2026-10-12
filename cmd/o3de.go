package cmd

import (
	"github.com/spf13/cobra"
)

var o3deCmd = &cobra.Command{
	Use:   "o3de",
	Short: "Glue for existing Open 3D Engine tools and workflows.",
	Long: `Helps discover and diagnose O3DE installations without replacing O3DE's
own engine, project, registration, or build tooling.

Examples:
  fe o3de doctor
  fe o3de doctor --engine /path/to/o3de
  fe o3de find engines
  fe o3de find projects`,
}

func init() {
	rootCmd.AddCommand(o3deCmd)
}
