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

func TestScanWorktreeIgnoresInvalidUUIDs(t *testing.T) {
	workspaceDir := t.TempDir()
	claudeDir := t.TempDir()

	worktreePath := filepath.Join(workspaceDir, "test-worktree")
	if err := os.MkdirAll(worktreePath, 0755); err != nil {
		t.Fatal(err)
	}

	sessionDir := filepath.Join(claudeDir, pathToClaudeDir(worktreePath))
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create invalid files that should be ignored
	invalidFiles := []string{
		"not-a-uuid.jsonl",
		"12345.jsonl",
		"abc.txt",
		".hidden-file",
		"abc12345-1234-1234-1234-123456789abc.json", // wrong extension
	}
	for _, f := range invalidFiles {
		if err := os.WriteFile(filepath.Join(sessionDir, f), []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	scanner := &SessionScanner{
		workspaceDir: workspaceDir,
		claudeDir:    claudeDir,
	}

	session := scanner.ScanWorktree("test-worktree")
	if session != nil {
		t.Errorf("expected nil (no valid sessions), got %+v", session)
	}
}

func TestScanWorktreeIgnoresDirectories(t *testing.T) {
	workspaceDir := t.TempDir()
	claudeDir := t.TempDir()

	worktreePath := filepath.Join(workspaceDir, "test-worktree")
	if err := os.MkdirAll(worktreePath, 0755); err != nil {
		t.Fatal(err)
	}

	sessionDir := filepath.Join(claudeDir, pathToClaudeDir(worktreePath))
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a directory with UUID-like name (should be ignored)
	uuidDir := filepath.Join(sessionDir, "abc12345-1234-1234-1234-123456789abc.jsonl")
	if err := os.MkdirAll(uuidDir, 0755); err != nil {
		t.Fatal(err)
	}

	scanner := &SessionScanner{
		workspaceDir: workspaceDir,
		claudeDir:    claudeDir,
	}

	session := scanner.ScanWorktree("test-worktree")
	if session != nil {
		t.Errorf("expected nil (directory should be ignored), got %+v", session)
	}
}

func TestScanWorktreeMixedValidInvalid(t *testing.T) {
	workspaceDir := t.TempDir()
	claudeDir := t.TempDir()

	worktreePath := filepath.Join(workspaceDir, "test-worktree")
	if err := os.MkdirAll(worktreePath, 0755); err != nil {
		t.Fatal(err)
	}

	sessionDir := filepath.Join(claudeDir, pathToClaudeDir(worktreePath))
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create mix of valid and invalid files
	validUUID := "abc12345-1234-1234-1234-123456789abc"
	files := map[string]bool{
		validUUID + ".jsonl":   true,  // valid
		"not-a-uuid.jsonl":     false, // invalid
		"random.txt":           false, // invalid
		"another-invalid.json": false, // invalid
	}

	for name := range files {
		if err := os.WriteFile(filepath.Join(sessionDir, name), []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	scanner := &SessionScanner{
		workspaceDir: workspaceDir,
		claudeDir:    claudeDir,
	}

	session := scanner.ScanWorktree("test-worktree")
	if session == nil {
		t.Fatal("expected session, got nil")
	}

	if session.SessionID != validUUID {
		t.Errorf("sessionID = %q, want %q", session.SessionID, validUUID)
	}
}

func TestScanWorktreeEmptySessionDir(t *testing.T) {
	workspaceDir := t.TempDir()
	claudeDir := t.TempDir()

	worktreePath := filepath.Join(workspaceDir, "test-worktree")
	if err := os.MkdirAll(worktreePath, 0755); err != nil {
		t.Fatal(err)
	}

	// Create empty session dir
	sessionDir := filepath.Join(claudeDir, pathToClaudeDir(worktreePath))
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	scanner := &SessionScanner{
		workspaceDir: workspaceDir,
		claudeDir:    claudeDir,
	}

	session := scanner.ScanWorktree("test-worktree")
	if session != nil {
		t.Errorf("expected nil for empty session dir, got %+v", session)
	}
}

func TestPathToClaudeDirEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"standard path", "/Users/josh/Projects/foo", "-Users-josh-Projects-foo"},
		{"linux path", "/home/user/work/project", "-home-user-work-project"},
		{"nested path", "/a/b/c/d/e/f", "-a-b-c-d-e-f"},
		{"root only", "/", "-"},
		{"single dir", "/foo", "-foo"},
		{"path with dashes", "/Users/josh/my-project", "-Users-josh-my-project"},
		{"path with dots", "/Users/josh/project.v2", "-Users-josh-project.v2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pathToClaudeDir(tt.input)
			if result != tt.expected {
				t.Errorf("pathToClaudeDir(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestUUIDPatternMatching(t *testing.T) {
	valid := []string{
		"abc12345-1234-1234-1234-123456789abc.jsonl",
		"00000000-0000-0000-0000-000000000000.jsonl",
		"ffffffff-ffff-ffff-ffff-ffffffffffff.jsonl",
		"12345678-abcd-ef01-2345-67890abcdef0.jsonl",
	}
	invalid := []string{
		"ABC12345-1234-1234-1234-123456789ABC.jsonl", // uppercase
		"abc12345-1234-1234-1234-123456789abc.json",  // wrong extension
		"abc12345-1234-1234-1234-123456789abc",       // no extension
		"abc12345-1234-1234-1234.jsonl",              // missing segment
		"abc1234-1234-1234-1234-123456789abc.jsonl",  // wrong length
		"not-a-uuid-at-all.jsonl",
		"",
	}

	for _, v := range valid {
		if !uuidPattern.MatchString(v) {
			t.Errorf("expected %q to match UUID pattern", v)
		}
	}

	for _, inv := range invalid {
		if uuidPattern.MatchString(inv) {
			t.Errorf("expected %q NOT to match UUID pattern", inv)
		}
	}
}
