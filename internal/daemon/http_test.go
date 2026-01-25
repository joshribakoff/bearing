package daemon

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/joshribakoff/bearing/internal/jsonl"
)

func setupTestStore(t *testing.T) (*jsonl.Store, string) {
	t.Helper()
	dir := t.TempDir()

	// Create test local.jsonl
	localPath := filepath.Join(dir, "local.jsonl")
	localData := `{"folder":"project-main","repo":"project","branch":"main","base":true}
{"folder":"project-feature","repo":"project","branch":"feature-1","base":false}
`
	if err := os.WriteFile(localPath, []byte(localData), 0644); err != nil {
		t.Fatal(err)
	}

	// Create test health.jsonl
	healthPath := filepath.Join(dir, "health.jsonl")
	healthData := `{"folder":"project-main","dirty":false,"unpushed":0,"lastCheck":"2024-01-01T00:00:00Z"}
{"folder":"project-feature","dirty":true,"unpushed":2,"prState":"OPEN","lastCheck":"2024-01-01T00:00:00Z"}
`
	if err := os.WriteFile(healthPath, []byte(healthData), 0644); err != nil {
		t.Fatal(err)
	}

	// Create test projects.jsonl
	projectsPath := filepath.Join(dir, "projects.jsonl")
	projectsData := `{"name":"project","github_repo":"user/project","path":"/path/to/project"}
`
	if err := os.WriteFile(projectsPath, []byte(projectsData), 0644); err != nil {
		t.Fatal(err)
	}

	return jsonl.NewStore(dir), dir
}

func TestHandleWorktrees(t *testing.T) {
	store, dir := setupTestStore(t)
	server := NewHTTPServer(store, dir, nil)
	handler := server.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/worktrees", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}

	var resp []WorktreeResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(resp))
	}

	// Check first worktree (base)
	if resp[0].Folder != "project-main" {
		t.Errorf("expected folder project-main, got %s", resp[0].Folder)
	}
	if !resp[0].Base {
		t.Error("expected first worktree to be base")
	}
	if resp[0].Dirty {
		t.Error("expected first worktree to not be dirty")
	}

	// Check second worktree (feature)
	if resp[1].Folder != "project-feature" {
		t.Errorf("expected folder project-feature, got %s", resp[1].Folder)
	}
	if resp[1].Base {
		t.Error("expected second worktree to not be base")
	}
	if !resp[1].Dirty {
		t.Error("expected second worktree to be dirty")
	}
	if resp[1].Unpushed != 2 {
		t.Errorf("expected 2 unpushed, got %d", resp[1].Unpushed)
	}
	if resp[1].PRState == nil || *resp[1].PRState != "OPEN" {
		t.Error("expected PRState to be OPEN")
	}
}

func TestHandleWorktrees_MethodNotAllowed(t *testing.T) {
	store, dir := setupTestStore(t)
	server := NewHTTPServer(store, dir, nil)
	handler := server.Handler()

	req := httptest.NewRequest(http.MethodPost, "/api/worktrees", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rec.Code)
	}
}

func TestHandleWorktrees_FilterByProject(t *testing.T) {
	store, dir := setupTestStore(t)
	server := NewHTTPServer(store, dir, nil)
	handler := server.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/worktrees?project=project", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp []WorktreeResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp) != 2 {
		t.Fatalf("expected 2 worktrees for project, got %d", len(resp))
	}

	// Test with non-existent project
	req = httptest.NewRequest(http.MethodGet, "/api/worktrees?project=nonexistent", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp) != 0 {
		t.Errorf("expected 0 worktrees for nonexistent project, got %d", len(resp))
	}
}

func TestHandleHealth(t *testing.T) {
	store, dir := setupTestStore(t)
	server := NewHTTPServer(store, dir, nil)
	handler := server.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.DaemonRunning {
		t.Error("expected daemon running to be true")
	}
	if resp.WorktreeCount != 2 {
		t.Errorf("expected 2 worktrees, got %d", resp.WorktreeCount)
	}
}

func TestHandleProjects(t *testing.T) {
	store, dir := setupTestStore(t)
	server := NewHTTPServer(store, dir, nil)
	handler := server.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp []ProjectResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp) != 1 {
		t.Fatalf("expected 1 project, got %d", len(resp))
	}

	if resp[0].Name != "project" {
		t.Errorf("expected project name 'project', got %s", resp[0].Name)
	}
	if resp[0].Count != 2 {
		t.Errorf("expected count 2, got %d", resp[0].Count)
	}
}

func TestHandlePRs(t *testing.T) {
	store, dir := setupTestStore(t)
	server := NewHTTPServer(store, dir, nil)
	handler := server.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/prs", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp []PRResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Only project-feature has a PR
	if len(resp) != 1 {
		t.Fatalf("expected 1 PR, got %d", len(resp))
	}

	if resp[0].State != "OPEN" {
		t.Errorf("expected PR state OPEN, got %s", resp[0].State)
	}
}

func TestHandlePlans(t *testing.T) {
	store, dir := setupTestStore(t)

	// Create plans directory with a test plan
	plansDir := filepath.Join(dir, "plans", "testproject")
	if err := os.MkdirAll(plansDir, 0755); err != nil {
		t.Fatal(err)
	}
	planContent := `---
title: Test Plan
status: active
issue: "#123"
---

# Test Plan

This is a test plan.
`
	if err := os.WriteFile(filepath.Join(plansDir, "001-test.md"), []byte(planContent), 0644); err != nil {
		t.Fatal(err)
	}

	server := NewHTTPServer(store, dir, nil)
	handler := server.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/plans", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp []PlanResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp) != 1 {
		t.Fatalf("expected 1 plan, got %d", len(resp))
	}

	if resp[0].Title != "Test Plan" {
		t.Errorf("expected title 'Test Plan', got %s", resp[0].Title)
	}
	if resp[0].Status != "active" {
		t.Errorf("expected status 'active', got %s", resp[0].Status)
	}
	if resp[0].Project != "testproject" {
		t.Errorf("expected project 'testproject', got %s", resp[0].Project)
	}
}

func TestHandleIssues(t *testing.T) {
	store, dir := setupTestStore(t)

	// Create plans directory with test plans
	plansDir := filepath.Join(dir, "plans", "myrepo")
	if err := os.MkdirAll(plansDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Plan with issue number
	planWithIssue := `---
title: Feature Plan
status: active
issue: "#42"
priority: 1
---

# Feature Plan
`
	if err := os.WriteFile(filepath.Join(plansDir, "abc12-feature.md"), []byte(planWithIssue), 0644); err != nil {
		t.Fatal(err)
	}

	// Plan without issue number (should be excluded)
	planNoIssue := `---
title: Draft Plan
status: draft
---

# Draft Plan
`
	if err := os.WriteFile(filepath.Join(plansDir, "xyz99-draft.md"), []byte(planNoIssue), 0644); err != nil {
		t.Fatal(err)
	}

	server := NewHTTPServer(store, dir, nil)
	handler := server.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/issues", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp []IssueResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Only plan with issue should be included
	if len(resp) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(resp))
	}

	if resp[0].Number != 42 {
		t.Errorf("expected issue number 42, got %d", resp[0].Number)
	}
	if resp[0].Title != "Feature Plan" {
		t.Errorf("expected title 'Feature Plan', got %s", resp[0].Title)
	}
	if resp[0].Status != "active" {
		t.Errorf("expected status 'active', got %s", resp[0].Status)
	}
	if resp[0].Priority != 1 {
		t.Errorf("expected priority 1, got %d", resp[0].Priority)
	}
	if resp[0].Repo != "myrepo" {
		t.Errorf("expected repo 'myrepo', got %s", resp[0].Repo)
	}
	if resp[0].PlanID != "abc12-feature" {
		t.Errorf("expected plan_id 'abc12-feature', got %s", resp[0].PlanID)
	}
}

func TestHandleStatus(t *testing.T) {
	store, dir := setupTestStore(t)
	server := NewHTTPServer(store, dir, nil)
	handler := server.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp StatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Running {
		t.Error("expected running to be true")
	}
	if resp.Version == "" {
		t.Error("expected version to be set")
	}
}

func TestStaticFileServing(t *testing.T) {
	store, dir := setupTestStore(t)

	// Create static files
	staticDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("<!DOCTYPE html><html><body>Hello</body></html>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staticDir, "app.js"), []byte("console.log('hello');"), 0644); err != nil {
		t.Fatal(err)
	}

	server := NewHTTPServer(store, dir, os.DirFS(staticDir))

	// Use httptest.Server for proper request handling
	ts := httptest.NewServer(server.Handler())
	defer ts.Close()

	// Test index.html
	resp, err := http.Get(ts.URL + "/index.html")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "<!DOCTYPE html><html><body>Hello</body></html>" {
		t.Errorf("unexpected body: %s", string(body))
	}

	// Test app.js
	resp2, err := http.Get(ts.URL + "/app.js")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp2.StatusCode)
	}
}

func TestEmptyStore(t *testing.T) {
	// Test with empty directory (no JSONL files)
	dir := t.TempDir()
	store := jsonl.NewStore(dir)
	server := NewHTTPServer(store, dir, nil)
	handler := server.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/worktrees", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp []WorktreeResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Should return empty array, not null
	if len(resp) != 0 {
		t.Errorf("expected empty array, got %d items", len(resp))
	}
}

func TestHandleEvents_SSE(t *testing.T) {
	store, dir := setupTestStore(t)
	server := NewHTTPServer(store, dir, nil)

	ts := httptest.NewServer(server.Handler())
	defer ts.Close()

	// Make SSE request
	resp, err := http.Get(ts.URL + "/api/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("expected Content-Type text/event-stream, got %s", ct)
	}

	// Read initial connected event
	buf := make([]byte, 1024)
	n, err := resp.Body.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	data := string(buf[:n])
	if !contains(data, "event: connected") {
		t.Errorf("expected connected event, got: %s", data)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestParsePlanFrontmatter(t *testing.T) {
	dir := t.TempDir()

	// Test with frontmatter
	planWithFM := `---
title: My Plan
status: active
issue: "#42"
---

# Content
`
	path1 := filepath.Join(dir, "plan1.md")
	if err := os.WriteFile(path1, []byte(planWithFM), 0644); err != nil {
		t.Fatal(err)
	}

	fm := parsePlanFrontmatter(path1)
	if fm["title"] != "My Plan" {
		t.Errorf("expected title 'My Plan', got %s", fm["title"])
	}
	if fm["status"] != "active" {
		t.Errorf("expected status 'active', got %s", fm["status"])
	}

	// Test without frontmatter (extract from heading)
	planNoFM := `# Heading Title

Some content.
`
	path2 := filepath.Join(dir, "plan2.md")
	if err := os.WriteFile(path2, []byte(planNoFM), 0644); err != nil {
		t.Fatal(err)
	}

	fm2 := parsePlanFrontmatter(path2)
	if fm2["title"] != "Heading Title" {
		t.Errorf("expected title 'Heading Title', got %s", fm2["title"])
	}
	if fm2["status"] != "draft" {
		t.Errorf("expected default status 'draft', got %s", fm2["status"])
	}
}
