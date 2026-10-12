package cmd

import (
	"fmt"

	"github.com/flushwhy/fe/internal/o3de"
	"github.com/flushwhy/fe/internal/ui"
	"github.com/spf13/cobra"
)

var o3deFindCmd = &cobra.Command{
	Use:   "find <engines|projects|gems|templates>",
	Short: "List O3DE objects registered in the official O3DE manifest.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := o3de.Registered(args[0])
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			fmt.Printf("No registered %s found.\n", args[0])
			fmt.Println("Use O3DE's official CLI to register existing objects; fe does not change registration automatically.")
			return nil
		}
		fmt.Println()
		for _, entry := range entries {
			name := entry.Name
			if entry.Version != "" {
				name += " (" + entry.Version + ")"
			}
			fmt.Printf("  %s  %s\n", ui.StyleBold.Render(name), ui.StyleDim.Render(entry.Path))
		}
		fmt.Printf("\n%d registered %s\n", len(entries), args[0])
		return nil
	},
}

func init() {
	o3deCmd.AddCommand(o3deFindCmd)
}
