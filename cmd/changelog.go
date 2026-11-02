package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/flushwhy/fe/internal/changelog"
	"github.com/spf13/cobra"
)

var changelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: "Generate release notes from Git history and tags.",
	Long: `Generates Markdown release notes from commits after the latest reachable tag.
Use --from and --to to select an explicit commit range. Conventional Commit
types are grouped into sections; breaking changes are called out.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		output, _ := cmd.Flags().GetString("output")
		if from == "" {
			tag := exec.Command("git", "describe", "--tags", "--abbrev=0")
			value, err := tag.Output()
			if err == nil {
				from = strings.TrimSpace(string(value))
			}
		}
		result, err := changelog.Generate(".", from, to)
		if err != nil {
			return err
		}
		if output == "" || output == "-" {
			fmt.Print(result)
			return nil
		}
		if err := os.WriteFile(output, []byte(result), 0644); err != nil {
			return fmt.Errorf("write changelog %q: %w", output, err)
		}
		fmt.Printf("Wrote %s\n", output)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(changelogCmd)
	changelogCmd.Flags().String("from", "", "starting tag or commit (defaults to latest reachable tag)")
	changelogCmd.Flags().String("to", "HEAD", "ending tag or commit")
	changelogCmd.Flags().StringP("output", "o", "-", "Markdown output path (use - for stdout)")
}
