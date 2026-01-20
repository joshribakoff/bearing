package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// CreateTestRepo creates a git repo for testing
func CreateTestRepo(t *testing.T, dir, name string) string {
	t.Helper()

	repoPath := filepath.Join(dir, name)
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatal(err)
	}

	// git init
	cmd := exec.Command("git", "init", "--initial-branch=main")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	// Configure git user
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoPath
	cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoPath
	cmd.Run()

	// Initial commit
	cmd = exec.Command("git", "commit", "--allow-empty", "-m", "initial")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	return repoPath
}

// InitWorkspace creates JSONL files for testing
func InitWorkspace(t *testing.T, dir string) {
	t.Helper()

	files := []string{"workflow.jsonl", "local.jsonl", "health.jsonl"}
	for _, f := range files {
		path := filepath.Join(dir, f)
		if err := os.WriteFile(path, []byte{}, 0644); err != nil {
			t.Fatal(err)
		}
	}
}

// RunBearing executes the bearing binary with the given args
func RunBearing(t *testing.T, workspaceDir string, args ...string) (string, error) {
	t.Helper()

	// Build args with workspace flag
	fullArgs := append([]string{"-w", workspaceDir}, args...)

	cmd := exec.Command("bearing", fullArgs...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
