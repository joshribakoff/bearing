package worktree

import (
	"os"
	"path/filepath"
)

// FindRoot finds the workspace root directory
// Uses SAILKIT_ROOT env var, or walks up from CWD looking for sailkit-dev folder
func FindRoot() string {
	if root := os.Getenv("SAILKIT_ROOT"); root != "" {
		return root
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}

	// Walk up looking for sailkit-dev folder
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "sailkit-dev")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, "workflow.jsonl")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return cwd
}
