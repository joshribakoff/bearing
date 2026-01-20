package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/joshribakoff/bearing/internal/jsonl"
	"github.com/spf13/cobra"
)

var (
	listJSON     bool
	listLocal    bool
	listWorkflow bool
)

var worktreeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List worktrees",
	RunE:  runWorktreeList,
}

func init() {
	worktreeListCmd.Flags().BoolVar(&listJSON, "json", false, "output as JSON")
	worktreeListCmd.Flags().BoolVar(&listLocal, "local", false, "show local worktrees only")
	worktreeListCmd.Flags().BoolVar(&listWorkflow, "workflow", false, "show workflow entries only")
	worktreeCmd.AddCommand(worktreeListCmd)
}

func runWorktreeList(cmd *cobra.Command, args []string) error {
	store := jsonl.NewStore(WorkspaceDir())

	if listWorkflow {
		return listWorkflowEntries(store)
	}
	return listLocalEntries(store)
}

func listLocalEntries(store *jsonl.Store) error {
	entries, err := store.ReadLocal()
	if err != nil {
		return err
	}

	if listJSON {
		return json.NewEncoder(os.Stdout).Encode(entries)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "FOLDER\tREPO\tBRANCH\tBASE")
	for _, e := range entries {
		base := ""
		if e.Base {
			base = "yes"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.Folder, e.Repo, e.Branch, base)
	}
	return w.Flush()
}

func listWorkflowEntries(store *jsonl.Store) error {
	entries, err := store.ReadWorkflow()
	if err != nil {
		return err
	}

	if listJSON {
		return json.NewEncoder(os.Stdout).Encode(entries)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "REPO\tBRANCH\tSTATUS\tPURPOSE")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.Repo, e.Branch, e.Status, e.Purpose)
	}
	return w.Flush()
}
