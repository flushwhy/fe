package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/flushwhy/fe/internal/gumroad"
	"github.com/flushwhy/fe/internal/ui"
	"github.com/spf13/cobra"
)

var gumroadCmd = &cobra.Command{
	Use:   "gumroad",
	Short: "Publish releases to Gumroad.",
}

var gumroadReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Upload a stable release to its configured Gumroad product.",
	Long: `Uploads the files in the existing artifact manifest to the Gumroad product
explicitly configured for the requested version.

Each stable version must have its own pre-created, empty Gumroad product.
Authentication is handled by the installed Gumroad CLI.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadProjectConfig()
		if err != nil {
			return err
		}
		version, _ := cmd.Flags().GetString("version")
		if version == "" {
			version = cfg.Butler.Userversion
		}
		directory, _ := cmd.Flags().GetString("dir")
		if directory == "" {
			directory = cfg.Butler.Directory
		}
		productID, files, err := gumroad.ValidateRelease(version, directory, cfg.Gumroad.Stable)
		if err != nil {
			return err
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		fmt.Printf("Gumroad stable release %s\n", version)
		fmt.Printf("Product: %s\n", productID)
		fmt.Printf("Artifacts: %d\n", len(files))
		for _, file := range files {
			fmt.Printf("  - %s\n", file)
		}
		if dryRun {
			fmt.Println("Dry run — no Gumroad commands were run and no files were uploaded.")
			return nil
		}
		if _, err := exec.LookPath("gumroad"); err != nil {
			return fmt.Errorf("Gumroad CLI not found; install it and authenticate with gumroad auth login")
		}

		logger := log.New(os.Stdout, "", 0)
		if err := ui.RunSpinner("Uploading stable release to Gumroad", func() error {
			return gumroad.Publish(version, directory, cfg.Gumroad.Stable, exec.Command, logger)
		}); err != nil {
			return err
		}
		ui.OK("Gumroad upload verified for version %s", strings.TrimSpace(version))
		return nil
	},
}

func init() {
	gumroadCmd.AddCommand(gumroadReleaseCmd)
	rootCmd.AddCommand(gumroadCmd)
	gumroadReleaseCmd.Flags().String("version", "", "stable release version (defaults to butler.userversion)")
	gumroadReleaseCmd.Flags().String("dir", "", "artifact directory (defaults to butler.directory)")
	gumroadReleaseCmd.Flags().Bool("dry-run", false, "show the configured product and artifacts without uploading")
}
