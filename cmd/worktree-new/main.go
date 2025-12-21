package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bearing-dev/bearing/pkg/worktree"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: worktree-new <repo> <branch>\n")
		os.Exit(1)
	}

	repo := os.Args[1]
	branch := os.Args[2]

	root := worktree.FindRoot()
	git := worktree.NewGit(root)
	state := worktree.NewState(root)

	if !git.RepoExists(repo) {
		fmt.Fprintf(os.Stderr, "Error: repo not found: %s\n", repo)
		os.Exit(1)
	}

	wtPath := git.WorktreePath(repo, branch)
	fmt.Printf("Creating worktree: %s\n", wtPath)

	if err := git.CreateWorktree(repo, branch); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	state.Append(worktree.Entry{
		Folder: filepath.Base(wtPath),
		Repo:   repo,
		Branch: branch,
		Base:   false,
	})

	fmt.Printf("Done. Worktree created at: %s\n", wtPath)
}
