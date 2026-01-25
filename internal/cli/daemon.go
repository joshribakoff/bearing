package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joshribakoff/bearing/internal/daemon"
	"github.com/spf13/cobra"
)

var (
	daemonInterval   int
	daemonForeground bool
	daemonStatusJSON bool
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage the background health daemon",
}

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the health monitoring daemon",
	RunE:  runDaemonStart,
}

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the health monitoring daemon",
	RunE:  runDaemonStop,
}

var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check daemon status",
	RunE:  runDaemonStatus,
}

func init() {
	daemonStartCmd.Flags().IntVar(&daemonInterval, "interval", 300, "check interval in seconds")
	daemonStartCmd.Flags().BoolVar(&daemonForeground, "foreground", false, "run in foreground")
	daemonStatusCmd.Flags().BoolVar(&daemonStatusJSON, "json", false, "output as JSON")

	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
	rootCmd.AddCommand(daemonCmd)
}

func newDaemon() *daemon.Daemon {
	// Find web directory - try next to binary first, then relative to bearing dir
	exe, _ := os.Executable()
	webDir := filepath.Join(filepath.Dir(exe), "web")

	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		webDir = filepath.Join(BearingDir(), "..", "web")
	}

	config := daemon.Config{
		WorkspaceDir: WorkspaceDir(),
		BearingDir:   BearingDir(),
		Interval:     time.Duration(daemonInterval) * time.Second,
	}

	if info, err := os.Stat(webDir); err == nil && info.IsDir() {
		config.StaticFS = os.DirFS(webDir)
	}

	return daemon.New(config)
}

func runDaemonStart(cmd *cobra.Command, args []string) error {
	d := newDaemon()
	return d.Start(daemonForeground)
}

func runDaemonStop(cmd *cobra.Command, args []string) error {
	d := newDaemon()
	return d.Stop()
}

type daemonStatusOutput struct {
	Running bool `json:"running"`
	PID     int  `json:"pid,omitempty"`
}

func runDaemonStatus(cmd *cobra.Command, args []string) error {
	d := newDaemon()
	running, pid := d.IsRunning()

	if daemonStatusJSON {
		out := daemonStatusOutput{Running: running}
		if running {
			out.PID = pid
		}
		return json.NewEncoder(os.Stdout).Encode(out)
	}

	if running {
		fmt.Printf("running (PID %d)\n", pid)
	} else {
		fmt.Println("not running")
	}
	return nil
}
