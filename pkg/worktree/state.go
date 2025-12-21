package worktree

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
)

// LocalEntry represents a worktree on this machine
type LocalEntry struct {
	Folder string `json:"folder"`
	Repo   string `json:"repo"`
	Branch string `json:"branch"`
	Base   bool   `json:"base"`
}

// WorkflowEntry represents a branch being worked on (portable)
type WorkflowEntry struct {
	Repo   string `json:"repo"`
	Branch string `json:"branch"`
	Status string `json:"status"` // in_progress, completed
}

// State manages the JSONL state files
type State struct {
	root string
}

// NewState creates a state manager for the given workspace root
func NewState(root string) *State {
	return &State{root: root}
}

func (s *State) localPath() string {
	return filepath.Join(s.root, "local.jsonl")
}

func (s *State) workflowPath() string {
	return filepath.Join(s.root, "workflow.jsonl")
}

// ReadLocal reads all local entries
func (s *State) ReadLocal() ([]LocalEntry, error) {
	return readJSONL[LocalEntry](s.localPath())
}

// ReadWorkflow reads all workflow entries
func (s *State) ReadWorkflow() ([]WorkflowEntry, error) {
	return readJSONL[WorkflowEntry](s.workflowPath())
}

// WriteLocal writes all local entries
func (s *State) WriteLocal(entries []LocalEntry) error {
	return writeJSONL(s.localPath(), entries)
}

// WriteWorkflow writes all workflow entries
func (s *State) WriteWorkflow(entries []WorkflowEntry) error {
	return writeJSONL(s.workflowPath(), entries)
}

// AppendLocal appends a local entry
func (s *State) AppendLocal(entry LocalEntry) error {
	return appendJSONL(s.localPath(), entry)
}

// AppendWorkflow appends a workflow entry
func (s *State) AppendWorkflow(entry WorkflowEntry) error {
	return appendJSONL(s.workflowPath(), entry)
}

// FindLocal finds a local entry by folder name
func (s *State) FindLocal(folder string) (*LocalEntry, error) {
	entries, err := s.ReadLocal()
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.Folder == folder {
			return &e, nil
		}
	}
	return nil, nil
}

// FindWorkflow finds a workflow entry by repo and branch
func (s *State) FindWorkflow(repo, branch string) (*WorkflowEntry, error) {
	entries, err := s.ReadWorkflow()
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.Repo == repo && e.Branch == branch {
			return &e, nil
		}
	}
	return nil, nil
}

// RemoveLocal removes a local entry by folder name
func (s *State) RemoveLocal(folder string) error {
	entries, err := s.ReadLocal()
	if err != nil {
		return err
	}
	var filtered []LocalEntry
	for _, e := range entries {
		if e.Folder != folder {
			filtered = append(filtered, e)
		}
	}
	return s.WriteLocal(filtered)
}

// UpdateWorkflowStatus updates the status of a workflow entry
func (s *State) UpdateWorkflowStatus(repo, branch, status string) error {
	entries, err := s.ReadWorkflow()
	if err != nil {
		return err
	}
	for i, e := range entries {
		if e.Repo == repo && e.Branch == branch {
			entries[i].Status = status
		}
	}
	return s.WriteWorkflow(entries)
}

func readJSONL[T any](path string) ([]T, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var results []T
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var item T
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, scanner.Err()
}

func writeJSONL[T any](path string, items []T) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			return err
		}
		f.Write(data)
		f.WriteString("\n")
	}
	return nil
}

func appendJSONL[T any](path string, item T) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	f.Write(data)
	f.WriteString("\n")
	return nil
}
