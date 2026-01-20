package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joshribakoff/bearing/internal/git"
	"github.com/joshribakoff/bearing/internal/jsonl"
	"github.com/spf13/cobra"
)

var worktreeSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Rebuild local.jsonl from git state",
	RunE:  runWorktreeSync,
}

func init() {
	worktreeCmd.AddCommand(worktreeSyncCmd)
}

func runWorktreeSync(cmd *cobra.Command, args []string) error {
	store := jsonl.NewStore(WorkspaceDir())

	// Find all git repos in workspace
	entries, err := os.ReadDir(WorkspaceDir())
	if err != nil {
		return err
	}

	var locals []jsonl.LocalEntry

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		folder := e.Name()
		folderPath := filepath.Join(WorkspaceDir(), folder)

		// Check if it's a git repo
		gitDir := filepath.Join(folderPath, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			continue
		}

		repo := git.NewRepo(folderPath)
		branch, err := repo.CurrentBranch()
		if err != nil {
			continue
		}

		// Determine if base folder
		isBase := branch == "main" || branch == "master"

		// Infer repo name
		repoName := folder
		if !isBase {
			// Check if folder name ends with -branch (with / sanitized to -)
			sanitizedBranch := strings.ReplaceAll(branch, "/", "-")
			suffix := "-" + sanitizedBranch
			if strings.HasSuffix(folder, suffix) {
				repoName = folder[:len(folder)-len(suffix)]
			}
		}

		locals = append(locals, jsonl.LocalEntry{
			Folder: folder,
			Repo:   repoName,
			Branch: branch,
			Base:   isBase,
		})

		fmt.Printf("Found: %s -> %s@%s", folder, repoName, branch)
		if isBase {
			fmt.Print(" (base)")
		}
		fmt.Println()
	}

	if err := store.WriteLocal(locals); err != nil {
		return fmt.Errorf("failed to write local.jsonl: %w", err)
	}

	fmt.Printf("\nSynced %d entries to local.jsonl\n", len(locals))
	return nil
}
