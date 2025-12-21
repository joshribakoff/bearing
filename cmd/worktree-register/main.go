package main

import (
	"fmt"
	"os"

	"github.com/bearing-dev/bearing/pkg/worktree"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: worktree-register <folder>\n")
		os.Exit(1)
	}

	folder := os.Args[1]
	root := worktree.FindRoot()
	git := worktree.NewGit(root)
	state := worktree.NewState(root)

	branch, err := git.GetBranch(folder)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: not a git repo: %s\n", folder)
		os.Exit(1)
	}

	state.Append(worktree.Entry{
		Folder: folder,
		Repo:   folder,
		Branch: branch,
		Base:   true,
	})

	fmt.Printf("Registered base repo: %s (branch: %s)\n", folder, branch)
}
