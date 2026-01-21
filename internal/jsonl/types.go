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
	LastCheck time.Time `json:"lastCheck"`
}

// ProjectEntry maps project names to GitHub repos in projects.jsonl
type ProjectEntry struct {
	Name       string `json:"name"`
	GitHubRepo string `json:"github_repo"`
	Path       string `json:"path"`
}
