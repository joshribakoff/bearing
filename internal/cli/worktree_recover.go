package cli

import (
	"fmt"
	"path/filepath"

	"github.com/joshribakoff/bearing/internal/git"
	"github.com/spf13/cobra"
)

var worktreeRecoverCmd = &cobra.Command{
	Use:   "recover <base-folder>",
	Short: "Recover worktrees for a base folder by re-attaching from remote branches",
	Args:  cobra.ExactArgs(1),
	RunE:  runWorktreeRecover,
}

func init() {
	worktreeCmd.AddCommand(worktreeRecoverCmd)
}

func runWorktreeRecover(cmd *cobra.Command, args []string) error {
	baseFolder := args[0]
	basePath := filepath.Join(WorkspaceDir(), baseFolder)
	repo := git.NewRepo(basePath)

	// Fetch latest remote state
	fmt.Println("Fetching remote branches...")
	if err := repo.Fetch(); err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}

	// List existing worktrees
	worktrees, err := repo.WorktreeList()
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	existingBranches := make(map[string]bool)
	for _, wt := range worktrees {
		existingBranches[wt.Branch] = true
	}

	// Note: Full recovery would require listing remote branches
	// For now, we just report what exists
	fmt.Println("Existing worktrees:")
	for _, wt := range worktrees {
		fmt.Printf("  %s -> %s\n", wt.Path, wt.Branch)
	}

	fmt.Println("\nTo recover a specific branch, use:")
	fmt.Printf("  bearing worktree new %s <branch-name>\n", baseFolder)

	return nil
}
