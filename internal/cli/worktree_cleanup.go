package cli

import (
	"fmt"
	"path/filepath"

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

	baseRepo := filepath.Join(WorkspaceDir(), repoName)
	folderName := fmt.Sprintf("%s-%s", repoName, branch)
	worktreePath := filepath.Join(WorkspaceDir(), folderName)
	store := jsonl.NewStore(WorkspaceDir())

	repo := git.NewRepo(baseRepo)

	// Remove the worktree
	fmt.Printf("Removing worktree: %s\n", worktreePath)
	if err := repo.WorktreeRemove(worktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	// Delete the branch if it exists and has been merged
	_ = repo.BranchDelete(branch, false)

	// Update workflow.jsonl - mark as merged
	workflows, err := store.ReadWorkflow()
	if err != nil {
		return err
	}
	for i, w := range workflows {
		if w.Repo == repoName && w.Branch == branch {
			workflows[i].Status = "merged"
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
