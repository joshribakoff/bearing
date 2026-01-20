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
}

func runPlanPush(cmd *cobra.Command, args []string) error {
	planFile := args[0]

	// Read and parse frontmatter
	fm, body, err := parsePlanFile(planFile)
	if err != nil {
		return err
	}

	if fm.Issue == "" {
		return fmt.Errorf("no issue number in frontmatter")
	}
	if fm.Repo == "" {
		return fmt.Errorf("no repo in frontmatter")
	}

	// Trim leading/trailing whitespace from body
	body = strings.TrimSpace(body)

	if planPushDryRun {
		fmt.Printf("Would update issue %s in %s:\n", fm.Issue, fm.Repo)
		fmt.Printf("Status: %s\n", fm.Status)
		fmt.Printf("Body:\n%s\n", body)
		return nil
	}

	// Push to GitHub using gh issue edit
	repoPath := filepath.Join(WorkspaceDir(), fm.Repo)
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

	scanner := bufio.NewScanner(f)
	// Bug 1 fix: Increase buffer size to handle long lines (up to 1MB)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "---" {
			if !inFrontmatter && !frontmatterDone {
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
					// Bug 3 fix: Validate issue number is numeric
					if !isNumeric(val) {
						return nil, "", fmt.Errorf("issue must be numeric, got: %q", val)
					}
					fm.Issue = val
				case "repo":
					fm.Repo = val
				case "status":
					fm.Status = val
				}
			}
		} else if frontmatterDone {
			body.WriteString(line)
			body.WriteString("\n")
		}
	}

	return fm, body.String(), scanner.Err()
}
