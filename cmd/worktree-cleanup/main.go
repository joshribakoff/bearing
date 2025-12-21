package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sailkit-dev/sailkit-dev/pkg/worktree"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: worktree-cleanup <repo> <branch>\n")
		os.Exit(1)
	}

	repo := os.Args[1]
	branch := os.Args[2]

	root := worktree.FindRoot()
	git := worktree.NewGit(root)
	state := worktree.NewState(root)

	folder := filepath.Base(git.WorktreePath(repo, branch))

	if err := git.RemoveWorktree(repo, branch); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	if err := state.RemoveLocal(folder); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating local.jsonl: %v\n", err)
	}

	if err := state.UpdateWorkflowStatus(repo, branch, "completed"); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating workflow.jsonl: %v\n", err)
	}

	fmt.Printf("Cleaned up worktree: %s-%s\n", repo, branch)
}
