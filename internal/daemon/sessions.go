package daemon

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/joshribakoff/bearing/internal/jsonl"
)

// SessionScanner discovers Claude Code sessions for worktrees
type SessionScanner struct {
	workspaceDir string
	claudeDir    string
}

// NewSessionScanner creates a session scanner
func NewSessionScanner(workspaceDir string) *SessionScanner {
	homeDir, _ := os.UserHomeDir()
	return &SessionScanner{
		workspaceDir: workspaceDir,
		claudeDir:    filepath.Join(homeDir, ".claude", "projects"),
	}
}

// uuidPattern matches Claude session file names
var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\.jsonl$`)

// pathToClaudeDir converts a filesystem path to Claude's directory naming scheme
// e.g., /Users/josh/Projects/foo -> -Users-josh-Projects-foo
func pathToClaudeDir(absPath string) string {
	return strings.ReplaceAll(absPath, string(os.PathSeparator), "-")
}

// ScanWorktree finds the most recent Claude session for a worktree folder
func (s *SessionScanner) ScanWorktree(folder string) *jsonl.ClaudeSessionEntry {
	absPath := filepath.Join(s.workspaceDir, folder)
	claudeDirName := pathToClaudeDir(absPath)
	sessionDir := filepath.Join(s.claudeDir, claudeDirName)

	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		return nil
	}

	var mostRecent *jsonl.ClaudeSessionEntry
	var latestMod time.Time

	for _, entry := range entries {
		if entry.IsDir() || !uuidPattern.MatchString(entry.Name()) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if mostRecent == nil || info.ModTime().After(latestMod) {
			sessionID := strings.TrimSuffix(entry.Name(), ".jsonl")
			latestMod = info.ModTime()
			mostRecent = &jsonl.ClaudeSessionEntry{
				Folder:    folder,
				SessionID: sessionID,
				LastUsed:  info.ModTime(),
			}
		}
	}

	return mostRecent
}

// ScanAll scans all worktrees and returns session entries
func (s *SessionScanner) ScanAll(folders []string) []jsonl.ClaudeSessionEntry {
	var sessions []jsonl.ClaudeSessionEntry
	for _, folder := range folders {
		if session := s.ScanWorktree(folder); session != nil {
			sessions = append(sessions, *session)
		}
	}
	return sessions
}
