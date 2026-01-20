package cli

import "github.com/spf13/cobra"

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Manage plan files",
}

func init() {
	rootCmd.AddCommand(planCmd)
}
