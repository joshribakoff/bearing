package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/joshribakoff/bearing/internal/gh"
	"github.com/joshribakoff/bearing/internal/git"
	"github.com/joshribakoff/bearing/internal/jsonl"
	"github.com/spf13/cobra"
)

var (
	statusJSON    bool
	statusRefresh bool
	statusCached  bool
)

var worktreeStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show worktree status with health info",
	RunE:  runWorktreeStatus,
}

func init() {
	worktreeStatusCmd.Flags().BoolVar(&statusJSON, "json", false, "output as JSON")
	worktreeStatusCmd.Flags().BoolVar(&statusRefresh, "refresh", false, "force refresh health data")
	worktreeStatusCmd.Flags().BoolVar(&statusCached, "cached", false, "use cached health data only")
	worktreeCmd.AddCommand(worktreeStatusCmd)
}

type worktreeStatus struct {
	Folder    string    `json:"folder"`
	Repo      string    `json:"repo"`
	Branch    string    `json:"branch"`
	Base      bool      `json:"base"`
	Dirty     bool      `json:"dirty"`
	Unpushed  int       `json:"unpushed"`
	PRState   *string   `json:"prState,omitempty"`
	LastCheck time.Time `json:"lastCheck,omitempty"`
}

func runWorktreeStatus(cmd *cobra.Command, args []string) error {
	store := jsonl.NewStore(WorkspaceDir())
	entries, err := store.ReadLocal()
	if err != nil {
		return err
	}

	// Try to load cached health data
	healthMap := make(map[string]jsonl.HealthEntry)
	if !statusRefresh {
		health, _ := store.ReadHealth()
		for _, h := range health {
			healthMap[h.Folder] = h
		}
	}

	var statuses []worktreeStatus
	var updatedHealth []jsonl.HealthEntry

	for _, e := range entries {
		s := worktreeStatus{
			Folder: e.Folder,
			Repo:   e.Repo,
			Branch: e.Branch,
			Base:   e.Base,
		}

		// Use cached data if available and not refreshing
		if cached, ok := healthMap[e.Folder]; ok && statusCached {
			s.Dirty = cached.Dirty
			s.Unpushed = cached.Unpushed
			s.PRState = cached.PRState
			s.LastCheck = cached.LastCheck
		} else {
			// Fetch fresh data
			folderPath := filepath.Join(WorkspaceDir(), e.Folder)
			repo := git.NewRepo(folderPath)

			s.Dirty, _ = repo.IsDirty()
			s.Unpushed, _ = repo.UnpushedCount(e.Branch)

			if !e.Base {
				ghClient := gh.NewClient(folderPath)
				if pr, _ := ghClient.GetPR(e.Branch); pr != nil {
					s.PRState = &pr.State
				}
			}
			s.LastCheck = time.Now()

			// Update health cache
			updatedHealth = append(updatedHealth, jsonl.HealthEntry{
				Folder:    e.Folder,
				Dirty:     s.Dirty,
				Unpushed:  s.Unpushed,
				PRState:   s.PRState,
				LastCheck: s.LastCheck,
			})
		}

		statuses = append(statuses, s)
	}

	// Write updated health data
	if len(updatedHealth) > 0 {
		store.WriteHealth(updatedHealth)
	}

	if statusJSON {
		return json.NewEncoder(os.Stdout).Encode(statuses)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "FOLDER\tBRANCH\tDIRTY\tUNPUSHED\tPR")
	for _, s := range statuses {
		dirty := ""
		if s.Dirty {
			dirty = "yes"
		}
		pr := "-"
		if s.PRState != nil {
			pr = *s.PRState
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n", s.Folder, s.Branch, dirty, s.Unpushed, pr)
	}
	return w.Flush()
}
