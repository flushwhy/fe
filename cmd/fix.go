package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flushwhy/fe/internal/compdb"
	"github.com/spf13/cobra"
)

var fixCompileCommandsCmd = &cobra.Command{
	Use:   "compile-commands [path]",
	Short: "Repair a C/C++ compilation database for clangd.",
	Long: `Finds the active CMake build directory (including CMake Tools settings in
.vscode/settings.json), enables CMAKE_EXPORT_COMPILE_COMMANDS, and validates
the generated database. Existing invalid databases are backed up as
compile_commands.json.bak before regeneration.

This uses CMake's actual configuration; it does not guess compiler flags or
run a build. Make-based projects are left untouched — use Bear with your
normal build command instead.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) == 1 {
			root = args[0]
		}
		root, err := filepath.Abs(root)
		if err != nil {
			return err
		}
		if info, err := os.Stat(root); err != nil || !info.IsDir() {
			return fmt.Errorf("project path %q is not a directory", root)
		}
		buildDir, count, err := compdb.Repair(root)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, buildDir)
		if err != nil {
			rel = buildDir
		}
		fmt.Printf("compile_commands.json is ready: %s (%d entries)\n", filepath.Join(rel, "compile_commands.json"), count)
		fmt.Println("clangd will use this database when its .clangd CompilationDatabase points to that directory.")
		return nil
	},
}

var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Repair common project configuration issues.",
}

func init() {
	fixCmd.AddCommand(fixCompileCommandsCmd)
	rootCmd.AddCommand(fixCmd)
}
