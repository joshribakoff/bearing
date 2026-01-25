package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Smart Refresh Queue", "smart-refresh-queue"},
		{"Hello World", "hello-world"},
		{"Already-kebab", "already-kebab"},
		{"UPPERCASE", "uppercase"},
		{"Mixed CASE Words", "mixed-case-words"},
		{"  Spaces  Around  ", "spaces-around"},
		{"Special!@#Characters", "special-characters"},
		{"Multiple   Spaces", "multiple-spaces"},
		{"123 Numbers", "123-numbers"},
		{"a", "a"},
		{"", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := toKebabCase(tc.input)
			if result != tc.expected {
				t.Errorf("toKebabCase(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestGenerateShortID(t *testing.T) {
	id := generateShortID()
	if len(id) != 5 {
		t.Errorf("expected ID length 5, got %d", len(id))
	}

	// Check all chars are alphanumeric lowercase
	for _, c := range id {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			t.Errorf("ID contains invalid character: %c", c)
		}
	}

	// Generate multiple IDs to check they're different
	ids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		ids[generateShortID()] = true
	}
	if len(ids) < 5 {
		t.Errorf("expected mostly unique IDs, got %d unique out of 10", len(ids))
	}
}

func TestRunPlanCreate(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Set up the project flag
	planCreateProject = "testproject"

	// Run the command
	err := runPlanCreate(nil, []string{"Test Plan Title"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that a file was created in the correct directory
	planDir := filepath.Join(tmpDir, "Projects", "plans", "testproject")
	entries, err := os.ReadDir(planDir)
	if err != nil {
		t.Fatalf("failed to read plan directory: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entries))
	}

	filename := entries[0].Name()
	if !strings.HasSuffix(filename, "-test-plan-title.md") {
		t.Errorf("expected filename ending with '-test-plan-title.md', got %s", filename)
	}
	if len(filename) != len("xxxxx-test-plan-title.md") {
		t.Errorf("unexpected filename length: %s", filename)
	}

	// Check file content
	content, err := os.ReadFile(filepath.Join(planDir, filename))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "repo: testproject") {
		t.Error("expected content to contain 'repo: testproject'")
	}
	if !strings.Contains(contentStr, "status: draft") {
		t.Error("expected content to contain 'status: draft'")
	}
	if !strings.Contains(contentStr, "# Test Plan Title") {
		t.Error("expected content to contain '# Test Plan Title'")
	}
	if !strings.Contains(contentStr, "id: ") {
		t.Error("expected content to contain 'id: '")
	}
}

func TestRunPlanCreate_EmptyProject(t *testing.T) {
	planCreateProject = ""
	err := runPlanCreate(nil, []string{"Some Title"})
	if err == nil {
		t.Error("expected error for empty project")
	}
	if !strings.Contains(err.Error(), "project name cannot be empty") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunPlanCreate_EmptyTitle(t *testing.T) {
	planCreateProject = "myproject"
	err := runPlanCreate(nil, []string{"   "})
	if err == nil {
		t.Error("expected error for empty title")
	}
	if !strings.Contains(err.Error(), "title cannot be empty") {
		t.Errorf("unexpected error message: %v", err)
	}
}
