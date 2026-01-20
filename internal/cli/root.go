package cli

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	workspaceDir string
	rootCmd      = &cobra.Command{
		Use:   "bearing",
		Short: "Worktree management for parallel AI-assisted development",
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&workspaceDir, "workspace", "w", "", "workspace directory (default: current directory)")
}

// WorkspaceDir returns the configured workspace directory
func WorkspaceDir() string {
	if workspaceDir != "" {
		return workspaceDir
	}
	dir, _ := os.Getwd()
	return dir
}

// BearingDir returns the ~/.bearing directory for daemon files
func BearingDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".bearing")
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
