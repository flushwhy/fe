package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	blenderaddon "github.com/flushwhy/fe/internal/blender"
	"github.com/spf13/cobra"
)

var blenderCmd = &cobra.Command{
	Use:   "blender",
	Short: "Validate and package Blender add-ons.",
}

var blenderValidateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validate a Blender add-on without running its code.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) == 1 {
			root = args[0]
		}
		report, err := blenderaddon.Validate(root)
		if err != nil {
			return err
		}
		kind := "Blender extension"
		if report.Addon.Legacy {
			kind = "legacy Blender add-on"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Valid %s: %s\n", kind, report.Addon.Name)
		if !report.Addon.Legacy {
			fmt.Fprintf(cmd.OutOrStdout(), "  ID: %s\n  Version: %s\n", report.Addon.ID, report.Addon.Version)
		}
		for _, warning := range report.Warnings {
			fmt.Fprintf(cmd.OutOrStdout(), "Warning: %s\n", warning)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Validation is static; add-on code was not imported or executed.")
		return nil
	},
}

var blenderPackageCmd = &cobra.Command{
	Use:   "package [path]",
	Short: "Create an installable ZIP for a Blender add-on.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) == 1 {
			root = args[0]
		}
		report, err := blenderaddon.Validate(root)
		if err != nil {
			return err
		}
		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			name := report.Addon.ID
			if name == "" {
				name = filepath.Base(report.Addon.Root)
			}
			output = filepath.Join("dist", name+"-"+report.Addon.Version+".zip")
		}
		if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}
		if err := blenderaddon.Package(root, output); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", output)
		for _, warning := range report.Warnings {
			fmt.Fprintf(cmd.OutOrStdout(), "Warning: %s\n", warning)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(blenderCmd)
	blenderCmd.AddCommand(blenderValidateCmd, blenderPackageCmd)
	blenderPackageCmd.Flags().StringP("output", "o", "", "ZIP output path (defaults to dist/<addon>-<version>.zip)")
}
