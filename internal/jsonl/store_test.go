package jsonl

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadWriteWorkflow(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	// Write entries
	entries := []WorkflowEntry{
		{Repo: "myrepo", Branch: "feature-a", Status: "active", Created: time.Now()},
		{Repo: "myrepo", Branch: "feature-b", Status: "merged", Created: time.Now()},
	}
	if err := store.WriteWorkflow(entries); err != nil {
		t.Fatal(err)
	}

	// Read back
	got, err := store.ReadWorkflow()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 entries, got %d", len(got))
	}
	if got[0].Branch != "feature-a" {
		t.Errorf("expected feature-a, got %s", got[0].Branch)
	}
}

func TestReadWriteLocal(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	entries := []LocalEntry{
		{Folder: "myrepo", Repo: "myrepo", Branch: "main", Base: true},
		{Folder: "myrepo-feature", Repo: "myrepo", Branch: "feature", Base: false},
	}
	if err := store.WriteLocal(entries); err != nil {
		t.Fatal(err)
	}

	got, err := store.ReadLocal()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 entries, got %d", len(got))
	}
	if !got[0].Base {
		t.Error("expected first entry to be base")
	}
}

func TestAppendWorkflow(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	// Append first
	store.AppendWorkflow(WorkflowEntry{Repo: "a", Branch: "x", Status: "active"})
	store.AppendWorkflow(WorkflowEntry{Repo: "b", Branch: "y", Status: "active"})

	got, _ := store.ReadWorkflow()
	if len(got) != 2 {
		t.Errorf("expected 2 entries, got %d", len(got))
	}
}

func TestReadNonexistent(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	// Should return empty, not error
	got, err := store.ReadWorkflow()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 entries, got %d", len(got))
	}
}

func TestReadMalformedLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "workflow.jsonl")

	// Write some valid and invalid JSON
	content := `{"repo":"good","branch":"a","status":"active"}
not json at all
{"repo":"also-good","branch":"b","status":"merged"}
`
	os.WriteFile(path, []byte(content), 0644)

	store := NewStore(dir)
	got, err := store.ReadWorkflow()
	if err != nil {
		t.Fatal(err)
	}
	// Should skip malformed line, return 2 valid entries
	if len(got) != 2 {
		t.Errorf("expected 2 valid entries, got %d", len(got))
	}
}
