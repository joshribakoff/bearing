package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/joshribakoff/bearing/internal/git"
	"github.com/joshribakoff/bearing/internal/jsonl"
	"github.com/spf13/cobra"
)

var (
	newBasedOn string
	newPurpose string
)

var worktreeNewCmd = &cobra.Command{
	Use:   "new <repo> <branch>",
	Short: "Create a new worktree",
	Args:  cobra.ExactArgs(2),
	RunE:  runWorktreeNew,
}

func init() {
	worktreeNewCmd.Flags().StringVar(&newBasedOn, "based-on", "", "base branch (default: main)")
	worktreeNewCmd.Flags().StringVar(&newPurpose, "purpose", "", "purpose description")
	worktreeCmd.AddCommand(worktreeNewCmd)
}

func runWorktreeNew(cmd *cobra.Command, args []string) error {
	repoName := args[0]
	branch := args[1]

	// Sanitize branch name for folder (replace / with -)
	folderBranch := strings.ReplaceAll(branch, "/", "-")
	folderName := fmt.Sprintf("%s-%s", repoName, folderBranch)

	baseRepo := filepath.Join(WorkspaceDir(), repoName)
	worktreePath := filepath.Join(WorkspaceDir(), folderName)
	store := jsonl.NewStore(WorkspaceDir())

	repo := git.NewRepo(baseRepo)

	// Determine start point
	basedOn := newBasedOn
	if basedOn == "" {
		basedOn = "main"
	}

	// Create the worktree from basedOn branch
	fmt.Printf("Creating worktree: %s\n", worktreePath)
	if err := repo.WorktreeAdd(worktreePath, branch, basedOn); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}
	if err := store.AppendWorkflow(jsonl.WorkflowEntry{
		Repo:    repoName,
		Branch:  branch,
		BasedOn: basedOn,
		Purpose: newPurpose,
		Status:  "active",
		Created: time.Now(),
	}); err != nil {
		return fmt.Errorf("failed to update workflow.jsonl: %w", err)
	}

	// Add to local.jsonl
	if err := store.AppendLocal(jsonl.LocalEntry{
		Folder: folderName,
		Repo:   repoName,
		Branch: branch,
		Base:   false,
	}); err != nil {
		return fmt.Errorf("failed to update local.jsonl: %w", err)
	}

	fmt.Printf("Done. Worktree created at: %s\n", worktreePath)
	return nil
}
