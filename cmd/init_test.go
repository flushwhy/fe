package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCommand(t *testing.T) {
	t.Run("successfully initializes a new project", func(t *testing.T) {
		tempDir := t.TempDir()
		projectName := "my-new-game"
		projectPath := filepath.Join(tempDir, projectName)

		originalWd, _ := os.Getwd()
		os.Chdir(tempDir)
		defer os.Chdir(originalWd)

		err := initCmd.RunE(initCmd, []string{projectName})

		if err != nil {
			t.Fatalf("initCmd.RunE() returned an unexpected error: %v", err)
		}

		expectedDirs := []string{
			"assets/audio",
			"assets/fonts",
			"assets/sprites",
			"builds",
			"src",
		}

		for _, dir := range expectedDirs {
			dirPath := filepath.Join(projectPath, dir)
			if _, err := os.Stat(dirPath); os.IsNotExist(err) {
				t.Errorf("expected directory to exist, but it doesn't: %s", dirPath)
			}
		}

		configPath := filepath.Join(projectPath, ".fe.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Fatalf("expected config file to exist, but it doesn't: %s", configPath)
		}

		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read config file: %v", err)
		}

		if strings.TrimSpace(string(content)) != strings.TrimSpace(defaultConfigContent) {
			t.Errorf("config file content mismatch")
		}
	})

	t.Run("fails if project directory already exists", func(t *testing.T) {
		tempDir := t.TempDir()
		projectName := "existing-project"
		projectPath := filepath.Join(tempDir, projectName)

		if err := os.Mkdir(projectPath, 0755); err != nil {
			t.Fatalf("failed to create pre-existing directory for test: %v", err)
		}

		originalWd, _ := os.Getwd()
		os.Chdir(tempDir)
		defer os.Chdir(originalWd)

		err := initCmd.RunE(initCmd, []string{projectName})

		if err == nil {
			t.Fatal("expected an error when directory exists, but got nil")
		}

		// ★ FIX: Removed the single quotes to match the actual error output.
		expectedErrorMsg := "directory existing-project already exists"
		if !strings.Contains(err.Error(), expectedErrorMsg) {
			t.Errorf("expected error message to contain '%s', but got: '%s'", expectedErrorMsg, err.Error())
		}
	})
}
