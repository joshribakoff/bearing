package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joshribakoff/bearing/internal/git"
	"github.com/joshribakoff/bearing/internal/jsonl"
	"github.com/spf13/cobra"
)

var (
	checkQuiet bool
	checkJSON  bool
)

var worktreeCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check worktree health",
	RunE:  runWorktreeCheck,
}

func init() {
	worktreeCheckCmd.Flags().BoolVarP(&checkQuiet, "quiet", "q", false, "only show problems")
	worktreeCheckCmd.Flags().BoolVar(&checkJSON, "json", false, "output as JSON")
	worktreeCmd.AddCommand(worktreeCheckCmd)
}

type checkResult struct {
	Folder   string   `json:"folder"`
	Problems []string `json:"problems,omitempty"`
	OK       bool     `json:"ok"`
}

func runWorktreeCheck(cmd *cobra.Command, args []string) error {
	store := jsonl.NewStore(WorkspaceDir())
	entries, err := store.ReadLocal()
	if err != nil {
		return err
	}

	var results []checkResult
	hasProblems := false

	for _, e := range entries {
		folderPath := filepath.Join(WorkspaceDir(), e.Folder)
		result := checkResult{Folder: e.Folder, OK: true}

		// Check folder exists
		if _, err := os.Stat(folderPath); os.IsNotExist(err) {
			result.Problems = append(result.Problems, "folder missing")
			result.OK = false
		} else {
			// Check git state
			repo := git.NewRepo(folderPath)
			branch, err := repo.CurrentBranch()
			if err != nil {
				result.Problems = append(result.Problems, "cannot determine branch")
				result.OK = false
			} else if branch != e.Branch {
				result.Problems = append(result.Problems, fmt.Sprintf("branch mismatch: expected %s, got %s", e.Branch, branch))
				result.OK = false
			}

			// Check for uncommitted changes on base folders
			if e.Base {
				dirty, _ := repo.IsDirty()
				if dirty {
					result.Problems = append(result.Problems, "base folder has uncommitted changes")
					result.OK = false
				}
			}
		}

		if !result.OK {
			hasProblems = true
		}
		results = append(results, result)
	}

	if checkJSON {
		// Output Claude Code hook format
		hookOutput := struct {
			Continue      bool   `json:"continue"`
			SystemMessage string `json:"systemMessage,omitempty"`
		}{Continue: true}

		if hasProblems {
			msg := "BEARING WARNING: Worktree violations detected. "
			for _, r := range results {
				if !r.OK {
					msg += fmt.Sprintf("'%s': %v. ", r.Folder, r.Problems)
				}
			}
			msg += "Ask user: fix issues or proceed anyway?"
			hookOutput.SystemMessage = msg
		}
		return json.NewEncoder(os.Stdout).Encode(hookOutput)
	}

	for _, r := range results {
		if checkQuiet && r.OK {
			continue
		}
		if r.OK {
			fmt.Printf("✓ %s\n", r.Folder)
		} else {
			fmt.Printf("✗ %s\n", r.Folder)
			for _, p := range r.Problems {
				fmt.Printf("  - %s\n", p)
			}
		}
	}

	if hasProblems {
		os.Exit(1)
	}
	return nil
}
