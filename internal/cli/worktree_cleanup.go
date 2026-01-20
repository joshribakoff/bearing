package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/joshribakoff/bearing/internal/git"
	"github.com/joshribakoff/bearing/internal/jsonl"
	"github.com/spf13/cobra"
)

var worktreeCleanupCmd = &cobra.Command{
	Use:   "cleanup <repo> <branch>",
	Short: "Remove a worktree and update manifests",
	Args:  cobra.ExactArgs(2),
	RunE:  runWorktreeCleanup,
}

func init() {
	worktreeCmd.AddCommand(worktreeCleanupCmd)
}

func runWorktreeCleanup(cmd *cobra.Command, args []string) error {
	repoName := args[0]
	branch := args[1]

	// Sanitize branch name for folder (replace / with -)
	folderBranch := strings.ReplaceAll(branch, "/", "-")
	folderName := fmt.Sprintf("%s-%s", repoName, folderBranch)

	baseRepo := filepath.Join(WorkspaceDir(), repoName)
	worktreePath := filepath.Join(WorkspaceDir(), folderName)
	store := jsonl.NewStore(WorkspaceDir())

	repo := git.NewRepo(baseRepo)

	// Remove the worktree
	fmt.Printf("Removing worktree: %s\n", worktreePath)
	if err := repo.WorktreeRemove(worktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	// Delete the branch if it has been merged (non-force delete)
	// If delete fails, branch has unmerged commits
	status := "cleaned"
	if err := repo.BranchDelete(branch, false); err == nil {
		status = "merged"
	}

	// Update workflow.jsonl with actual status
	workflows, err := store.ReadWorkflow()
	if err != nil {
		return err
	}
	for i, w := range workflows {
		if w.Repo == repoName && w.Branch == branch {
			workflows[i].Status = status
		}
	}
	if err := store.WriteWorkflow(workflows); err != nil {
		return fmt.Errorf("failed to update workflow.jsonl: %w", err)
	}

	// Update local.jsonl - remove entry
	locals, err := store.ReadLocal()
	if err != nil {
		return err
	}
	var newLocals []jsonl.LocalEntry
	for _, l := range locals {
		if l.Folder != folderName {
			newLocals = append(newLocals, l)
		}
	}
	if err := store.WriteLocal(newLocals); err != nil {
		return fmt.Errorf("failed to update local.jsonl: %w", err)
	}

	fmt.Printf("Done. Worktree removed: %s\n", folderName)
	return nil
}
