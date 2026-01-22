package daemon

import (
	"fmt"
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

// Config holds daemon configuration
type Config struct {
	WorkspaceDir string
	BearingDir   string
	Interval     time.Duration
}

// Daemon manages the health monitoring background process
type Daemon struct {
	config          Config
	stop            chan struct{}
	activityTracker *ActivityTracker
}

// New creates a new daemon instance
func New(config Config) *Daemon {
	store := jsonl.NewStore(config.WorkspaceDir)
	return &Daemon{
		config:          config,
		stop:            make(chan struct{}),
		activityTracker: NewActivityTracker(store),
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
	entries, err := store.ReadLocal()
	if err != nil {
		fmt.Printf("Error reading local.jsonl: %v\n", err)
		return
	}

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

		var ghClient *gh.Client
		if !e.Base {
			ghClient = gh.NewClient(folderPath)
			if pr, _ := ghClient.GetPR(e.Branch); pr != nil {
				h.PRState = &pr.State
				h.PRTitle = &pr.Title
			}
		}

		// Track activity events
		d.activityTracker.CheckForActivity(folderPath, e, ghClient, repo)

		health = append(health, h)
	}

	if err := store.WriteHealth(health); err != nil {
		fmt.Printf("Error writing health.jsonl: %v\n", err)
	}
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
