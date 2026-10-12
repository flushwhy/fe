package cmd

import (
	"fmt"

	"github.com/flushwhy/fe/internal/o3de"
	"github.com/flushwhy/fe/internal/ui"
	"github.com/spf13/cobra"
)

var o3deFindCmd = &cobra.Command{
	Use:   "find <engines|projects|gems|templates>",
	Short: "Find registered and local O3DE objects without changing registration.",
	Long: `Searches O3DE's manifest and scans local directories for engine.json,
project.json, gem.json, or template.json markers.

By default, searches the current directory and common development folders in
your home directory. Pass --path to scan specific roots, or --registered-only
to limit results to the O3DE manifest. Add --interactive to browse results and
select one with the keyboard.

Examples:
  fe o3de find engines
  fe o3de find engines --interactive
  fe o3de find projects --path ~/Games --path ~/work
  fe o3de find gems --depth 3`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		roots, _ := cmd.Flags().GetStringArray("path")
		depth, _ := cmd.Flags().GetInt("depth")
		registeredOnly, _ := cmd.Flags().GetBool("registered-only")
		interactive, _ := cmd.Flags().GetBool("interactive")
		entries, err := o3de.Discover(args[0], roots, depth, registeredOnly)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			fmt.Printf("No %s found in the searched locations.\n", args[0])
			if !registeredOnly {
				fmt.Println("Try --path /directory/to/search, or use O3DE's official CLI to register an existing object.")
			} else {
				fmt.Println("Use O3DE's official CLI to register existing objects; fe does not change registration automatically.")
			}
			return nil
		}
		if interactive {
			selected, err := runO3DEFinder(args[0], entries)
			if err != nil {
				return err
			}
			if selected != nil {
				fmt.Printf("Selected %s: %s\nPath: %s\n", selected.Kind, selected.Name, selected.Path)
			}
			return nil
		}
		fmt.Println()
		for _, entry := range entries {
			name := entry.Name
			if entry.Version != "" {
				name += " (" + entry.Version + ")"
			}
			fmt.Printf("  %s  %s\n", ui.StyleBold.Render(name), ui.StyleDim.Render(entry.Path))
			fmt.Printf("    %s\n", ui.StyleDim.Render(entry.Source))
		}
		fmt.Printf("\n%d %s found\n", len(entries), args[0])
		return nil
	},
}

func init() {
	o3deCmd.AddCommand(o3deFindCmd)
	o3deFindCmd.Flags().StringArray("path", nil, "additional search root (repeatable)")
	o3deFindCmd.Flags().Int("depth", 4, "maximum directory depth below each search root")
	o3deFindCmd.Flags().Bool("registered-only", false, "list only objects registered in the O3DE manifest")
	o3deFindCmd.Flags().Bool("interactive", false, "browse results in a keyboard-driven picker")
}
