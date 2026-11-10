// Package cmd defines Fe's command-line interface.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flushwhy/fe/internal/compdb"
	"github.com/flushwhy/fe/internal/tooling"
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
	RunE: func(_ *cobra.Command, args []string) error {
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
		configured, err := tooling.ConfigureClangd(root, buildDir)
		if err != nil {
			return fmt.Errorf("database repaired, but updating .clangd failed: %w", err)
		}
		if configured {
			fmt.Println(".clangd is configured to use this database.")
		} else {
			fmt.Println("Your custom .clangd was left unchanged; set CompileFlags.CompilationDatabase to", rel)
		}
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
