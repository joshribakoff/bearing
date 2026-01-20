package cli

import "github.com/spf13/cobra"

var worktreeCmd = &cobra.Command{
	Use:   "worktree",
	Short: "Manage git worktrees",
}

func init() {
	rootCmd.AddCommand(worktreeCmd)
}
