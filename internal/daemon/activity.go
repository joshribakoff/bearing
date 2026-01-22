package daemon

import (
	"time"

	"github.com/joshribakoff/bearing/internal/gh"
	"github.com/joshribakoff/bearing/internal/git"
	"github.com/joshribakoff/bearing/internal/jsonl"
)

// ActivityTracker tracks workspace activity events
type ActivityTracker struct {
	store    *jsonl.Store
	prStates map[string]string // folder -> last known PR state
	commits  map[string]string // folder -> last known commit
}

// NewActivityTracker creates a new activity tracker
func NewActivityTracker(store *jsonl.Store) *ActivityTracker {
	return &ActivityTracker{
		store:    store,
		prStates: make(map[string]string),
		commits:  make(map[string]string),
	}
}

// CheckForActivity checks a worktree for new activity and emits events
func (t *ActivityTracker) CheckForActivity(folderPath string, entry jsonl.LocalEntry, ghClient *gh.Client, repo *git.Repo) {
	if !entry.Base {
		t.checkPRActivity(folderPath, entry, ghClient)
	}
	t.checkCommitActivity(folderPath, entry, repo)
}

func (t *ActivityTracker) checkPRActivity(folderPath string, entry jsonl.LocalEntry, ghClient *gh.Client) {
	pr, err := ghClient.GetPR(entry.Branch)
	if err != nil || pr == nil {
		return
	}

	lastState, seen := t.prStates[entry.Folder]
	currentState := pr.State

	if !seen {
		// First time seeing this PR, just record state
		t.prStates[entry.Folder] = currentState
		return
	}

	if lastState == currentState {
		return
	}

	// State changed - emit event
	var eventType string
	switch currentState {
	case "OPEN":
		eventType = "pr_opened"
	case "MERGED":
		eventType = "pr_merged"
	case "CLOSED":
		eventType = "pr_closed"
	default:
		return
	}

	event := jsonl.ActivityEvent{
		Timestamp: time.Now().UTC(),
		Type:      eventType,
		Repo:      entry.Repo,
		Branch:    entry.Branch,
		PRNumber:  pr.Number,
		Title:     pr.Title,
	}
	t.store.AppendActivity(event)
	t.prStates[entry.Folder] = currentState
}

func (t *ActivityTracker) checkCommitActivity(folderPath string, entry jsonl.LocalEntry, repo *git.Repo) {
	commit, err := repo.HeadCommit()
	if err != nil {
		return
	}

	lastCommit, seen := t.commits[entry.Folder]
	if !seen {
		t.commits[entry.Folder] = commit
		return
	}

	if lastCommit == commit {
		return
	}

	// New commit detected
	msg, _ := repo.CommitMessage(commit)
	event := jsonl.ActivityEvent{
		Timestamp: time.Now().UTC(),
		Type:      "commit_pushed",
		Repo:      entry.Repo,
		Branch:    entry.Branch,
		Commit:    commit,
		Message:   msg,
	}
	t.store.AppendActivity(event)
	t.commits[entry.Folder] = commit
}
