package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bearing-dev/bearing/pkg/worktree"
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

	// Sync state first
	syncState(root, git, state)

	// Check invariants
	entries, _ := state.Read()
	var violations []string

	for _, entry := range entries {
		if entry.Base && !worktree.IsMainBranch(entry.Branch) {
			violations = append(violations,
				fmt.Sprintf("'%s' is on '%s' (should be main)", entry.Folder, entry.Branch))
		}
		if entry.Base {
			if dirty, _ := git.IsDirty(entry.Folder); dirty {
				violations = append(violations,
					fmt.Sprintf("'%s' has uncommitted changes", entry.Folder))
			}
		}
	}

	if *jsonMode {
		resp := HookResponse{Continue: true}
		if len(violations) > 0 {
			resp.SystemMessage = fmt.Sprintf(
				"BEARING WARNING: Base folder violation: %s. "+
					"Ask user before proceeding.",
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

	fmt.Println("BEARING: Worktree invariant violations detected!")
	for _, v := range violations {
		fmt.Printf("  VIOLATION: Base folder %s\n", v)
	}
}

func syncState(root string, git *worktree.Git, state *worktree.State) {
	state.Write(nil)
	entries, _ := os.ReadDir(root)
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		name := entry.Name()
		gitDir := filepath.Join(root, name, ".git")
		if info, err := os.Stat(gitDir); err != nil || !info.IsDir() {
			continue
		}
		branch, _ := git.GetBranch(name)
		state.Append(worktree.Entry{
			Folder: name, Repo: name, Branch: branch, Base: true,
		})
		worktrees, _ := git.ListWorktrees(name)
		for _, wtPath := range worktrees {
			wtName := filepath.Base(wtPath)
			if wtName == name {
				continue
			}
			wtBranch, _ := git.GetBranch(wtName)
			state.Append(worktree.Entry{
				Folder: wtName, Repo: name, Branch: wtBranch, Base: false,
			})
		}
	}
}
