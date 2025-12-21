package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/sailkit-dev/sailkit-dev/pkg/worktree"
)

func main() {
	root := worktree.FindRoot()
	state := worktree.NewState(root)

	// Local worktrees
	local, err := state.ReadLocal()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading local.jsonl: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Local Worktrees ===")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "FOLDER\tREPO\tBRANCH\tBASE")
	for _, e := range local {
		base := ""
		if e.Base {
			base = "yes"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.Folder, e.Repo, e.Branch, base)
	}
	w.Flush()

	// Workflow branches
	workflow, err := state.ReadWorkflow()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading workflow.jsonl: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n=== Workflow Branches ===")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "REPO\tBRANCH\tSTATUS")
	for _, e := range workflow {
		fmt.Fprintf(w, "%s\t%s\t%s\n", e.Repo, e.Branch, e.Status)
	}
	w.Flush()
}
