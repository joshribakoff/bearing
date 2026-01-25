package daemon

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joshribakoff/bearing/internal/jsonl"
)

// HTTPServer serves the web dashboard API and static files
type HTTPServer struct {
	store     *jsonl.Store
	workspace string
	mu        sync.RWMutex
	staticFS  fs.FS
	clients   map[chan []byte]bool
	clientsMu sync.RWMutex
}

// NewHTTPServer creates a new HTTP server for the dashboard
func NewHTTPServer(store *jsonl.Store, workspace string, staticFS fs.FS) *HTTPServer {
	return &HTTPServer{
		store:     store,
		workspace: workspace,
		staticFS:  staticFS,
		clients:   make(map[chan []byte]bool),
	}
}

// Handler returns the http.Handler for the server
func (s *HTTPServer) Handler() http.Handler {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/projects", s.handleProjects)
	mux.HandleFunc("/api/worktrees", s.handleWorktrees)
	mux.HandleFunc("/api/plans", s.handlePlans)
	mux.HandleFunc("/api/issues", s.handleIssues)
	mux.HandleFunc("/api/prs", s.handlePRs)
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/events", s.handleEvents)

	// Static files
	if s.staticFS != nil {
		fileServer := http.FileServer(http.FS(s.staticFS))
		mux.Handle("/", fileServer)
	}

	return mux
}

// ProjectResponse for API
type ProjectResponse struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func (s *HTTPServer) handleProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	local, err := s.store.ReadLocal()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Count worktrees per repo
	counts := make(map[string]int)
	for _, e := range local {
		counts[e.Repo]++
	}

	// Build unique project list with counts
	projects := make([]ProjectResponse, 0)
	seen := make(map[string]bool)
	for _, e := range local {
		if !seen[e.Repo] {
			projects = append(projects, ProjectResponse{
				Name:  e.Repo,
				Count: counts[e.Repo],
			})
			seen[e.Repo] = true
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

// WorktreeResponse combines local and health data for API response
type WorktreeResponse struct {
	Folder   string  `json:"folder"`
	Repo     string  `json:"repo"`
	Branch   string  `json:"branch"`
	Base     bool    `json:"base"`
	Purpose  string  `json:"purpose,omitempty"`
	Status   string  `json:"status,omitempty"`
	Dirty    bool    `json:"dirty"`
	Unpushed int     `json:"unpushed"`
	PRState  *string `json:"prState,omitempty"`
}

func (s *HTTPServer) handleWorktrees(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	project := r.URL.Query().Get("project")

	s.mu.RLock()
	defer s.mu.RUnlock()

	local, err := s.store.ReadLocal()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	workflow, _ := s.store.ReadWorkflow()
	health, _ := s.store.ReadHealth()

	// Build lookup maps
	workflowMap := make(map[string]jsonl.WorkflowEntry)
	for _, wf := range workflow {
		key := wf.Repo + "/" + wf.Branch
		workflowMap[key] = wf
	}

	healthMap := make(map[string]jsonl.HealthEntry)
	for _, h := range health {
		healthMap[h.Folder] = h
	}

	// Combine data - always return array, not null
	resp := make([]WorktreeResponse, 0)
	for _, l := range local {
		if project != "" && l.Repo != project {
			continue
		}

		wt := WorktreeResponse{
			Folder: l.Folder,
			Repo:   l.Repo,
			Branch: l.Branch,
			Base:   l.Base,
		}

		if wf, ok := workflowMap[l.Repo+"/"+l.Branch]; ok {
			wt.Purpose = wf.Purpose
			wt.Status = wf.Status
		}

		if h, ok := healthMap[l.Folder]; ok {
			wt.Dirty = h.Dirty
			wt.Unpushed = h.Unpushed
			wt.PRState = h.PRState
		}

		resp = append(resp, wt)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// PlanResponse for API
type PlanResponse struct {
	Project string `json:"project"`
	Title   string `json:"title"`
	Issue   string `json:"issue,omitempty"`
	Status  string `json:"status"`
	Path    string `json:"path"`
}

// IssueResponse for API - plans with issue numbers
type IssueResponse struct {
	Number   int    `json:"number"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Priority int    `json:"priority"`
	Repo     string `json:"repo"`
	PlanID   string `json:"plan_id"`
}

func (s *HTTPServer) handlePlans(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	plansDir := filepath.Join(s.workspace, "plans")
	plans := make([]PlanResponse, 0)

	filepath.WalkDir(plansDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		rel, _ := filepath.Rel(plansDir, path)
		parts := strings.Split(rel, string(os.PathSeparator))
		if len(parts) < 2 {
			return nil
		}

		project := parts[0]
		fm := parsePlanFrontmatter(path)

		plans = append(plans, PlanResponse{
			Project: project,
			Title:   fm["title"],
			Issue:   fm["issue"],
			Status:  fm["status"],
			Path:    rel,
		})

		return nil
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plans)
}

func (s *HTTPServer) handleIssues(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	plansDir := filepath.Join(s.workspace, "plans")
	issues := make([]IssueResponse, 0)
	issueRe := regexp.MustCompile(`#?(\d+)`)

	filepath.WalkDir(plansDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		rel, _ := filepath.Rel(plansDir, path)
		parts := strings.Split(rel, string(os.PathSeparator))
		if len(parts) < 2 {
			return nil
		}

		project := parts[0]
		fm := parsePlanFrontmatter(path)

		// Only include plans with issue numbers
		if fm["issue"] == "" {
			return nil
		}

		// Parse issue number from formats like "#123" or "123"
		matches := issueRe.FindStringSubmatch(fm["issue"])
		if len(matches) < 2 {
			return nil
		}
		num, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil
		}

		// Parse priority (default to 0)
		priority := 0
		if fm["priority"] != "" {
			if p, err := strconv.Atoi(fm["priority"]); err == nil {
				priority = p
			}
		}

		// Extract plan ID from filename (e.g., "abc12-feature.md" -> "abc12")
		filename := filepath.Base(path)
		planID := strings.TrimSuffix(filename, ".md")

		issues = append(issues, IssueResponse{
			Number:   num,
			Title:    fm["title"],
			Status:   fm["status"],
			Priority: priority,
			Repo:     project,
			PlanID:   planID,
		})

		return nil
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(issues)
}

// PRResponse for API
type PRResponse struct {
	Folder string `json:"folder"`
	Repo   string `json:"repo"`
	Branch string `json:"branch"`
	State  string `json:"state"`
}

func (s *HTTPServer) handlePRs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	health, err := s.store.ReadHealth()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	local, _ := s.store.ReadLocal()
	localMap := make(map[string]jsonl.LocalEntry)
	for _, e := range local {
		localMap[e.Folder] = e
	}

	prs := make([]PRResponse, 0)
	for _, h := range health {
		if h.PRState == nil {
			continue
		}
		if l, ok := localMap[h.Folder]; ok {
			prs = append(prs, PRResponse{
				Folder: h.Folder,
				Repo:   l.Repo,
				Branch: l.Branch,
				State:  *h.PRState,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prs)
}

// HealthResponse for API
type HealthResponse struct {
	DaemonRunning bool      `json:"daemonRunning"`
	LastCheck     time.Time `json:"lastCheck"`
	WorktreeCount int       `json:"worktreeCount"`
}

func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	health, err := s.store.ReadHealth()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var lastCheck time.Time
	for _, h := range health {
		if h.LastCheck.After(lastCheck) {
			lastCheck = h.LastCheck
		}
	}

	resp := HealthResponse{
		DaemonRunning: true,
		LastCheck:     lastCheck,
		WorktreeCount: len(health),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// StatusResponse contains daemon status info
type StatusResponse struct {
	Running bool   `json:"running"`
	Version string `json:"version"`
}

func (s *HTTPServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := StatusResponse{
		Running: true,
		Version: "0.1.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleEvents implements Server-Sent Events for real-time updates
func (s *HTTPServer) handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	client := make(chan []byte, 10)
	s.clientsMu.Lock()
	s.clients[client] = true
	s.clientsMu.Unlock()

	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, client)
		s.clientsMu.Unlock()
		close(client)
	}()

	// Send initial connected event
	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"ok\"}\n\n")
	flusher.Flush()

	for {
		select {
		case msg := <-client:
			fmt.Fprintf(w, "event: update\ndata: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// Broadcast sends an event to all connected SSE clients
func (s *HTTPServer) Broadcast(eventType string, data interface{}) {
	msg, err := json.Marshal(map[string]interface{}{
		"type": eventType,
		"data": data,
	})
	if err != nil {
		return
	}

	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for client := range s.clients {
		select {
		case client <- msg:
		default:
			// Client buffer full, skip
		}
	}
}

// parsePlanFrontmatter parses YAML frontmatter from a plan file
func parsePlanFrontmatter(path string) map[string]string {
	content, err := os.ReadFile(path)
	if err != nil {
		return map[string]string{"title": filepath.Base(path)}
	}

	lines := strings.Split(string(content), "\n")
	fm := make(map[string]string)
	inFrontmatter := false

	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if !inFrontmatter && i == 0 {
				inFrontmatter = true
				continue
			} else if inFrontmatter {
				break
			}
		} else if inFrontmatter {
			if idx := strings.Index(line, ":"); idx > 0 {
				key := strings.TrimSpace(line[:idx])
				value := strings.TrimSpace(line[idx+1:])
				value = strings.Trim(value, "\"'")
				fm[key] = value
			}
		}
	}

	// Extract title from first heading if not in frontmatter
	if fm["title"] == "" {
		headingRe := regexp.MustCompile(`^#\s+(.+)$`)
		for _, line := range lines {
			if m := headingRe.FindStringSubmatch(line); len(m) > 1 {
				fm["title"] = m[1]
				break
			}
		}
	}

	if fm["title"] == "" {
		fm["title"] = filepath.Base(path)
	}
	if fm["status"] == "" {
		fm["status"] = "draft"
	}

	return fm
}

// EmbeddedStaticFS is a placeholder for the embedded web assets
var EmbeddedStaticFS embed.FS
