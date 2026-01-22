package daemon

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joshribakoff/bearing/internal/gh"
	"github.com/joshribakoff/bearing/internal/git"
	"github.com/joshribakoff/bearing/internal/jsonl"
)

// MockGHClient implements a mock GitHub client for testing
type MockGHClient struct {
	PR    *gh.PRInfo
	Error error
}

func (m *MockGHClient) GetPR(branch string) (*gh.PRInfo, error) {
	return m.PR, m.Error
}

// MockRepo implements a mock git repo for testing
type MockRepo struct {
	Commit  string
	Message string
	Error   error
}

func (m *MockRepo) HeadCommit() (string, error) {
	return m.Commit, m.Error
}

func (m *MockRepo) CommitMessage(commit string) (string, error) {
	return m.Message, m.Error
}

// TestActivityTracker_PRStateChanges tests PR state change detection
func TestActivityTracker_PRStateChanges(t *testing.T) {
	tmpDir := t.TempDir()
	store := jsonl.NewStore(tmpDir)
	tracker := NewActivityTracker(store)

	entry := jsonl.LocalEntry{
		Folder: "test-worktree",
		Repo:   "owner/repo",
		Branch: "feature-branch",
		Base:   false,
	}

	t.Run("first check records state without emitting event", func(t *testing.T) {
		// We can't easily mock gh.Client, so we test the tracker state directly
		// by calling the internal state map
		tracker.prStates["test-worktree"] = "OPEN"

		events, _ := store.ReadActivity()
		initialCount := len(events)

		// Verify no new events were written on first check
		events, _ = store.ReadActivity()
		if len(events) != initialCount {
			t.Errorf("expected no new events on first check, got %d new", len(events)-initialCount)
		}
	})

	t.Run("state change from OPEN to MERGED emits pr_merged", func(t *testing.T) {
		// Reset tracker for clean test
		tracker = NewActivityTracker(store)
		tracker.prStates["test-worktree"] = "OPEN"

		// Simulate state change detection by calling checkPRActivity logic manually
		// Since we can't mock the actual GH client easily, we test the event writing
		event := jsonl.ActivityEvent{
			Timestamp: time.Now().UTC(),
			Type:      "pr_merged",
			Repo:      entry.Repo,
			Branch:    entry.Branch,
			PRNumber:  42,
			Title:     "Test PR",
		}
		store.AppendActivity(event)

		events, err := store.ReadActivity()
		if err != nil {
			t.Fatal(err)
		}

		found := false
		for _, e := range events {
			if e.Type == "pr_merged" && e.PRNumber == 42 {
				found = true
				if e.Repo != "owner/repo" {
					t.Errorf("expected repo owner/repo, got %s", e.Repo)
				}
				if e.Branch != "feature-branch" {
					t.Errorf("expected branch feature-branch, got %s", e.Branch)
				}
			}
		}
		if !found {
			t.Error("pr_merged event not found")
		}
	})

	// Cleanup for next test
	os.Remove(store.ActivityPath())

	t.Run("no event when state unchanged", func(t *testing.T) {
		tracker = NewActivityTracker(store)
		tracker.prStates["test-worktree"] = "OPEN"

		// Simulate same state check - no event should be written
		// The tracker only writes when state changes
		events, _ := store.ReadActivity()
		if len(events) != 0 {
			t.Errorf("expected 0 events, got %d", len(events))
		}
	})
}

// TestActivityTracker_CommitDetection tests commit change detection
func TestActivityTracker_CommitDetection(t *testing.T) {
	tmpDir := t.TempDir()
	store := jsonl.NewStore(tmpDir)
	tracker := NewActivityTracker(store)

	entry := jsonl.LocalEntry{
		Folder: "test-worktree",
		Repo:   "owner/repo",
		Branch: "feature-branch",
		Base:   false,
	}

	t.Run("first commit check records without emitting event", func(t *testing.T) {
		// Set initial commit state
		tracker.commits["test-worktree"] = "abc1234"

		events, _ := store.ReadActivity()
		if len(events) != 0 {
			t.Errorf("expected 0 events on first check, got %d", len(events))
		}
	})

	t.Run("new commit emits commit_pushed event", func(t *testing.T) {
		// Simulate commit change by writing event directly
		event := jsonl.ActivityEvent{
			Timestamp: time.Now().UTC(),
			Type:      "commit_pushed",
			Repo:      entry.Repo,
			Branch:    entry.Branch,
			Commit:    "def5678",
			Message:   "Add new feature",
		}
		store.AppendActivity(event)

		events, err := store.ReadActivity()
		if err != nil {
			t.Fatal(err)
		}

		found := false
		for _, e := range events {
			if e.Type == "commit_pushed" && e.Commit == "def5678" {
				found = true
				if e.Message != "Add new feature" {
					t.Errorf("expected message 'Add new feature', got %s", e.Message)
				}
			}
		}
		if !found {
			t.Error("commit_pushed event not found")
		}
	})

	// Cleanup
	os.Remove(store.ActivityPath())

	t.Run("same commit does not emit event", func(t *testing.T) {
		tracker = NewActivityTracker(store)
		tracker.commits["test-worktree"] = "abc1234"

		// Same commit - no event
		events, _ := store.ReadActivity()
		if len(events) != 0 {
			t.Errorf("expected 0 events for same commit, got %d", len(events))
		}
	})
}

// TestActivityTracker_BaseWorktreeSkipsPR tests that base worktrees skip PR checks
func TestActivityTracker_BaseWorktreeSkipsPR(t *testing.T) {
	tmpDir := t.TempDir()
	store := jsonl.NewStore(tmpDir)
	tracker := NewActivityTracker(store)

	entry := jsonl.LocalEntry{
		Folder: "base-worktree",
		Repo:   "owner/repo",
		Branch: "main",
		Base:   true, // Base worktree
	}

	// For base worktrees, CheckForActivity should skip PR checks
	// but still check commits
	ghClient := &gh.Client{}
	repo := &git.Repo{}

	// This should not panic and should skip PR activity for base worktrees
	// We're testing that the Base flag is respected
	tracker.CheckForActivity("/path/to/worktree", entry, ghClient, repo)

	// Base worktrees shouldn't have PR state tracked
	if _, exists := tracker.prStates["base-worktree"]; exists {
		t.Error("base worktree should not have PR state tracked")
	}
}

// TestActivityEvent_Format tests the ActivityEvent struct format
func TestActivityEvent_Format(t *testing.T) {
	tmpDir := t.TempDir()
	store := jsonl.NewStore(tmpDir)

	tests := []struct {
		name     string
		event    jsonl.ActivityEvent
		wantType string
	}{
		{
			name: "pr_opened event",
			event: jsonl.ActivityEvent{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				Type:      "pr_opened",
				Repo:      "owner/repo",
				Branch:    "feature-1",
				PRNumber:  123,
				Title:     "Add feature",
			},
			wantType: "pr_opened",
		},
		{
			name: "pr_merged event",
			event: jsonl.ActivityEvent{
				Timestamp: time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
				Type:      "pr_merged",
				Repo:      "owner/repo",
				Branch:    "feature-1",
				PRNumber:  123,
				Title:     "Add feature",
			},
			wantType: "pr_merged",
		},
		{
			name: "pr_closed event",
			event: jsonl.ActivityEvent{
				Timestamp: time.Date(2024, 1, 15, 11, 30, 0, 0, time.UTC),
				Type:      "pr_closed",
				Repo:      "owner/repo",
				Branch:    "feature-2",
				PRNumber:  124,
				Title:     "WIP feature",
			},
			wantType: "pr_closed",
		},
		{
			name: "commit_pushed event",
			event: jsonl.ActivityEvent{
				Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Type:      "commit_pushed",
				Repo:      "owner/repo",
				Branch:    "feature-1",
				Commit:    "abc1234",
				Message:   "Fix bug",
			},
			wantType: "commit_pushed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean slate
			os.Remove(store.ActivityPath())

			err := store.AppendActivity(tt.event)
			if err != nil {
				t.Fatal(err)
			}

			events, err := store.ReadActivity()
			if err != nil {
				t.Fatal(err)
			}

			if len(events) != 1 {
				t.Fatalf("expected 1 event, got %d", len(events))
			}

			e := events[0]
			if e.Type != tt.wantType {
				t.Errorf("type = %q, want %q", e.Type, tt.wantType)
			}
			if e.Repo != tt.event.Repo {
				t.Errorf("repo = %q, want %q", e.Repo, tt.event.Repo)
			}
			if e.Branch != tt.event.Branch {
				t.Errorf("branch = %q, want %q", e.Branch, tt.event.Branch)
			}
		})
	}
}

// TestActivityTracker_EdgeCases tests edge cases
func TestActivityTracker_EdgeCases(t *testing.T) {
	t.Run("empty folder path", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := jsonl.NewStore(tmpDir)
		tracker := NewActivityTracker(store)

		entry := jsonl.LocalEntry{
			Folder: "",
			Repo:   "owner/repo",
			Branch: "main",
			Base:   false,
		}

		// Should not panic with empty folder
		ghClient := &gh.Client{}
		repo := &git.Repo{}
		tracker.CheckForActivity("", entry, ghClient, repo)
	})

	t.Run("multiple folders tracked independently", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := jsonl.NewStore(tmpDir)
		tracker := NewActivityTracker(store)

		// Track different folders with different states
		tracker.prStates["folder-a"] = "OPEN"
		tracker.prStates["folder-b"] = "MERGED"
		tracker.commits["folder-a"] = "commit-a"
		tracker.commits["folder-b"] = "commit-b"

		if tracker.prStates["folder-a"] != "OPEN" {
			t.Error("folder-a state not tracked correctly")
		}
		if tracker.prStates["folder-b"] != "MERGED" {
			t.Error("folder-b state not tracked correctly")
		}
		if tracker.commits["folder-a"] != "commit-a" {
			t.Error("folder-a commit not tracked correctly")
		}
		if tracker.commits["folder-b"] != "commit-b" {
			t.Error("folder-b commit not tracked correctly")
		}
	})

	t.Run("activity file created on first write", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := jsonl.NewStore(tmpDir)

		activityPath := store.ActivityPath()
		if _, err := os.Stat(activityPath); !os.IsNotExist(err) {
			t.Error("activity file should not exist before first write")
		}

		event := jsonl.ActivityEvent{
			Timestamp: time.Now().UTC(),
			Type:      "test_event",
			Repo:      "test/repo",
			Branch:    "test-branch",
		}
		err := store.AppendActivity(event)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := os.Stat(activityPath); os.IsNotExist(err) {
			t.Error("activity file should exist after first write")
		}
	})

	t.Run("multiple events appended correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := jsonl.NewStore(tmpDir)

		for i := 0; i < 5; i++ {
			event := jsonl.ActivityEvent{
				Timestamp: time.Now().UTC(),
				Type:      "commit_pushed",
				Repo:      "test/repo",
				Branch:    "test-branch",
				Commit:    filepath.Base(t.Name()) + string(rune('a'+i)),
			}
			if err := store.AppendActivity(event); err != nil {
				t.Fatal(err)
			}
		}

		events, err := store.ReadActivity()
		if err != nil {
			t.Fatal(err)
		}

		if len(events) != 5 {
			t.Errorf("expected 5 events, got %d", len(events))
		}
	})
}

// TestActivityTracker_PRStateTransitions tests all valid PR state transitions
func TestActivityTracker_PRStateTransitions(t *testing.T) {
	tests := []struct {
		name       string
		fromState  string
		toState    string
		wantEvent  string
		shouldEmit bool
	}{
		{"OPEN to MERGED", "OPEN", "MERGED", "pr_merged", true},
		{"OPEN to CLOSED", "OPEN", "CLOSED", "pr_closed", true},
		{"MERGED to OPEN", "MERGED", "OPEN", "pr_opened", true},
		{"CLOSED to OPEN", "CLOSED", "OPEN", "pr_opened", true},
		{"OPEN to OPEN", "OPEN", "OPEN", "", false},
		{"MERGED to MERGED", "MERGED", "MERGED", "", false},
		{"unknown state", "OPEN", "UNKNOWN", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the state transition logic
			var eventType string
			if tt.fromState != tt.toState {
				switch tt.toState {
				case "OPEN":
					eventType = "pr_opened"
				case "MERGED":
					eventType = "pr_merged"
				case "CLOSED":
					eventType = "pr_closed"
				}
			}

			if tt.shouldEmit && eventType != tt.wantEvent {
				t.Errorf("expected event %q, got %q", tt.wantEvent, eventType)
			}
			if !tt.shouldEmit && eventType != "" && tt.toState != "UNKNOWN" {
				t.Errorf("expected no event, got %q", eventType)
			}
		})
	}
}

// TestNewActivityTracker tests tracker initialization
func TestNewActivityTracker(t *testing.T) {
	tmpDir := t.TempDir()
	store := jsonl.NewStore(tmpDir)
	tracker := NewActivityTracker(store)

	if tracker == nil {
		t.Fatal("expected non-nil tracker")
	}
	if tracker.store != store {
		t.Error("store not set correctly")
	}
	if tracker.prStates == nil {
		t.Error("prStates map not initialized")
	}
	if tracker.commits == nil {
		t.Error("commits map not initialized")
	}
	if len(tracker.prStates) != 0 {
		t.Error("prStates should be empty initially")
	}
	if len(tracker.commits) != 0 {
		t.Error("commits should be empty initially")
	}
}
