package daemon

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPathToClaudeDir(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/Users/josh/Projects/foo", "-Users-josh-Projects-foo"},
		{"/home/user/work/project", "-home-user-work-project"},
	}

	for _, tt := range tests {
		result := pathToClaudeDir(tt.input)
		if result != tt.expected {
			t.Errorf("pathToClaudeDir(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestScanWorktree(t *testing.T) {
	// Create temp workspace
	workspaceDir := t.TempDir()
	worktreeDir := filepath.Join(workspaceDir, "test-worktree")
	if err := os.MkdirAll(worktreeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create mock Claude sessions dir
	claudeDir := t.TempDir()
	claudeDirName := pathToClaudeDir(worktreeDir)
	sessionDir := filepath.Join(claudeDir, claudeDirName)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create mock session files
	sessionFile1 := filepath.Join(sessionDir, "abc12345-1234-1234-1234-123456789abc.jsonl")
	sessionFile2 := filepath.Join(sessionDir, "def67890-1234-1234-1234-123456789def.jsonl")

	if err := os.WriteFile(sessionFile1, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond) // Ensure different mod times
	if err := os.WriteFile(sessionFile2, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create scanner with custom Claude dir
	scanner := &SessionScanner{
		workspaceDir: workspaceDir,
		claudeDir:    claudeDir,
	}

	session := scanner.ScanWorktree("test-worktree")
	if session == nil {
		t.Fatal("expected session, got nil")
	}

	if session.Folder != "test-worktree" {
		t.Errorf("folder = %q, want %q", session.Folder, "test-worktree")
	}

	// Should return the most recent session (sessionFile2)
	if session.SessionID != "def67890-1234-1234-1234-123456789def" {
		t.Errorf("sessionID = %q, want %q", session.SessionID, "def67890-1234-1234-1234-123456789def")
	}
}

func TestScanWorktreeNoSessions(t *testing.T) {
	workspaceDir := t.TempDir()
	claudeDir := t.TempDir()

	scanner := &SessionScanner{
		workspaceDir: workspaceDir,
		claudeDir:    claudeDir,
	}

	session := scanner.ScanWorktree("nonexistent-folder")
	if session != nil {
		t.Errorf("expected nil, got %+v", session)
	}
}

func TestScanAll(t *testing.T) {
	workspaceDir := t.TempDir()
	claudeDir := t.TempDir()

	// Create two worktrees with sessions
	for _, folder := range []string{"worktree1", "worktree2"} {
		worktreePath := filepath.Join(workspaceDir, folder)
		if err := os.MkdirAll(worktreePath, 0755); err != nil {
			t.Fatal(err)
		}

		claudeDirName := pathToClaudeDir(worktreePath)
		sessionDir := filepath.Join(claudeDir, claudeDirName)
		if err := os.MkdirAll(sessionDir, 0755); err != nil {
			t.Fatal(err)
		}

		sessionFile := filepath.Join(sessionDir, "11111111-2222-3333-4444-555555555555.jsonl")
		if err := os.WriteFile(sessionFile, []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	scanner := &SessionScanner{
		workspaceDir: workspaceDir,
		claudeDir:    claudeDir,
	}

	sessions := scanner.ScanAll([]string{"worktree1", "worktree2", "worktree3"})
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(sessions))
	}
}
