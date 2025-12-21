package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sailkit-dev/sailkit-dev/pkg/worktree"
)

func main() {
	root := worktree.FindRoot()
	git := worktree.NewGit(root)
	state := worktree.NewState(root)

	// Clear local.jsonl (will rebuild from git state)
	state.WriteLocal(nil)

	// Scan workspace for git repos
	entries, err := os.ReadDir(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading workspace: %v\n", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		// Check if it's a git repo (has .git directory)
		gitDir := filepath.Join(root, name, ".git")
		info, err := os.Stat(gitDir)
		if err != nil || !info.IsDir() {
			continue
		}

		// It's a base repo
		branch, _ := git.GetBranch(name)
		state.AppendLocal(worktree.LocalEntry{
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
			state.AppendLocal(worktree.LocalEntry{
				Folder: wtName,
				Repo:   name,
				Branch: wtBranch,
				Base:   false,
			})
		}
	}

	fmt.Println("Synced local.jsonl from git state")
}
