package cmd

import (
	"fmt"

	"github.com/flushwhy/fe/internal/scaffold"
	"github.com/flushwhy/fe/internal/ui"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Browse and manage cached project templates.",
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List templates available from fe-templates.",
	RunE: func(cmd *cobra.Command, args []string) error {
		templates, err := scaffold.ListRemoteTemplates()
		if err != nil {
			return err
		}
		if len(templates) == 0 {
			fmt.Println("No templates are available.")
			return nil
		}
		fmt.Printf("%-24s %-12s %s\\n", "TEMPLATE", "CATEGORY", "DESCRIPTION")
		for _, item := range templates {
			fmt.Printf("%-24s %-12s %s\\n", item.ID, item.Category, item.Description)
		}
		return nil
	},
}

var templateUpdateCmd = &cobra.Command{
	Use:   "update <template-id>",
	Short: "Download the latest version of a template into the local cache.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return ui.RunSpinner("Updating template "+args[0], func() error {
			_, err := scaffold.ResolveTemplate(args[0], true)
			if err != nil {
				return err
			}
			ui.OK("Template %q cached and ready", args[0])
			return nil
		})
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
	templateCmd.AddCommand(templateListCmd, templateUpdateCmd)
}
