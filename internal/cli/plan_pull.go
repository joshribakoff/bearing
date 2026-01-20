package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var planPullCmd = &cobra.Command{
	Use:   "pull <repo> <issue>",
	Short: "Create a plan file from a GitHub issue",
	Args:  cobra.ExactArgs(2),
	RunE:  runPlanPull,
}

func init() {
	planCmd.AddCommand(planPullCmd)
}

func runPlanPull(cmd *cobra.Command, args []string) error {
	repo := args[0]
	issue := args[1]

	repoPath := filepath.Join(WorkspaceDir(), repo)

	// Fetch issue details using gh
	ghCmd := exec.Command("gh", "issue", "view", issue, "--json", "title,body,labels")
	ghCmd.Dir = repoPath
	output, err := ghCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to fetch issue: %w", err)
	}

	// Parse issue (simplified - just extract title/body)
	issueStr := string(output)

	// Create plan file
	planDir := filepath.Join(WorkspaceDir(), "plans", repo)
	if err := os.MkdirAll(planDir, 0755); err != nil {
		return err
	}

	planFile := filepath.Join(planDir, fmt.Sprintf("%s.md", sanitizeFilename(issue)))

	content := fmt.Sprintf(`---
issue: %s
repo: %s
status: draft
---

# Issue %s

%s
`, issue, repo, issue, issueStr)

	if err := os.WriteFile(planFile, []byte(content), 0644); err != nil {
		return err
	}

	fmt.Printf("Created plan file: %s\n", planFile)
	return nil
}

func sanitizeFilename(s string) string {
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, " ", "-")
	return strings.ToLower(s)
}
