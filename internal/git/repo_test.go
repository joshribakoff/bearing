package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func createTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "repo")
	os.MkdirAll(repoPath, 0755)

	cmds := [][]string{
		{"git", "init", "--initial-branch=main"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "commit", "--allow-empty", "-m", "initial"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}
	return repoPath
}

func TestCurrentBranch(t *testing.T) {
	repoPath := createTestRepo(t)
	repo := NewRepo(repoPath)

	branch, err := repo.CurrentBranch()
	if err != nil {
		t.Fatal(err)
	}
	if branch != "main" {
		t.Errorf("expected main, got %s", branch)
	}
}

func TestIsDirty(t *testing.T) {
	repoPath := createTestRepo(t)
	repo := NewRepo(repoPath)

	// Clean repo
	dirty, err := repo.IsDirty()
	if err != nil {
		t.Fatal(err)
	}
	if dirty {
		t.Error("expected clean repo")
	}

	// Create untracked file
	os.WriteFile(filepath.Join(repoPath, "new.txt"), []byte("test"), 0644)

	dirty, err = repo.IsDirty()
	if err != nil {
		t.Fatal(err)
	}
	if !dirty {
		t.Error("expected dirty repo")
	}
}

func TestWorktreeAddRemove(t *testing.T) {
	repoPath := createTestRepo(t)
	repo := NewRepo(repoPath)

	worktreePath := filepath.Join(filepath.Dir(repoPath), "worktree-feature")

	// Add worktree
	if err := repo.WorktreeAdd(worktreePath, "feature"); err != nil {
		t.Fatalf("WorktreeAdd failed: %v", err)
	}

	// Verify it exists
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		t.Error("worktree directory not created")
	}

	// Verify branch
	wtRepo := NewRepo(worktreePath)
	branch, _ := wtRepo.CurrentBranch()
	if branch != "feature" {
		t.Errorf("expected feature branch, got %s", branch)
	}

	// List worktrees
	worktrees, err := repo.WorktreeList()
	if err != nil {
		t.Fatal(err)
	}
	if len(worktrees) != 2 {
		t.Errorf("expected 2 worktrees, got %d", len(worktrees))
	}

	// Remove worktree
	if err := repo.WorktreeRemove(worktreePath); err != nil {
		t.Fatalf("WorktreeRemove failed: %v", err)
	}

	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		t.Error("worktree directory still exists")
	}
}

func TestParseWorktreeList(t *testing.T) {
	output := `worktree /path/to/main
branch refs/heads/main

worktree /path/to/feature
branch refs/heads/feature

`
	worktrees := parseWorktreeList(output)
	if len(worktrees) != 2 {
		t.Errorf("expected 2 worktrees, got %d", len(worktrees))
	}
	if worktrees[0].Branch != "main" {
		t.Errorf("expected main, got %s", worktrees[0].Branch)
	}
	if worktrees[1].Branch != "feature" {
		t.Errorf("expected feature, got %s", worktrees[1].Branch)
	}
}
