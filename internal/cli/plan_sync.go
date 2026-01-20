package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	planSyncProject string
	planSyncDryRun  bool
)

var planSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync all plan files with GitHub issues",
	RunE:  runPlanSync,
}

func init() {
	planSyncCmd.Flags().StringVar(&planSyncProject, "project", "", "sync only plans for this project")
	planSyncCmd.Flags().BoolVar(&planSyncDryRun, "dry-run", false, "show what would be synced")
	planCmd.AddCommand(planSyncCmd)
}

func runPlanSync(cmd *cobra.Command, args []string) error {
	plansDir := filepath.Join(WorkspaceDir(), "plans")

	if planSyncProject != "" {
		plansDir = filepath.Join(plansDir, planSyncProject)
	}

	// Find all .md files
	var planFiles []string
	err := filepath.Walk(plansDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".md" {
			planFiles = append(planFiles, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if len(planFiles) == 0 {
		fmt.Println("No plan files found")
		return nil
	}

	fmt.Printf("Found %d plan files\n", len(planFiles))

	for _, pf := range planFiles {
		fm, _, err := parsePlanFile(pf)
		if err != nil {
			fmt.Printf("  %s: error parsing (%v)\n", pf, err)
			continue
		}

		if fm.Issue == "" {
			fmt.Printf("  %s: no issue linked\n", pf)
			continue
		}

		if planSyncDryRun {
			fmt.Printf("  %s: would sync to %s#%s\n", pf, fm.Repo, fm.Issue)
		} else {
			fmt.Printf("  %s: syncing to %s#%s... ", pf, fm.Repo, fm.Issue)
			if err := pushPlanToIssue(pf, fm); err != nil {
				fmt.Printf("ERROR: %v\n", err)
			} else {
				fmt.Printf("OK\n")
			}
		}
	}

	return nil
}

func pushPlanToIssue(planFile string, fm *planFrontmatter) error {
	// Re-read the file to get body
	_, body, err := parsePlanFile(planFile)
	if err != nil {
		return err
	}

	body = strings.TrimSpace(body)

	// Push to GitHub using gh issue edit
	repoPath := filepath.Join(WorkspaceDir(), fm.Repo)
	ghCmd := exec.Command("gh", "issue", "edit", fm.Issue, "--body", body)
	ghCmd.Dir = repoPath
	var stderr bytes.Buffer
	ghCmd.Stderr = &stderr

	if err := ghCmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, stderr.String())
	}

	return nil
}
