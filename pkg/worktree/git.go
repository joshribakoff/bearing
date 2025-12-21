package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Git provides git operations for worktree management
type Git struct {
	root string
}

// NewGit creates a git helper for the given workspace root
func NewGit(root string) *Git {
	return &Git{root: root}
}

// RepoPath returns the path to a repo folder
func (g *Git) RepoPath(repo string) string {
	return filepath.Join(g.root, repo)
}

// WorktreePath returns the path for a worktree
func (g *Git) WorktreePath(repo, branch string) string {
	return filepath.Join(g.root, fmt.Sprintf("%s-%s", repo, branch))
}

// RepoExists checks if a repo directory exists and is a git repo
func (g *Git) RepoExists(repo string) bool {
	path := g.RepoPath(repo)
	info, err := os.Stat(filepath.Join(path, ".git"))
	if err != nil {
		return false
	}
	return info.IsDir()
}

// CreateWorktree creates a new git worktree
func (g *Git) CreateWorktree(repo, branch string) error {
	repoPath := g.RepoPath(repo)
	wtPath := g.WorktreePath(repo, branch)

	cmd := exec.Command("git", "worktree", "add", wtPath, "-b", branch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add failed: %s: %w", output, err)
	}
	return nil
}

// RemoveWorktree removes a git worktree
func (g *Git) RemoveWorktree(repo, branch string) error {
	repoPath := g.RepoPath(repo)
	wtPath := g.WorktreePath(repo, branch)

	// Remove from git
	cmd := exec.Command("git", "worktree", "remove", wtPath)
	cmd.Dir = repoPath
	cmd.CombinedOutput() // Ignore errors if already removed

	// Delete branch
	cmd = exec.Command("git", "branch", "-D", branch)
	cmd.Dir = repoPath
	cmd.CombinedOutput() // Ignore errors

	return nil
}

// GetBranch returns the current branch of a folder
func (g *Git) GetBranch(folder string) (string, error) {
	path := filepath.Join(g.root, folder)
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// ListWorktrees returns all worktrees for a repo
func (g *Git) ListWorktrees(repo string) ([]string, error) {
	repoPath := g.RepoPath(repo)
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "worktree ") {
			paths = append(paths, strings.TrimPrefix(line, "worktree "))
		}
	}
	return paths, nil
}

// IsMainBranch checks if the branch is main or master
func IsMainBranch(branch string) bool {
	return branch == "main" || branch == "master"
}
