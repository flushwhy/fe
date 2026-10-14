package cmd

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/spf13/cobra"
)

var o3deProjectCmd = &cobra.Command{
    Use:   "project",
    Short: "Create and manage O3DE projects.",
}

var o3deProjectCreateCmd = &cobra.Command{
    Use:   "create",
    Short: "Create an O3DE project using the official O3DE CLI.",
    Args:  cobra.NoArgs,
    RunE: func(cmd *cobra.Command, args []string) error {
        path, _ := cmd.Flags().GetString("path")
        name, _ := cmd.Flags().GetString("name")
        template, _ := cmd.Flags().GetString("template")
        if strings.TrimSpace(path) == "" {
            return fmt.Errorf("--path is required")
        }
        abs, err := filepath.Abs(path)
        if err != nil {
            return fmt.Errorf("resolve project path: %w", err)
        }
        if _, err := os.Stat(abs); err == nil {
            return fmt.Errorf("project destination %q already exists", abs)
        } else if !os.IsNotExist(err) {
            return fmt.Errorf("inspect project destination %q: %w", abs, err)
        }
        cliArgs := []string{"create-project", "--project-path", abs}
        if strings.TrimSpace(name) != "" {
            cliArgs = append(cliArgs, "--project-name", name)
        }
        if strings.TrimSpace(template) != "" {
            cliArgs = append(cliArgs, "--template-name", template)
        }
        return runO3DECLI(cmd, cliArgs...)
    },
}

var o3deProjectRegisterCmd = &cobra.Command{
    Use:   "register",
    Short: "Register an existing O3DE project.",
    Args:  cobra.NoArgs,
    RunE: func(cmd *cobra.Command, args []string) error {
        path, _ := cmd.Flags().GetString("path")
        if strings.TrimSpace(path) == "" {
            return fmt.Errorf("--path is required")
        }
        abs, err := filepath.Abs(path)
        if err != nil {
            return fmt.Errorf("resolve project path: %w", err)
        }
        if info, err := os.Stat(filepath.Join(abs, "project.json")); err != nil || info.IsDir() {
            return fmt.Errorf("%q does not contain a project.json manifest", abs)
        }
        cliArgs := []string{"register", "--project-path", abs}
        engine, _ := cmd.Flags().GetString("engine-path")
        if strings.TrimSpace(engine) != "" {
            engineAbs, err := filepath.Abs(engine)
            if err != nil {
                return fmt.Errorf("resolve engine path: %w", err)
            }
            cliArgs = append(cliArgs, "--engine-path", engineAbs)
        }
        return runO3DECLI(cmd, cliArgs...)
    },
}

var o3deProjectListCmd = &cobra.Command{
    Use:   "list",
    Short: "List projects registered with O3DE.",
    Args:  cobra.NoArgs,
    RunE: func(cmd *cobra.Command, args []string) error {
        return runO3DECLI(cmd, "register-show", "--projects")
    },
}

var o3deProjectSetGlobalCmd = &cobra.Command{
    Use:   "set-global",
    Short: "Set the default project used by O3DE tools.",
    Args:  cobra.NoArgs,
    RunE: func(cmd *cobra.Command, args []string) error {
        project, _ := cmd.Flags().GetString("project")
        if strings.TrimSpace(project) == "" {
            return fmt.Errorf("--project is required (registered project name or project directory)")
        }
        if info, err := os.Stat(filepath.Join(project, "project.json")); err == nil && !info.IsDir() {
            abs, err := filepath.Abs(project)
            if err != nil {
                return fmt.Errorf("resolve project path: %w", err)
            }
            return runO3DECLI(cmd, "set-global-project", "--project-path", abs)
        }
        return runO3DECLI(cmd, "set-global-project", "--project-name", project)
    },
}

var o3deEngineCmd = &cobra.Command{
    Use:   "engine",
    Short: "Register and inspect O3DE engines.",
}

var o3deEngineRegisterCmd = &cobra.Command{
    Use:   "register",
    Short: "Register an O3DE engine with the O3DE manifest.",
    Args:  cobra.NoArgs,
    RunE: func(cmd *cobra.Command, args []string) error {
        path, _ := cmd.Flags().GetString("path")
        if strings.TrimSpace(path) == "" {
            path, _ = cmd.Flags().GetString("engine")
        }
        if strings.TrimSpace(path) == "" {
            return fmt.Errorf("--path is required unless --engine identifies the engine")
        }
        abs, err := filepath.Abs(path)
        if err != nil {
            return fmt.Errorf("resolve engine path: %w", err)
        }
        if info, err := os.Stat(filepath.Join(abs, "engine.json")); err != nil || info.IsDir() {
            return fmt.Errorf("%q does not contain an engine.json manifest", abs)
        }
        return runO3DECLI(cmd, "register", "--engine-path", abs)
    },
}

var o3deEngineListCmd = &cobra.Command{
    Use:   "list",
    Short: "List engines registered with O3DE.",
    Args:  cobra.NoArgs,
    RunE: func(cmd *cobra.Command, args []string) error {
        return runO3DECLI(cmd, "register-show", "--engines")
    },
}

func init() {
    o3deCmd.AddCommand(o3deProjectCmd, o3deEngineCmd)
    o3deProjectCmd.PersistentFlags().String("engine", "", "O3DE engine root (auto-detected when omitted)")
    o3deEngineCmd.PersistentFlags().String("engine", "", "O3DE engine root (auto-detected when omitted)")

    o3deProjectCreateCmd.Flags().String("path", "", "destination directory for the new project")
    o3deProjectCreateCmd.Flags().String("name", "", "project name (defaults to the destination directory name)")
    o3deProjectCreateCmd.Flags().String("template", "", "registered project template name")
    o3deProjectRegisterCmd.Flags().String("path", "", "existing project directory")
    o3deProjectRegisterCmd.Flags().String("engine-path", "", "engine to associate with the project")
    o3deProjectSetGlobalCmd.Flags().String("project", "", "registered project name or project directory")
    o3deProjectCmd.AddCommand(o3deProjectCreateCmd, o3deProjectRegisterCmd, o3deProjectListCmd, o3deProjectSetGlobalCmd)

    o3deEngineRegisterCmd.Flags().String("path", "", "engine directory containing engine.json")
    o3deEngineCmd.AddCommand(o3deEngineRegisterCmd, o3deEngineListCmd)
}
