package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repo wraps git operations for a repository
type Repo struct {
	path string
}

// NewRepo creates a Repo for the given path
func NewRepo(path string) *Repo {
	return &Repo{path: path}
}

// run executes a git command in the repo directory
func (r *Repo) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.path
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

// CurrentBranch returns the current branch name
func (r *Repo) CurrentBranch() (string, error) {
	return r.run("rev-parse", "--abbrev-ref", "HEAD")
}

// IsDirty returns true if there are uncommitted changes
func (r *Repo) IsDirty() (bool, error) {
	out, err := r.run("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return out != "", nil
}

// UnpushedCount returns the number of commits ahead of origin
func (r *Repo) UnpushedCount(branch string) (int, error) {
	out, err := r.run("rev-list", "--count", fmt.Sprintf("origin/%s..%s", branch, branch))
	if err != nil {
		// Branch might not have upstream
		return 0, nil
	}
	var count int
	fmt.Sscanf(out, "%d", &count)
	return count, nil
}

// WorktreeAdd creates a new worktree with optional start point
func (r *Repo) WorktreeAdd(path, branch, startPoint string) error {
	args := []string{"worktree", "add", "-b", branch, path}
	if startPoint != "" {
		args = append(args, startPoint)
	}
	_, err := r.run(args...)
	return err
}

// WorktreeAddExisting attaches to an existing branch
func (r *Repo) WorktreeAddExisting(path, branch string) error {
	_, err := r.run("worktree", "add", path, branch)
	return err
}

// WorktreeRemove removes a worktree
func (r *Repo) WorktreeRemove(path string) error {
	_, err := r.run("worktree", "remove", path)
	return err
}

// WorktreeList lists all worktrees
func (r *Repo) WorktreeList() ([]WorktreeInfo, error) {
	out, err := r.run("worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return parseWorktreeList(out), nil
}

// BranchDelete deletes a branch
func (r *Repo) BranchDelete(branch string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	_, err := r.run("branch", flag, branch)
	return err
}

// Fetch fetches from origin
func (r *Repo) Fetch() error {
	_, err := r.run("fetch", "--prune")
	return err
}

// RemoteBranchExists checks if a remote branch exists
func (r *Repo) RemoteBranchExists(branch string) bool {
	_, err := r.run("rev-parse", "--verify", fmt.Sprintf("origin/%s", branch))
	return err == nil
}

// ListRemoteBranches returns all remote branch names (without origin/ prefix)
func (r *Repo) ListRemoteBranches() ([]string, error) {
	out, err := r.run("branch", "-r", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}
	var branches []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "origin/HEAD" {
			continue
		}
		// Strip origin/ prefix
		if strings.HasPrefix(line, "origin/") {
			branches = append(branches, strings.TrimPrefix(line, "origin/"))
		}
	}
	return branches, nil
}

// WorktreeInfo contains information about a git worktree
type WorktreeInfo struct {
	Path   string
	Branch string
	Bare   bool
}

func parseWorktreeList(output string) []WorktreeInfo {
	var worktrees []WorktreeInfo
	var current WorktreeInfo

	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = WorktreeInfo{}
			}
			continue
		}
		if strings.HasPrefix(line, "worktree ") {
			current.Path = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "branch ") {
			ref := strings.TrimPrefix(line, "branch ")
			current.Branch = filepath.Base(ref)
		} else if line == "bare" {
			current.Bare = true
		}
	}
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}
	return worktrees
}
