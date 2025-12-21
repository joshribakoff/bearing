package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sailkit-dev/sailkit-dev/pkg/worktree"
)

type HookResponse struct {
	Continue      bool   `json:"continue"`
	SystemMessage string `json:"systemMessage,omitempty"`
}

func main() {
	jsonMode := flag.Bool("json", false, "Output JSON for Claude Code hooks")
	quiet := flag.Bool("quiet", false, "Suppress output on success")
	flag.Parse()

	root := worktree.FindRoot()
	git := worktree.NewGit(root)
	state := worktree.NewState(root)

	// Run sync first
	syncState(root, git, state)

	// Check invariants
	local, _ := state.ReadLocal()
	var violations []string

	for _, entry := range local {
		if entry.Base && !worktree.IsMainBranch(entry.Branch) {
			violations = append(violations,
				fmt.Sprintf("'%s' is on '%s' (should be main)", entry.Folder, entry.Branch))
		}
	}

	if *jsonMode {
		resp := HookResponse{Continue: true}
		if len(violations) > 0 {
			resp.SystemMessage = fmt.Sprintf(
				"SAILKIT WARNING: Base folders on wrong branch. %s. "+
					"Ask user: fix with 'git -C <folder> checkout main' or proceed anyway?",
				violations[0])
		}
		json.NewEncoder(os.Stdout).Encode(resp)
		return
	}

	if len(violations) == 0 {
		if !*quiet {
			fmt.Println("All invariants satisfied")
		}
		return
	}

	fmt.Println("SAILKIT: Worktree invariant violations detected!")
	for _, v := range violations {
		fmt.Printf("  VIOLATION: Base folder %s\n", v)
	}
}

func syncState(root string, git *worktree.Git, state *worktree.State) {
	// Simplified inline sync
	state.WriteLocal(nil)
	entries, _ := os.ReadDir(root)
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name()[0] == '.' {
			continue
		}
		name := entry.Name()
		gitDir := root + "/" + name + "/.git"
		if info, err := os.Stat(gitDir); err != nil || !info.IsDir() {
			continue
		}
		branch, _ := git.GetBranch(name)
		state.AppendLocal(worktree.LocalEntry{
			Folder: name, Repo: name, Branch: branch, Base: true,
		})
		worktrees, _ := git.ListWorktrees(name)
		for _, wtPath := range worktrees {
			wtName := filepath.Base(wtPath)
			if wtName == name {
				continue
			}
			wtBranch, _ := git.GetBranch(wtName)
			state.AppendLocal(worktree.LocalEntry{
				Folder: wtName, Repo: name, Branch: wtBranch, Base: false,
			})
		}
	}
}
