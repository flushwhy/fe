package cmd

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/flushwhy/fe/internal/update"
	"github.com/flushwhy/fe/internal/version"
	"github.com/spf13/cobra"
)

var (
	updateChannel string
	updateMethod  string
	updateDryRun  bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for and install Fe updates using the detected installation method.",
	Long: `Checks GitHub Releases for Fe updates and delegates package-managed installations
to their existing package manager. Fe does not change package-manager sources.

Examples:
  fe update check
  fe update
  fe update --channel rolling
  fe update --dry-run
  fe update --method auto`,
	Args: cobra.NoArgs,
	RunE: runFeUpdate,
}

var updateCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check the latest Fe release and detected installation method.",
	Args:  cobra.NoArgs,
	RunE:  runFeUpdateCheck,
}

func init() {
	updateCmd.PersistentFlags().StringVar(&updateChannel, "channel", "stable", "release channel: stable or rolling")
	updateCmd.PersistentFlags().StringVar(&updateMethod, "method", "auto", "installation method: auto, chocolatey, homebrew, apt, or direct")
	updateCmd.Flags().BoolVar(&updateDryRun, "dry-run", false, "show the exact update command without running it")
	updateCmd.AddCommand(updateCheckCmd)
	rootCmd.AddCommand(updateCmd)
}

type updateInfo struct {
	install      update.Installation
	release      update.Release
	candidate    string
	candidateErr error
}

func runFeUpdateCheck(cmd *cobra.Command, _ []string) error {
	info, err := inspectFeUpdate(cmd)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	install := info.install
	if install.Method == "" || install.Method == update.Direct {
		if install.Method == update.Direct {
			fmt.Fprintln(out, "Installation method: direct binary (explicit override; not verified)")
		} else {
			fmt.Fprintln(out, "Installation method: unknown (no package-manager ownership record)")
		}

		current := version.Current()
		if current == "dev" || current == "rolling" {
			fmt.Fprintln(out, "Installed version: unknown")
		} else {
			fmt.Fprintf(out, "Installed version: %s\n", current)
		}
		printReleaseStatus(out, updateChannel, current, info.release)
		fmt.Fprintln(out, "No update was attempted. Direct-binary updating is not part of this stage.")
		return nil
	}

	fmt.Fprintf(out, "Installation method: %s\n", install.Method)
	fmt.Fprintf(out, "Installed package: fe %s\n", install.PackageVersion)
	if updateChannel == "rolling" {
		printReleaseStatus(out, updateChannel, install.PackageVersion, info.release)
		fmt.Fprintln(out, "Package managers do not publish Fe's rolling channel; rolling installation updates are not enabled yet.")
		return nil
	}

	fmt.Fprintf(out, "Latest stable release: %s\n", info.release.Version)
	if info.candidateErr != nil {
		fmt.Fprintf(out, "Package source candidate: unavailable (%v)\n", info.candidateErr)
	} else {
		fmt.Fprintf(out, "Package source candidate: %s\n", info.candidate)
	}

	switch {
	case update.ComparePackageVersions(install.Method, install.PackageVersion, info.release.Version) >= 0:
		fmt.Fprintln(out, "Status: up to date.")
	case info.candidateErr != nil || update.ComparePackageVersions(install.Method, info.candidate, info.release.Version) < 0:
		fmt.Fprintf(out, "Status: the configured %s package source does not yet offer the latest Fe release. No update will be attempted.\n", install.Method)
		fmt.Fprintf(out, "Release: %s\n", info.release.URL)
	default:
		fmt.Fprintln(out, "Status: an update is available through the detected package manager.")
	}
	return nil
}

func printReleaseStatus(out io.Writer, channel, current string, release update.Release) {
	if channel == "rolling" {
		fmt.Fprintf(out, "Latest rolling commit: %s\n", release.Commit[:min(7, len(release.Commit))])
		switch {
		case version.Commit == "":
			fmt.Fprintln(out, "Status: this build has no embedded commit; update status cannot be compared.")
		case strings.EqualFold(version.Commit, release.Commit):
			fmt.Fprintln(out, "Status: up to date.")
		default:
			fmt.Fprintf(out, "Status: a newer rolling build is available.\nRelease: %s\n", release.URL)
		}
		return
	}

	fmt.Fprintf(out, "Latest stable release: %s\n", release.Version)
	switch {
	case current == "dev" || current == "rolling":
		fmt.Fprintln(out, "Status: unable to compare this non-stable build with the latest release.")
	case update.CompareVersions(current, release.Version) >= 0:
		fmt.Fprintln(out, "Status: up to date.")
	default:
		fmt.Fprintf(out, "Status: Fe %s is available.\nRelease: %s\n", release.Version, release.URL)
	}
}

func runFeUpdate(cmd *cobra.Command, _ []string) error {
	info, err := inspectFeUpdate(cmd)
	if err != nil {
		return err
	}

	install := info.install
	if install.Method == "" {
		return fmt.Errorf("could not verify package-manager ownership; specify --method only when you know how Fe was installed")
	}
	if install.Method == update.Direct {
		return fmt.Errorf("direct-binary updates are not implemented yet; Fe will not replace this executable")
	}
	if updateChannel != "stable" {
		return fmt.Errorf("%s does not publish Fe's rolling channel; rolling binary updates will be added with the direct-binary updater", install.Method)
	}
	if update.ComparePackageVersions(install.Method, install.PackageVersion, info.release.Version) >= 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "Fe %s is already up to date.\n", install.PackageVersion)
		return nil
	}
	if info.candidateErr != nil {
		return fmt.Errorf("cannot query the configured %s package source: %w", install.Method, info.candidateErr)
	}
	if update.ComparePackageVersions(install.Method, info.candidate, info.release.Version) < 0 {
		return fmt.Errorf("%s currently offers Fe %q, while GitHub's latest stable release is %q; refusing to run an update that cannot reach the current release", install.Method, info.candidate, info.release.Version)
	}

	command, err := update.UpdateCommand(install.Method, updateChannel)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Detected installation: %s (Fe %s)\n", install.Method, install.PackageVersion)
	fmt.Fprintln(out, "Update command:")
	fmt.Fprintf(out, "  %s %s\n", command.Name, strings.Join(command.Args, " "))
	if updateDryRun {
		fmt.Fprintln(out, "Dry run: no changes made.")
		return nil
	}

	fmt.Fprint(out, "Run this command? [y/N] ")
	var answer string
	if _, err := fmt.Fscan(cmd.InOrStdin(), &answer); err != nil {
		fmt.Fprintln(out, "Update cancelled.")
		return nil
	}
	if !strings.EqualFold(answer, "y") && !strings.EqualFold(answer, "yes") {
		fmt.Fprintln(out, "Update cancelled.")
		return nil
	}

	process := exec.CommandContext(cmd.Context(), command.Name, command.Args...)
	process.Stdin = cmd.InOrStdin()
	process.Stdout = out
	process.Stderr = cmd.ErrOrStderr()
	return process.Run()
}

func inspectFeUpdate(cmd *cobra.Command) (updateInfo, error) {
	install, err := update.Detect(cmd.Context(), update.Method(updateMethod))
	if err != nil {
		return updateInfo{}, err
	}
	release, err := update.Latest(cmd.Context(), updateChannel)
	if err != nil {
		return updateInfo{}, err
	}

	info := updateInfo{install: install, release: release}
	if updateChannel != "rolling" && install.Method != "" && install.Method != update.Direct {
		info.candidate, info.candidateErr = update.Candidate(cmd.Context(), install)
	}
	return info, nil
}

