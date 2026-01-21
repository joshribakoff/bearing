package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParsePlanFile_Basic(t *testing.T) {
	content := `---
issue: 123
repo: myrepo
status: in-progress
---
# Plan Title

Some body content.
`
	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	fm, body, err := parsePlanFile(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Issue != "123" {
		t.Errorf("expected issue 123, got %s", fm.Issue)
	}
	if fm.Repo != "myrepo" {
		t.Errorf("expected repo myrepo, got %s", fm.Repo)
	}
	if fm.Status != "in-progress" {
		t.Errorf("expected status in-progress, got %s", fm.Status)
	}
	if !strings.Contains(body, "# Plan Title") {
		t.Errorf("body should contain title, got %s", body)
	}
}

func TestParsePlanFile_StripQuotes(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "double quotes",
			content: `---
issue: "456"
repo: myrepo
---
body`,
			expected: "456",
		},
		{
			name: "single quotes",
			content: `---
issue: '789'
repo: myrepo
---
body`,
			expected: "789",
		},
		{
			name: "no quotes",
			content: `---
issue: 101
repo: myrepo
---
body`,
			expected: "101",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile := createTempFile(t, tc.content)
			defer os.Remove(tmpFile)

			fm, _, err := parsePlanFile(tmpFile)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if fm.Issue != tc.expected {
				t.Errorf("expected issue %s, got %s", tc.expected, fm.Issue)
			}
		})
	}
}

func TestParsePlanFile_NonNumericIssue(t *testing.T) {
	tests := []struct {
		name    string
		issue   string
		wantErr bool
	}{
		{"numeric", "123", false},
		{"alpha", "abc", true},
		{"mixed", "123abc", true},
		{"empty", "", false}, // empty is valid (triggers auto-create)
		{"with hash", "#123", true},
		{"float", "12.34", true},
		{"negative", "-5", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			content := "---\nissue: " + tc.issue + "\nrepo: myrepo\n---\nbody"
			tmpFile := createTempFile(t, content)
			defer os.Remove(tmpFile)

			_, _, err := parsePlanFile(tmpFile)
			if tc.wantErr && err == nil {
				t.Errorf("expected error for issue %q, got nil", tc.issue)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error for issue %q: %v", tc.issue, err)
			}
		})
	}
}

func TestParsePlanFile_NullBytesRejected(t *testing.T) {
	content := "---\nissue: 123\nrepo: my\x00repo\n---\nbody"
	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	_, _, err := parsePlanFile(tmpFile)
	if err == nil {
		t.Error("expected error for null byte in frontmatter")
	}
	if !strings.Contains(err.Error(), "control characters") {
		t.Errorf("error should mention control characters, got: %v", err)
	}
}

func TestParsePlanFile_ControlCharsRejected(t *testing.T) {
	// Test with bell character
	content := "---\nissue: 123\nrepo: my\x07repo\n---\nbody"
	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	_, _, err := parsePlanFile(tmpFile)
	if err == nil {
		t.Error("expected error for control character in frontmatter")
	}
}

func TestParsePlanFile_LongLines(t *testing.T) {
	// Create a line longer than the default 64KB scanner buffer
	longValue := strings.Repeat("x", 100*1024)
	content := "---\nissue: 123\nrepo: myrepo\nstatus: " + longValue + "\n---\nbody"
	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	fm, _, err := parsePlanFile(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error for long line: %v", err)
	}
	if fm.Status != longValue {
		t.Errorf("expected status to be preserved, got length %d", len(fm.Status))
	}
}

func TestStripQuotes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`'hello'`, "hello"},
		{`hello`, "hello"},
		{`"hello`, `"hello`},
		{`hello"`, `hello"`},
		{`""`, ""},
		{`''`, ""},
		{`"`, `"`},
		{``, ``},
	}

	for _, tc := range tests {
		result := stripQuotes(tc.input)
		if result != tc.expected {
			t.Errorf("stripQuotes(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestContainsControlChars(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"hello", false},
		{"hello world", false},
		{"hello\tworld", false}, // tab is allowed
		{"hello\x00world", true},
		{"hello\x07world", true}, // bell
		{"hello\nworld", true},   // newline is control char
		{"", false},
	}

	for _, tc := range tests {
		result := containsControlChars(tc.input)
		if result != tc.expected {
			t.Errorf("containsControlChars(%q) = %v, want %v", tc.input, result, tc.expected)
		}
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"0", true},
		{"", false},
		{"abc", false},
		{"12.34", false},
		{"-5", false},
		{"123abc", false},
	}

	for _, tc := range tests {
		result := isNumeric(tc.input)
		if result != tc.expected {
			t.Errorf("isNumeric(%q) = %v, want %v", tc.input, result, tc.expected)
		}
	}
}

func createTempFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-plan.md")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return tmpFile
}
