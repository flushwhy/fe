package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flushwhy/fe/internal/o3de"
	"github.com/spf13/cobra"
)

var o3deGemCmd = &cobra.Command{
	Use:   "gem",
	Short: "Create, register, and manage O3DE Gems using the official O3DE CLI.",
	Long: `Provides convenient wrappers around O3DE's own Gem commands. Fe does not
edit the O3DE manifest or project Gem configuration directly.

Examples:
  fe o3de gem list
  fe o3de gem create --name MyGem --path ./Gems
  fe o3de gem register --path ./Gems/MyGem
  fe o3de gem enable MyGem --project ./MyProject
  fe o3de gem disable MyGem --project ./MyProject`,
}

var o3deGemListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Gems registered in the O3DE user manifest.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := o3de.Registered("gems")
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			fmt.Println("No registered Gems found.")
			return nil
		}
		for _, entry := range entries {
			fmt.Printf("%s\t%s\n", entry.Name, entry.Path)
		}
		return nil
	},
}

var o3deGemCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Gem using O3DE's create-gem command.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		path, _ := cmd.Flags().GetString("path")
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("--name is required")
		}
		if strings.TrimSpace(path) == "" {
			return fmt.Errorf("--path is required")
		}
		if _, err := os.Stat(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("inspect Gem destination %q: %w", path, err)
		}
		return runO3DECLI(cmd, "create-gem", "--gem-name", name, "--gem-path", path)
	},
}

var o3deGemRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register an existing Gem with O3DE.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		if strings.TrimSpace(path) == "" {
			return fmt.Errorf("--path is required")
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("resolve Gem path: %w", err)
		}
		if info, err := os.Stat(filepath.Join(abs, "gem.json")); err != nil || info.IsDir() {
			return fmt.Errorf("%q does not contain a gem.json manifest", abs)
		}
		return runO3DECLI(cmd, "register", "--gem-path", abs)
	},
}

func newO3DEGemProjectCommand(use, short, operation string) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <gem-name>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, _ := cmd.Flags().GetString("project")
			if strings.TrimSpace(project) == "" {
				return fmt.Errorf("--project is required (use a registered project name or path)")
			}
			cliArgs := []string{operation}
			gemPath, _ := cmd.Flags().GetString("gem-path")
			if strings.TrimSpace(gemPath) != "" {
				absGem, err := filepath.Abs(gemPath)
				if err != nil {
					return fmt.Errorf("resolve Gem path: %w", err)
				}
				if info, err := os.Stat(filepath.Join(absGem, "gem.json")); err != nil || info.IsDir() {
					return fmt.Errorf("%q does not contain a gem.json manifest", absGem)
				}
				cliArgs = append(cliArgs, "--gem-path", absGem)
			} else {
				cliArgs = append(cliArgs, "--gem-name", args[0])
			}
			if info, err := os.Stat(filepath.Join(project, "project.json")); err == nil && !info.IsDir() {
				abs, err := filepath.Abs(project)
				if err != nil {
					return fmt.Errorf("resolve project path: %w", err)
				}
				cliArgs = append(cliArgs, "--project-path", abs)
			} else {
				cliArgs = append(cliArgs, "--project-name", project)
			}
			return runO3DECLI(cmd, cliArgs...)
		},
	}
}

func runO3DECLI(cmd *cobra.Command, args ...string) error {
	enginePath, _ := cmd.Flags().GetString("engine")
	engine, err := o3de.FindEngine(enginePath)
	if err != nil {
		return err
	}
	cliPath := o3de.EngineCLI(engine.Path)
	if cliPath == "" {
		return fmt.Errorf("could not find O3DE CLI for engine %q", engine.Path)
	}
	child := exec.CommandContext(cmd.Context(), cliPath, args...)
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr
	child.Stdin = os.Stdin
	if err := child.Run(); err != nil {
		return fmt.Errorf("O3DE CLI %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

func init() {
	o3deCmd.AddCommand(o3deGemCmd)
	o3deGemCmd.PersistentFlags().String("engine", "", "O3DE engine root (auto-detected when omitted)")
	o3deGemCmd.AddCommand(o3deGemListCmd, o3deGemCreateCmd, o3deGemRegisterCmd)
	o3deGemCreateCmd.Flags().String("name", "", "Gem name")
	o3deGemCreateCmd.Flags().String("path", "", "destination directory for the new Gem")
	o3deGemRegisterCmd.Flags().String("path", "", "path to an existing Gem directory")
	enableCmd := newO3DEGemProjectCommand("enable", "Enable a Gem for an O3DE project.", "enable-gem")
	disableCmd := newO3DEGemProjectCommand("disable", "Disable a Gem for an O3DE project.", "disable-gem")
	enableCmd.Flags().String("project", "", "registered O3DE project name or project directory")
	disableCmd.Flags().String("project", "", "registered O3DE project name or project directory")
	enableCmd.Flags().String("gem-path", "", "explicit Gem directory (alternative to the Gem name)")
	disableCmd.Flags().String("gem-path", "", "explicit Gem directory (alternative to the Gem name)")
	o3deGemCmd.AddCommand(enableCmd, disableCmd)
}
