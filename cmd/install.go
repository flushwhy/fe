package cmd

import (
	"fmt"
	"strings"

	"github.com/flushwhy/fe/internal/odinpkg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var installCmd = &cobra.Command{
	Use:   "install <package>",
	Short: "Install an Odin package by alias or repository URL.",
	Long: `Downloads an Odin source repository into the local package directory.

Examples:
  fe install ltdk
  fe get https://github.com/flushwhy/ltdk.odin
  fe install ltdk --dir vendor/odin

The default destination is ./mylibs/<package>. Existing destinations are never overwritten.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = viper.GetString("packages.directory")
		}
		if dir == "" {
			dir = "mylibs"
		}
		dest, err := odinpkg.Install(args[0], dir)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Installed %s\n", dest)
		return nil
	},
}

var getCmd = &cobra.Command{
	Use:   "get <repository-url>",
	Short: "Install an Odin package from a repository URL.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !strings.HasPrefix(args[0], "https://") {
			return fmt.Errorf("get expects a repository URL; try: fe install %s", args[0])
		}
		return installCmd.RunE(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(installCmd, getCmd)
	installCmd.Flags().String("dir", "", "package directory (default: mylibs or packages.directory in .fe.yaml)")
	getCmd.Flags().String("dir", "", "package directory (default: mylibs or packages.directory in .fe.yaml)")
}
