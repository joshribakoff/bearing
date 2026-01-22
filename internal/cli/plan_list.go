package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var planListProject string

var planListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all plans",
	Long: `List all plans, optionally filtered by project.

Example:
  bearing plan list
  bearing plan list --project bearing`,
	RunE: runPlanList,
}

var planSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search plan titles and content",
	Long: `Search plan titles and content for a query string.

Example:
  bearing plan search "activity"`,
	Args: cobra.ExactArgs(1),
	RunE: runPlanSearch,
}

func init() {
	planListCmd.Flags().StringVarP(&planListProject, "project", "p", "", "filter by project name")
	planCmd.AddCommand(planListCmd)
	planCmd.AddCommand(planSearchCmd)
}

// planListInfo holds parsed plan metadata for list/search commands
type planListInfo struct {
	ID      string
	Repo    string
	Status  string
	Title   string
	Content string
}

func getPlansDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, "Projects", "plans"), nil
}

func parsePlanForList(path string) (*planListInfo, error) {
	fm, body, err := parsePlanFile(path)
	if err != nil {
		return nil, err
	}

	// Extract title from body if not in frontmatter
	title := fm.Title
	if title == "" {
		title = extractTitleFromBody(body)
	}

	return &planListInfo{
		ID:      "", // ID not in planFrontmatter, will be extracted from filename
		Repo:    fm.Repo,
		Status:  fm.Status,
		Title:   title,
		Content: body,
	}, nil
}

// extractIDFromFilename extracts the ID prefix from plan filename (e.g., "ppqiw" from "ppqiw-activity-feed.md")
func extractIDFromFilename(filename string) string {
	base := filepath.Base(filename)
	base = strings.TrimSuffix(base, ".md")
	if idx := strings.Index(base, "-"); idx > 0 {
		return base[:idx]
	}
	return ""
}

func loadPlans(projectFilter string) ([]*planListInfo, error) {
	plansDir, err := getPlansDir()
	if err != nil {
		return nil, err
	}

	var plans []*planListInfo
	var projects []string

	if projectFilter != "" {
		projects = []string{projectFilter}
	} else {
		// Read all project directories
		entries, err := os.ReadDir(plansDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read plans directory: %w", err)
		}
		for _, e := range entries {
			if e.IsDir() {
				projects = append(projects, e.Name())
			}
		}
	}

	for _, project := range projects {
		projectDir := filepath.Join(plansDir, project)
		files, err := os.ReadDir(projectDir)
		if err != nil {
			continue // Skip inaccessible directories
		}
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".md") {
				continue
			}
			filePath := filepath.Join(projectDir, f.Name())
			plan, err := parsePlanForList(filePath)
			if err != nil {
				continue // Skip unparseable files
			}
			// Extract ID from filename if not in frontmatter
			plan.ID = extractIDFromFilename(f.Name())
			// Use directory name as repo if not in frontmatter
			if plan.Repo == "" {
				plan.Repo = project
			}
			plans = append(plans, plan)
		}
	}

	return plans, nil
}

func runPlanList(cmd *cobra.Command, args []string) error {
	plans, err := loadPlans(planListProject)
	if err != nil {
		return err
	}

	if len(plans) == 0 {
		fmt.Println("No plans found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATUS\tTITLE")
	for _, p := range plans {
		id := p.ID
		if id == "" {
			id = "-"
		}
		status := p.Status
		if status == "" {
			status = "-"
		}
		title := p.Title
		if title == "" {
			title = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", id, status, title)
	}
	return w.Flush()
}

func runPlanSearch(cmd *cobra.Command, args []string) error {
	query := strings.ToLower(args[0])

	plans, err := loadPlans("")
	if err != nil {
		return err
	}

	var matches []*planListInfo
	for _, p := range plans {
		titleLower := strings.ToLower(p.Title)
		contentLower := strings.ToLower(p.Content)
		if strings.Contains(titleLower, query) || strings.Contains(contentLower, query) {
			matches = append(matches, p)
		}
	}

	if len(matches) == 0 {
		fmt.Println("No plans found matching query")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tPROJECT\tTITLE")
	for _, p := range matches {
		id := p.ID
		if id == "" {
			id = "-"
		}
		repo := p.Repo
		if repo == "" {
			repo = "-"
		}
		title := p.Title
		if title == "" {
			title = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", id, repo, title)
	}
	return w.Flush()
}
