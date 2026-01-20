package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/joshribakoff/bearing/internal/git"
	"github.com/joshribakoff/bearing/internal/jsonl"
	"github.com/spf13/cobra"
)

var (
	recoverDryRun bool
)

var worktreeRecoverCmd = &cobra.Command{
	Use:   "recover <base-folder>",
	Short: "Recover worktrees for a base folder by re-attaching from remote branches",
	Args:  cobra.ExactArgs(1),
	RunE:  runWorktreeRecover,
}

func init() {
	worktreeRecoverCmd.Flags().BoolVar(&recoverDryRun, "dry-run", false, "show what would be recovered without doing it")
	worktreeCmd.AddCommand(worktreeRecoverCmd)
}

func runWorktreeRecover(cmd *cobra.Command, args []string) error {
	baseFolder := args[0]
	basePath := filepath.Join(WorkspaceDir(), baseFolder)
	repo := git.NewRepo(basePath)
	store := jsonl.NewStore(WorkspaceDir())

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

	// List remote branches
	remoteBranches, err := repo.ListRemoteBranches()
	if err != nil {
		return fmt.Errorf("failed to list remote branches: %w", err)
	}

	// Find branches that have remote but no local worktree
	var toRecover []string
	for _, branch := range remoteBranches {
		// Skip main/master
		if branch == "main" || branch == "master" {
			continue
		}
		if !existingBranches[branch] {
			toRecover = append(toRecover, branch)
		}
	}

	if len(toRecover) == 0 {
		fmt.Println("No branches to recover.")
		return nil
	}

	fmt.Printf("Found %d branches to recover:\n", len(toRecover))
	for _, branch := range toRecover {
		fmt.Printf("  - %s\n", branch)
	}

	if recoverDryRun {
		fmt.Println("\nDry run - no changes made.")
		return nil
	}

	fmt.Println("\nRecovering worktrees...")
	for _, branch := range toRecover {
		sanitizedBranch := strings.ReplaceAll(branch, "/", "-")
		folderName := fmt.Sprintf("%s-%s", baseFolder, sanitizedBranch)
		worktreePath := filepath.Join(WorkspaceDir(), folderName)

		fmt.Printf("  Creating: %s\n", folderName)

		// Attach to existing remote branch
		if err := repo.WorktreeAddExisting(worktreePath, branch); err != nil {
			fmt.Printf("    Error: %v\n", err)
			continue
		}

		// Add to local.jsonl
		if err := store.AppendLocal(jsonl.LocalEntry{
			Folder: folderName,
			Repo:   baseFolder,
			Branch: branch,
			Base:   false,
		}); err != nil {
			fmt.Printf("    Warning: failed to update local.jsonl: %v\n", err)
		}
	}

	fmt.Printf("\nRecovered %d worktrees.\n", len(toRecover))
	return nil
}
