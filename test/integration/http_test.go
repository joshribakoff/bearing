package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/joshribakoff/bearing/internal/daemon"
	"github.com/joshribakoff/bearing/internal/jsonl"
)

// TestHTTPServerIntegration tests the HTTP server with real JSONL files
func TestHTTPServerIntegration(t *testing.T) {
	// Create temp workspace
	workspace := t.TempDir()

	// Create realistic local.jsonl
	localData := []jsonl.LocalEntry{
		{Folder: "project-main", Repo: "project", Branch: "main", Base: true},
		{Folder: "project-feature", Repo: "project", Branch: "feature/add-tests", Base: false},
		{Folder: "other-main", Repo: "other", Branch: "main", Base: true},
	}
	writeJSONL(t, filepath.Join(workspace, "local.jsonl"), localData)

	// Create realistic health.jsonl
	healthData := []jsonl.HealthEntry{
		{Folder: "project-main", Dirty: false, Unpushed: 0},
		{Folder: "project-feature", Dirty: true, Unpushed: 3},
		{Folder: "other-main", Dirty: false, Unpushed: 0},
	}
	writeJSONL(t, filepath.Join(workspace, "health.jsonl"), healthData)

	// Create plans directory
	plansDir := filepath.Join(workspace, "plans", "project")
	if err := os.MkdirAll(plansDir, 0755); err != nil {
		t.Fatal(err)
	}
	planContent := `---
title: Add Testing Infrastructure
status: in-progress
issue: "#42"
---

# Add Testing Infrastructure

This plan covers adding comprehensive tests.
`
	if err := os.WriteFile(filepath.Join(plansDir, "001-testing.md"), []byte(planContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create static files
	staticDir := filepath.Join(workspace, "web")
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		t.Fatal(err)
	}
	indexHTML := `<!DOCTYPE html>
<html>
<head><title>Bearing Dashboard</title></head>
<body><h1>Bearing</h1></body>
</html>`
	if err := os.WriteFile(filepath.Join(staticDir, "index.html"), []byte(indexHTML), 0644); err != nil {
		t.Fatal(err)
	}

	// Create server
	store := jsonl.NewStore(workspace)
	server := daemon.NewHTTPServer(store, workspace, os.DirFS(staticDir))
	ts := httptest.NewServer(server.Handler())
	defer ts.Close()

	t.Run("serves index.html", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/index.html")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		if len(body) == 0 {
			t.Error("expected non-empty body")
		}
		if !contains(string(body), "Bearing") {
			t.Error("expected body to contain 'Bearing'")
		}
	})

	t.Run("worktrees returns combined data", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/worktrees")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		var worktrees []daemon.WorktreeResponse
		if err := json.NewDecoder(resp.Body).Decode(&worktrees); err != nil {
			t.Fatal(err)
		}

		if len(worktrees) != 3 {
			t.Fatalf("expected 3 worktrees, got %d", len(worktrees))
		}

		// Find feature worktree and check health data is merged
		for _, wt := range worktrees {
			if wt.Folder == "project-feature" {
				if !wt.Dirty {
					t.Error("expected project-feature to be dirty")
				}
				if wt.Unpushed != 3 {
					t.Errorf("expected 3 unpushed, got %d", wt.Unpushed)
				}
			}
		}
	})

	t.Run("worktrees filters by project", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/worktrees?project=project")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		var worktrees []daemon.WorktreeResponse
		if err := json.NewDecoder(resp.Body).Decode(&worktrees); err != nil {
			t.Fatal(err)
		}

		if len(worktrees) != 2 {
			t.Fatalf("expected 2 worktrees for 'project', got %d", len(worktrees))
		}

		for _, wt := range worktrees {
			if wt.Repo != "project" {
				t.Errorf("expected repo 'project', got '%s'", wt.Repo)
			}
		}
	})

	t.Run("projects returns unique projects with counts", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/projects")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		var projects []daemon.ProjectResponse
		if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
			t.Fatal(err)
		}

		if len(projects) != 2 {
			t.Fatalf("expected 2 projects, got %d", len(projects))
		}

		// Check counts
		for _, p := range projects {
			if p.Name == "project" && p.Count != 2 {
				t.Errorf("expected count 2 for 'project', got %d", p.Count)
			}
			if p.Name == "other" && p.Count != 1 {
				t.Errorf("expected count 1 for 'other', got %d", p.Count)
			}
		}
	})

	t.Run("plans returns plan list", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/plans")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		var plans []daemon.PlanResponse
		if err := json.NewDecoder(resp.Body).Decode(&plans); err != nil {
			t.Fatal(err)
		}

		if len(plans) != 1 {
			t.Fatalf("expected 1 plan, got %d", len(plans))
		}

		if plans[0].Title != "Add Testing Infrastructure" {
			t.Errorf("expected title 'Add Testing Infrastructure', got '%s'", plans[0].Title)
		}
		if plans[0].Status != "in-progress" {
			t.Errorf("expected status 'in-progress', got '%s'", plans[0].Status)
		}
	})

	t.Run("health returns aggregated health", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/health")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		var health daemon.HealthResponse
		if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
			t.Fatal(err)
		}

		if !health.DaemonRunning {
			t.Error("expected daemon running to be true")
		}
		if health.WorktreeCount != 3 {
			t.Errorf("expected 3 worktrees, got %d", health.WorktreeCount)
		}
	})
}

// TestHTTPServerEmptyWorkspace tests behavior with no data
func TestHTTPServerEmptyWorkspace(t *testing.T) {
	workspace := t.TempDir()
	store := jsonl.NewStore(workspace)
	server := daemon.NewHTTPServer(store, workspace, nil)
	ts := httptest.NewServer(server.Handler())
	defer ts.Close()

	t.Run("worktrees returns empty array", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/worktrees")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		// Should be empty array, not null
		if string(body) != "[]\n" {
			t.Errorf("expected '[]\\n', got '%s'", string(body))
		}
	})

	t.Run("projects returns empty array", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/projects")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "[]\n" {
			t.Errorf("expected '[]\\n', got '%s'", string(body))
		}
	})

	t.Run("plans returns empty array", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/plans")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "[]\n" {
			t.Errorf("expected '[]\\n', got '%s'", string(body))
		}
	})
}

func writeJSONL[T any](t *testing.T, path string, entries []T) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			t.Fatal(err)
		}
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
