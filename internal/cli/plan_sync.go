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

	created := 0
	updated := 0
	errors := 0

	for _, pf := range planFiles {
		fm, body, err := parsePlanFile(pf)
		if err != nil {
			fmt.Printf("  %s: error parsing (%v)\n", filepath.Base(pf), err)
			errors++
			continue
		}

		// Auto-infer repo from path if missing
		if fm.Repo == "" {
			fm.Repo = inferRepoFromPath(pf)
			if fm.Repo == "" {
				fmt.Printf("  %s: no repo configured\n", filepath.Base(pf))
				errors++
				continue
			}
		}

		// Auto-infer title from markdown heading if missing
		if fm.Title == "" {
			fm.Title = extractTitleFromBody(body)
			if fm.Title == "" {
				fmt.Printf("  %s: no title or heading\n", filepath.Base(pf))
				errors++
				continue
			}
		}

		repoPath := GetRepoPath(fm.Repo)
		body = strings.TrimSpace(body)

		if fm.Issue == "" {
			// Create new issue
			if planSyncDryRun {
				fmt.Printf("  %s: would create issue in %s\n", filepath.Base(pf), fm.Repo)
				created++
			} else {
				fmt.Printf("  %s: creating issue in %s... ", filepath.Base(pf), fm.Repo)
				issueNum, err := createIssueForPlan(repoPath, pf, fm.Title, body)
				if err != nil {
					fmt.Printf("ERROR: %v\n", err)
					errors++
				} else {
					fmt.Printf("OK (#%s)\n", issueNum)
					created++
				}
			}
		} else {
			// Update existing issue
			if planSyncDryRun {
				fmt.Printf("  %s: would sync to %s#%s\n", filepath.Base(pf), fm.Repo, fm.Issue)
				updated++
			} else {
				fmt.Printf("  %s: syncing to %s#%s... ", filepath.Base(pf), fm.Repo, fm.Issue)
				if err := pushPlanToIssue(pf, fm); err != nil {
					fmt.Printf("ERROR: %v\n", err)
					errors++
				} else {
					fmt.Printf("OK\n")
					updated++
				}
			}
		}
	}

	fmt.Printf("\nSummary: %d created, %d updated, %d errors\n", created, updated, errors)
	if planSyncDryRun {
		fmt.Println("(dry run - no changes made)")
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
	repoPath := GetRepoPath(fm.Repo)
	ghCmd := exec.Command("gh", "issue", "edit", fm.Issue, "--body", body)
	ghCmd.Dir = repoPath
	var stderr bytes.Buffer
	ghCmd.Stderr = &stderr

	if err := ghCmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, stderr.String())
	}

	return nil
}

func createIssueForPlan(repoPath, planFile, title, body string) (string, error) {
	ghCmd := exec.Command("gh", "issue", "create",
		"--title", title,
		"--body", body,
		"--label", "plan")
	ghCmd.Dir = repoPath
	var stdout, stderr bytes.Buffer
	ghCmd.Stdout = &stdout
	ghCmd.Stderr = &stderr

	if err := ghCmd.Run(); err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}

	// Parse issue number from output URL: https://github.com/owner/repo/issues/123
	url := strings.TrimSpace(stdout.String())
	parts := strings.Split(url, "/")
	issueNum := parts[len(parts)-1]

	// Update frontmatter with issue number
	if err := updateFrontmatter(planFile, "issue", issueNum); err != nil {
		return issueNum, fmt.Errorf("created issue but failed to update frontmatter: %w", err)
	}

	return issueNum, nil
}
