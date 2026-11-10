package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flushwhy/fe/internal/archivepkg"
	"github.com/spf13/cobra"
)

var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "Create a ZIP or tar.gz archive from build artifacts.",
	Long: `Packages the contents of the configured build directory without changing
the source files. Symlinks and non-regular files are excluded.

Examples:
  fe package --format zip
  fe package --dir builds/linux-x64 --format tar.gz --version 1.2.0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadProjectConfig()
		if err != nil {
			return err
		}
		directory, _ := cmd.Flags().GetString("dir")
		format, _ := cmd.Flags().GetString("format")
		name, _ := cmd.Flags().GetString("name")
		version, _ := cmd.Flags().GetString("version")
		output, _ := cmd.Flags().GetString("output")
		if directory == "" {
			directory = cfg.Butler.Directory
		}
		if name == "" {
			name, err = os.Getwd()
			if err != nil {
				return err
			}
			name = filepath.Base(name)
		}
		if version == "" {
			version = cfg.Butler.Userversion
		}
		if version == "" {
			version = "dev"
		}
		if output == "" {
			output = filepath.Join("dist", name+"-"+version+archivepkg.FormatExtension(format))
		}
		if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
			return fmt.Errorf("create package output directory: %w", err)
		}
		if err := archivepkg.Create(directory, output, format); err != nil {
			return err
		}
		fmt.Printf("Created %s\n", output)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(packageCmd)
	packageCmd.Flags().String("dir", "", "directory containing build artifacts (defaults to butler.directory)")
	packageCmd.Flags().String("format", "zip", "archive format: zip or tar.gz")
	packageCmd.Flags().String("name", "", "package name (defaults to the current directory name)")
	packageCmd.Flags().String("version", "", "package version (defaults to FE_USERVERSION or dev)")
	packageCmd.Flags().StringP("output", "o", "", "archive output path (defaults to dist/<name>-<version>.<format>)")
}
