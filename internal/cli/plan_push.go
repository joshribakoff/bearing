package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
)

var planPushDryRun bool

var planPushCmd = &cobra.Command{
	Use:   "push <file>",
	Short: "Push plan updates to GitHub issue",
	Args:  cobra.ExactArgs(1),
	RunE:  runPlanPush,
}

func init() {
	planPushCmd.Flags().BoolVar(&planPushDryRun, "dry-run", false, "show what would be pushed")
	planCmd.AddCommand(planPushCmd)
}

type planFrontmatter struct {
	Issue  string
	Repo   string
	Status string
	Title  string
}

func runPlanPush(cmd *cobra.Command, args []string) error {
	planFile := args[0]

	// Read and parse frontmatter
	fm, body, err := parsePlanFile(planFile)
	if err != nil {
		return err
	}

	// Auto-infer repo from path: plans/<project>/xxx.md
	if fm.Repo == "" {
		fm.Repo = inferRepoFromPath(planFile)
		if fm.Repo == "" {
			return fmt.Errorf("no repo in frontmatter and couldn't infer from path")
		}
		fmt.Printf("Auto-detected repo: %s\n", fm.Repo)
	}

	// Auto-infer title from first markdown heading if not in frontmatter
	if fm.Title == "" {
		fm.Title = extractTitleFromBody(body)
		if fm.Title == "" {
			return fmt.Errorf("no title in frontmatter and no markdown heading found")
		}
		fmt.Printf("Auto-detected title: %s\n", fm.Title)
	}

	// Trim leading/trailing whitespace from body
	body = strings.TrimSpace(body)

	repoPath := GetRepoPath(fm.Repo)

	if fm.Issue == "" {
		// Create new issue
		if planPushDryRun {
			fmt.Printf("Would create issue in %s:\n", fm.Repo)
			fmt.Printf("Title: %s\n", fm.Title)
			fmt.Printf("Body:\n%s\n", body)
			return nil
		}

		ghCmd := exec.Command("gh", "issue", "create",
			"--title", fm.Title,
			"--body", body,
			"--label", "plan")
		ghCmd.Dir = repoPath
		var stdout, stderr bytes.Buffer
		ghCmd.Stdout = &stdout
		ghCmd.Stderr = &stderr

		if err := ghCmd.Run(); err != nil {
			return fmt.Errorf("failed to create issue: %w\n%s", err, stderr.String())
		}

		// Parse issue number from output (URL format: https://github.com/owner/repo/issues/123)
		url := strings.TrimSpace(stdout.String())
		parts := strings.Split(url, "/")
		issueNum := parts[len(parts)-1]

		// Update frontmatter with issue number
		if err := updateFrontmatter(planFile, "issue", issueNum); err != nil {
			fmt.Printf("Warning: created issue #%s but failed to update frontmatter: %v\n", issueNum, err)
		}

		fmt.Printf("Created issue #%s in %s\n", issueNum, fm.Repo)
		fmt.Printf("URL: %s\n", url)
		return nil
	}

	// Update existing issue
	if planPushDryRun {
		fmt.Printf("Would update issue %s in %s:\n", fm.Issue, fm.Repo)
		fmt.Printf("Status: %s\n", fm.Status)
		fmt.Printf("Body:\n%s\n", body)
		return nil
	}

	ghCmd := exec.Command("gh", "issue", "edit", fm.Issue, "--body", body)
	ghCmd.Dir = repoPath
	var stderr bytes.Buffer
	ghCmd.Stderr = &stderr

	if err := ghCmd.Run(); err != nil {
		return fmt.Errorf("failed to update issue: %w\n%s", err, stderr.String())
	}

	fmt.Printf("Updated issue %s in %s\n", fm.Issue, fm.Repo)
	return nil
}

// inferRepoFromPath extracts repo name from plan path: plans/<project>/xxx.md
func inferRepoFromPath(planFile string) string {
	// Get absolute path
	absPath, err := filepath.Abs(planFile)
	if err != nil {
		return ""
	}

	// Look for "plans" directory in path
	parts := strings.Split(absPath, string(filepath.Separator))
	for i, part := range parts {
		if part == "plans" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// extractTitleFromBody extracts the first markdown heading from the body
func extractTitleFromBody(body string) string {
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return ""
}

// updateFrontmatter updates or adds a field in the frontmatter
func updateFrontmatter(planFile, key, value string) error {
	content, err := os.ReadFile(planFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var result []string
	inFrontmatter := false
	updated := false

	for i, line := range lines {
		if line == "---" {
			if !inFrontmatter && i == 0 {
				inFrontmatter = true
				result = append(result, line)
				continue
			} else if inFrontmatter {
				// End of frontmatter - add field if not updated
				if !updated {
					result = append(result, fmt.Sprintf("%s: %s", key, value))
				}
				inFrontmatter = false
			}
		}

		if inFrontmatter && strings.HasPrefix(line, key+":") {
			result = append(result, fmt.Sprintf("%s: %s", key, value))
			updated = true
			continue
		}

		result = append(result, line)
	}

	return os.WriteFile(planFile, []byte(strings.Join(result, "\n")), 0644)
}

// stripQuotes removes leading/trailing single or double quotes from a string.
func stripQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// containsControlChars checks if a string contains null bytes or other control characters.
func containsControlChars(s string) bool {
	for _, r := range s {
		if r == 0 || (unicode.IsControl(r) && r != '\t') {
			return true
		}
	}
	return false
}

// isNumeric checks if a string contains only digits.
var numericRegex = regexp.MustCompile(`^\d+$`)

func isNumeric(s string) bool {
	return numericRegex.MatchString(s)
}

func parsePlanFile(path string) (*planFrontmatter, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	fm := &planFrontmatter{}
	var body strings.Builder
	inFrontmatter := false
	frontmatterDone := false
	lineNum := 0

	scanner := bufio.NewScanner(f)
	// Bug 1 fix: Increase buffer size to handle long lines (up to 1MB)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		if line == "---" {
			if !inFrontmatter && !frontmatterDone && lineNum == 1 {
				// Only start frontmatter if --- is on line 1
				inFrontmatter = true
				continue
			} else if inFrontmatter {
				inFrontmatter = false
				frontmatterDone = true
				continue
			}
		}

		if inFrontmatter {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				// Bug 2 fix: Strip quotes from frontmatter values
				val = stripQuotes(val)
				// Bug 4 fix: Reject null bytes and control characters
				if containsControlChars(val) {
					return nil, "", fmt.Errorf("frontmatter field %q contains invalid control characters", key)
				}
				switch key {
				case "issue":
					// Bug 3 fix: Validate issue number is numeric (skip "null")
					if val != "null" && val != "" {
						if !isNumeric(val) {
							return nil, "", fmt.Errorf("issue must be numeric, got: %q", val)
						}
						fm.Issue = val
					}
				case "repo":
					fm.Repo = val
				case "status":
					fm.Status = val
				case "title":
					fm.Title = val
				}
			}
		} else {
			// No frontmatter or after frontmatter - treat as body
			body.WriteString(line)
			body.WriteString("\n")
		}
	}

	return fm, body.String(), scanner.Err()
}
