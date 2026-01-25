package jsonl

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
)

// Store manages JSONL file operations with locking
type Store struct {
	baseDir string
}

// NewStore creates a store for the given workspace directory
func NewStore(baseDir string) *Store {
	return &Store{baseDir: baseDir}
}

// WorkflowPath returns the path to workflow.jsonl
func (s *Store) WorkflowPath() string {
	return filepath.Join(s.baseDir, "workflow.jsonl")
}

// LocalPath returns the path to local.jsonl
func (s *Store) LocalPath() string {
	return filepath.Join(s.baseDir, "local.jsonl")
}

// HealthPath returns the path to health.jsonl
func (s *Store) HealthPath() string {
	return filepath.Join(s.baseDir, "health.jsonl")
}

// ProjectsPath returns the path to projects.jsonl
func (s *Store) ProjectsPath() string {
	return filepath.Join(s.baseDir, "projects.jsonl")
}

// ReadWorkflow reads all workflow entries
func (s *Store) ReadWorkflow() ([]WorkflowEntry, error) {
	return readJSONL[WorkflowEntry](s.WorkflowPath())
}

// ReadLocal reads all local entries
func (s *Store) ReadLocal() ([]LocalEntry, error) {
	return readJSONL[LocalEntry](s.LocalPath())
}

// ReadHealth reads all health entries
func (s *Store) ReadHealth() ([]HealthEntry, error) {
	return readJSONL[HealthEntry](s.HealthPath())
}

// ReadProjects reads all project entries
func (s *Store) ReadProjects() ([]ProjectEntry, error) {
	return readJSONL[ProjectEntry](s.ProjectsPath())
}

// WriteWorkflow writes all workflow entries (overwrites)
func (s *Store) WriteWorkflow(entries []WorkflowEntry) error {
	return writeJSONL(s.WorkflowPath(), entries)
}

// WriteLocal writes all local entries (overwrites)
func (s *Store) WriteLocal(entries []LocalEntry) error {
	return writeJSONL(s.LocalPath(), entries)
}

// WriteHealth writes all health entries (overwrites)
func (s *Store) WriteHealth(entries []HealthEntry) error {
	return writeJSONL(s.HealthPath(), entries)
}

// AppendWorkflow appends a workflow entry
func (s *Store) AppendWorkflow(entry WorkflowEntry) error {
	return appendJSONL(s.WorkflowPath(), entry)
}

// AppendLocal appends a local entry
func (s *Store) AppendLocal(entry LocalEntry) error {
	return appendJSONL(s.LocalPath(), entry)
}

func readJSONL[T any](path string) ([]T, error) {
	lock, err := NewFileLock(path)
	if err != nil {
		return nil, err
	}
	defer lock.Close()

	if err := lock.RLock(); err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []T
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var entry T
		if err := json.Unmarshal(line, &entry); err != nil {
			continue // skip malformed lines
		}
		entries = append(entries, entry)
	}
	return entries, scanner.Err()
}

func writeJSONL[T any](path string, entries []T) error {
	lock, err := NewFileLock(path)
	if err != nil {
		return err
	}
	defer lock.Close()

	if err := lock.Lock(); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, entry := range entries {
		if err := enc.Encode(entry); err != nil {
			return err
		}
	}
	return nil
}

func appendJSONL[T any](path string, entry T) error {
	lock, err := NewFileLock(path)
	if err != nil {
		return err
	}
	defer lock.Close()

	if err := lock.Lock(); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(entry)
}
