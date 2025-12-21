package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bearing-dev/bearing/pkg/worktree"
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

	if err := state.Remove(folder); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating state: %v\n", err)
	}

	fmt.Printf("Cleaned up worktree: %s\n", folder)
}
