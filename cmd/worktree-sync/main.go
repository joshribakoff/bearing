package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bearing-dev/bearing/pkg/worktree"
)

func main() {
	root := worktree.FindRoot()
	git := worktree.NewGit(root)
	state := worktree.NewState(root)

	// Clear state (will rebuild from git)
	state.Write(nil)

	// Scan workspace for git repos
	entries, err := os.ReadDir(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading workspace: %v\n", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		name := entry.Name()

		// Check if it's a git repo (has .git directory)
		gitDir := filepath.Join(root, name, ".git")
		info, err := os.Stat(gitDir)
		if err != nil || !info.IsDir() {
			continue
		}

		// It's a base repo
		branch, _ := git.GetBranch(name)
		state.Append(worktree.Entry{
			Folder: name,
			Repo:   name,
			Branch: branch,
			Base:   true,
		})

		// Find its worktrees
		worktrees, err := git.ListWorktrees(name)
		if err != nil {
			continue
		}

		for _, wtPath := range worktrees {
			wtName := filepath.Base(wtPath)
			if wtName == name {
				continue // Skip base repo itself
			}
			wtBranch, _ := git.GetBranch(wtName)
			state.Append(worktree.Entry{
				Folder: wtName,
				Repo:   name,
				Branch: wtBranch,
				Base:   false,
			})
		}
	}

	fmt.Println("Synced worktrees.jsonl from git state")
}
