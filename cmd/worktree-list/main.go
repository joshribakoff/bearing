package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/bearing-dev/bearing/pkg/worktree"
)

func main() {
	root := worktree.FindRoot()
	state := worktree.NewState(root)

	entries, err := state.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading state: %v\n", err)
		os.Exit(1)
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
	w.Flush()
}
