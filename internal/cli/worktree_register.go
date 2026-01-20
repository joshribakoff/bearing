package cli

import (
	"fmt"
	"path/filepath"

	"github.com/joshribakoff/bearing/internal/git"
	"github.com/joshribakoff/bearing/internal/jsonl"
	"github.com/spf13/cobra"
)

var worktreeRegisterCmd = &cobra.Command{
	Use:   "register <folder>",
	Short: "Register an existing folder as a worktree",
	Args:  cobra.ExactArgs(1),
	RunE:  runWorktreeRegister,
}

func init() {
	worktreeCmd.AddCommand(worktreeRegisterCmd)
}

func runWorktreeRegister(cmd *cobra.Command, args []string) error {
	folder := args[0]
	folderPath := filepath.Join(WorkspaceDir(), folder)
	store := jsonl.NewStore(WorkspaceDir())

	repo := git.NewRepo(folderPath)

	// Get current branch
	branch, err := repo.CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get branch: %w", err)
	}

	// Determine if this is a base folder (branch is main/master)
	isBase := branch == "main" || branch == "master"

	// Infer repo name (folder name without branch suffix for worktrees)
	repoName := folder
	if !isBase && len(branch) > 0 {
		suffix := "-" + branch
		if len(folder) > len(suffix) && folder[len(folder)-len(suffix):] == suffix {
			repoName = folder[:len(folder)-len(suffix)]
		}
	}

	entry := jsonl.LocalEntry{
		Folder: folder,
		Repo:   repoName,
		Branch: branch,
		Base:   isBase,
	}

	if err := store.AppendLocal(entry); err != nil {
		return fmt.Errorf("failed to update local.jsonl: %w", err)
	}

	baseStr := ""
	if isBase {
		baseStr = " (base)"
	}
	fmt.Printf("Registered: %s -> %s@%s%s\n", folder, repoName, branch, baseStr)
	return nil
}
