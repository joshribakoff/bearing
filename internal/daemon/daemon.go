package daemon

import (
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/joshribakoff/bearing/internal/gh"
	"github.com/joshribakoff/bearing/internal/git"
	"github.com/joshribakoff/bearing/internal/jsonl"
)

// DefaultHTTPPort is the preferred port for the HTTP server
const DefaultHTTPPort = 8374

// Config holds daemon configuration
type Config struct {
	WorkspaceDir string
	BearingDir   string
	Interval     time.Duration
	StaticFS     fs.FS // Optional: embedded static files for web dashboard
}

// Daemon manages the health monitoring background process
type Daemon struct {
	config     Config
	stop       chan struct{}
	httpServer *HTTPServer
}

// New creates a new daemon instance
func New(config Config) *Daemon {
	return &Daemon{
		config: config,
		stop:   make(chan struct{}),
	}
}

// PIDFile returns the path to the PID file
func (d *Daemon) PIDFile() string {
	return filepath.Join(d.config.BearingDir, "bearing.pid")
}

// LogFile returns the path to the log file
func (d *Daemon) LogFile() string {
	return filepath.Join(d.config.BearingDir, "daemon.log")
}

// IsRunning checks if the daemon is currently running
func (d *Daemon) IsRunning() (bool, int) {
	data, err := os.ReadFile(d.PIDFile())
	if err != nil {
		return false, 0
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return false, 0
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, 0
	}

	// Signal 0 tests if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil, pid
}

// Start starts the daemon (foreground or background)
func (d *Daemon) Start(foreground bool) error {
	if d.config.Interval <= 0 {
		return fmt.Errorf("interval must be positive")
	}

	// Ensure bearing dir exists
	if err := os.MkdirAll(d.config.BearingDir, 0755); err != nil {
		return err
	}

	// Check if already running
	if running, pid := d.IsRunning(); running {
		return fmt.Errorf("daemon already running with PID %d", pid)
	}

	if foreground {
		return d.run()
	}

	// Fork to background
	return d.startBackground()
}

func (d *Daemon) startBackground() error {
	// Re-exec ourselves with foreground flag
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	logFile, err := os.OpenFile(d.LogFile(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	attr := &os.ProcAttr{
		Dir: d.config.WorkspaceDir,
		Env: os.Environ(),
		Files: []*os.File{
			nil,     // stdin
			logFile, // stdout
			logFile, // stderr
		},
	}

	args := []string{exe, "-w", d.config.WorkspaceDir, "daemon", "start", "--foreground", "--interval", strconv.Itoa(int(d.config.Interval.Seconds()))}
	process, err := os.StartProcess(exe, args, attr)
	if err != nil {
		return err
	}

	// Don't wait for the child
	pid := process.Pid
	process.Release()
	fmt.Printf("Daemon started with PID %d\n", pid)
	return nil
}

func (d *Daemon) run() error {
	// Write PID file
	if err := os.WriteFile(d.PIDFile(), []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		return err
	}

	// Start HTTP server for web dashboard
	store := jsonl.NewStore(d.config.WorkspaceDir)
	d.httpServer = NewHTTPServer(store, d.config.WorkspaceDir, d.config.StaticFS)

	go func() {
		// Try preferred port first, fall back to any available port
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", DefaultHTTPPort))
		if err != nil {
			listener, err = net.Listen("tcp", ":0")
			if err != nil {
				fmt.Printf("HTTP server error: %v\n", err)
				return
			}
		}

		port := listener.Addr().(*net.TCPAddr).Port
		portFile := filepath.Join(d.config.BearingDir, "http.port")
		os.WriteFile(portFile, []byte(fmt.Sprintf("%d", port)), 0644)

		fmt.Printf("HTTP server listening on http://localhost:%d\n", port)
		if err := http.Serve(listener, d.httpServer.Handler()); err != nil {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	ticker := time.NewTicker(d.config.Interval)
	defer ticker.Stop()

	fmt.Printf("Daemon running (PID %d), checking every %v\n", os.Getpid(), d.config.Interval)

	// Initial check
	d.runHealthCheck()

	for {
		select {
		case <-ticker.C:
			d.runHealthCheck()
		case sig := <-sigChan:
			fmt.Printf("Received signal %v, stopping...\n", sig)
			os.Remove(d.PIDFile())
			return nil
		case <-d.stop:
			fmt.Println("Daemon stopping...")
			os.Remove(d.PIDFile())
			return nil
		}
	}
}

func (d *Daemon) runHealthCheck() {
	store := jsonl.NewStore(d.config.WorkspaceDir)
	entries := d.discoverWorktrees(store)

	var health []jsonl.HealthEntry
	for _, e := range entries {
		folderPath := filepath.Join(d.config.WorkspaceDir, e.Folder)
		h := jsonl.HealthEntry{
			Folder:    e.Folder,
			LastCheck: time.Now(),
		}

		repo := git.NewRepo(folderPath)
		h.Dirty, _ = repo.IsDirty()
		h.Unpushed, _ = repo.UnpushedCount(e.Branch)

		if !e.Base {
			ghClient := gh.NewClient(folderPath)
			if pr, _ := ghClient.GetPR(e.Branch); pr != nil {
				h.PRState = &pr.State
				h.PRTitle = &pr.Title
			}
		}

		health = append(health, h)
	}

	if err := store.WriteHealth(health); err != nil {
		fmt.Printf("Error writing health.jsonl: %v\n", err)
	}

	// Broadcast update to connected web clients
	if d.httpServer != nil {
		d.httpServer.Broadcast("health", map[string]interface{}{
			"timestamp":     time.Now(),
			"worktreeCount": len(health),
		})
	}
}

// discoverWorktrees finds all worktrees by scanning projects from projects.jsonl
// and running `git worktree list` for each. Merges with local.jsonl for local-only entries.
func (d *Daemon) discoverWorktrees(store *jsonl.Store) []jsonl.LocalEntry {
	discovered := make(map[string]jsonl.LocalEntry) // keyed by folder name

	// Read projects and discover worktrees via git
	projects, err := store.ReadProjects()
	if err != nil {
		fmt.Printf("Error reading projects.jsonl: %v\n", err)
	}

	for _, proj := range projects {
		basePath := filepath.Join(d.config.WorkspaceDir, proj.Path)
		repo := git.NewRepo(basePath)
		worktrees, err := repo.WorktreeList()
		if err != nil {
			fmt.Printf("Error listing worktrees for %s: %v\n", proj.Name, err)
			continue
		}

		for _, wt := range worktrees {
			if wt.Bare {
				continue
			}
			// Get folder name relative to workspace
			folder := filepath.Base(wt.Path)
			isBase := folder == proj.Path
			discovered[folder] = jsonl.LocalEntry{
				Folder: folder,
				Repo:   proj.Name,
				Branch: wt.Branch,
				Base:   isBase,
			}
		}
	}

	// Merge with local.jsonl for any local-only entries not discovered via git
	localEntries, err := store.ReadLocal()
	if err != nil {
		fmt.Printf("Error reading local.jsonl: %v\n", err)
	}
	for _, le := range localEntries {
		if _, exists := discovered[le.Folder]; !exists {
			discovered[le.Folder] = le
		}
	}

	// Convert map to slice
	result := make([]jsonl.LocalEntry, 0, len(discovered))
	for _, entry := range discovered {
		result = append(result, entry)
	}
	return result
}

// Stop sends a stop signal to a running daemon
func (d *Daemon) Stop() error {
	running, pid := d.IsRunning()
	if !running {
		return fmt.Errorf("daemon is not running")
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return err
	}

	fmt.Printf("Sent stop signal to daemon (PID %d)\n", pid)
	return nil
}
