package jsonl

import "time"

// WorkflowEntry tracks worktree lifecycle in workflow.jsonl
type WorkflowEntry struct {
	Repo    string    `json:"repo"`
	Branch  string    `json:"branch"`
	BasedOn string    `json:"basedOn,omitempty"`
	Purpose string    `json:"purpose,omitempty"`
	Status  string    `json:"status"` // active, merged, abandoned
	Created time.Time `json:"created"`
}

// LocalEntry tracks local worktree folders in local.jsonl
type LocalEntry struct {
	Folder string `json:"folder"`
	Repo   string `json:"repo"`
	Branch string `json:"branch"`
	Base   bool   `json:"base"`
}

// HealthEntry tracks worktree health status in health.jsonl
type HealthEntry struct {
	Folder    string    `json:"folder"`
	Dirty     bool      `json:"dirty"`
	Unpushed  int       `json:"unpushed"`
	PRState   *string   `json:"prState,omitempty"`
	PRTitle   *string   `json:"prTitle,omitempty"`
	LastCheck time.Time `json:"lastCheck"`
}

// ProjectEntry maps project names to GitHub repos in projects.jsonl
type ProjectEntry struct {
	Name       string `json:"name"`
	GitHubRepo string `json:"github_repo"`
	Path       string `json:"path"`
}

// ActivityEvent tracks workspace activity in activity.jsonl
type ActivityEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // pr_opened, pr_merged, pr_closed, commit_pushed, test_pass, test_fail
	Repo      string    `json:"repo"`
	Branch    string    `json:"branch"`
	PRNumber  int       `json:"pr_number,omitempty"`
	Title     string    `json:"title,omitempty"`
	Commit    string    `json:"commit,omitempty"`
	Message   string    `json:"message,omitempty"`
}
