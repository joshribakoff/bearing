package cli

import (
	"fmt"

	"github.com/joshribakoff/bearing/internal/ai"
	"github.com/joshribakoff/bearing/internal/jsonl"
	"github.com/spf13/cobra"
)

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "AI-powered commands (requires BEARING_AI_ENABLED=1)",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if !ai.IsEnabled() {
			return fmt.Errorf("AI features disabled. Set BEARING_AI_ENABLED=1 to enable")
		}
		return nil
	},
}

var aiSummarizeCmd = &cobra.Command{
	Use:   "summarize",
	Short: "Summarize workspace state",
	RunE:  runAISummarize,
}

var aiClassifyCmd = &cobra.Command{
	Use:   "classify-priority",
	Short: "Classify worktrees by priority",
	RunE:  runAIClassify,
}

var aiSuggestCmd = &cobra.Command{
	Use:   "suggest-fix",
	Short: "Suggest fixes for health issues",
	RunE:  runAISuggestFix,
}

func init() {
	aiCmd.AddCommand(aiSummarizeCmd)
	aiCmd.AddCommand(aiClassifyCmd)
	aiCmd.AddCommand(aiSuggestCmd)
	rootCmd.AddCommand(aiCmd)
}

func runAISummarize(cmd *cobra.Command, args []string) error {
	store := jsonl.NewStore(WorkspaceDir())

	locals, _ := store.ReadLocal()
	health, _ := store.ReadHealth()
	workflows, _ := store.ReadWorkflow()

	prompt := fmt.Sprintf(`Summarize this workspace state concisely:

Worktrees: %d local folders
Active workflows: %d
Health entries: %d

Focus on: what needs attention, blocked items, suggested next actions.
Keep response under 200 words.`, len(locals), len(workflows), len(health))

	client := ai.NewClient()
	response, err := client.Prompt(prompt)
	if err != nil {
		return err
	}

	fmt.Println(response)
	return nil
}

func runAIClassify(cmd *cobra.Command, args []string) error {
	store := jsonl.NewStore(WorkspaceDir())
	workflows, _ := store.ReadWorkflow()

	if len(workflows) == 0 {
		fmt.Println("No workflows to classify")
		return nil
	}

	prompt := "Classify these worktrees by priority (high/medium/low):\n\n"
	for _, w := range workflows {
		prompt += fmt.Sprintf("- %s/%s: %s (status: %s)\n", w.Repo, w.Branch, w.Purpose, w.Status)
	}
	prompt += "\nReturn a prioritized list with brief reasoning."

	client := ai.NewClient()
	response, err := client.Prompt(prompt)
	if err != nil {
		return err
	}

	fmt.Println(response)
	return nil
}

func runAISuggestFix(cmd *cobra.Command, args []string) error {
	store := jsonl.NewStore(WorkspaceDir())
	health, _ := store.ReadHealth()
	locals, _ := store.ReadLocal()

	localMap := make(map[string]jsonl.LocalEntry)
	for _, l := range locals {
		localMap[l.Folder] = l
	}

	// Find issues
	var issues []string
	for _, h := range health {
		local := localMap[h.Folder]
		if local.Base && h.Dirty {
			issues = append(issues, fmt.Sprintf("- %s: base folder has uncommitted changes", h.Folder))
		}
		if !local.Base && h.Unpushed > 0 {
			issues = append(issues, fmt.Sprintf("- %s: %d unpushed commits", h.Folder, h.Unpushed))
		}
	}

	if len(issues) == 0 {
		fmt.Println("No health issues found")
		return nil
	}

	prompt := "Suggest fixes for these workspace issues:\n\n"
	for _, issue := range issues {
		prompt += issue + "\n"
	}
	prompt += "\nProvide specific commands to resolve each issue."

	client := ai.NewClient()
	response, err := client.Prompt(prompt)
	if err != nil {
		return err
	}

	fmt.Println(response)
	return nil
}
