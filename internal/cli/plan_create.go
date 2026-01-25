package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var planCreateProject string

var planCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new plan file with a GUID identifier",
	Long: `Create a new plan file with a GUID identifier.

Example:
  bearing plan create --project bearing "Smart Refresh Queue"

Creates: ~/Projects/plans/bearing/a3f2c-smart-refresh-queue.md`,
	Args: cobra.ExactArgs(1),
	RunE: runPlanCreate,
}

func init() {
	planCreateCmd.Flags().StringVarP(&planCreateProject, "project", "p", "", "project name (required)")
	planCreateCmd.MarkFlagRequired("project")
	planCmd.AddCommand(planCreateCmd)
}

func runPlanCreate(cmd *cobra.Command, args []string) error {
	title := args[0]

	if strings.TrimSpace(planCreateProject) == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title cannot be empty")
	}

	// Generate 5-character ID
	id := generateShortID()
	slug := toKebabCase(title)
	filename := fmt.Sprintf("%s-%s.md", id, slug)

	// Create plan directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	planDir := filepath.Join(home, "Projects", "plans", planCreateProject)
	if err := os.MkdirAll(planDir, 0755); err != nil {
		return fmt.Errorf("failed to create plan directory: %w", err)
	}

	planFile := filepath.Join(planDir, filename)

	content := fmt.Sprintf(`---
id: %s
repo: %s
status: draft
---

# %s

`, id, planCreateProject, title)

	if err := os.WriteFile(planFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write plan file: %w", err)
	}

	fmt.Printf("Created plan file: %s\n", planFile)
	return nil
}

// generateShortID returns a 5-character alphanumeric ID
func generateShortID() string {
	b := make([]byte, 5)
	f, err := os.Open("/dev/urandom")
	if err != nil {
		// Fallback: use current time-based seed
		return "00000"
	}
	defer f.Close()
	f.Read(b)

	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 5)
	for i, v := range b {
		result[i] = charset[int(v)%len(charset)]
	}
	return string(result)
}

// toKebabCase converts a title to kebab-case
func toKebabCase(s string) string {
	// Replace non-alphanumeric with hyphens
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	s = reg.ReplaceAllString(s, "-")
	// Lowercase
	s = strings.ToLower(s)
	// Trim leading/trailing hyphens
	s = strings.Trim(s, "-")
	return s
}
