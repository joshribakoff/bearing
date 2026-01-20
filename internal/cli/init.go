package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize bearing hooks in the workspace",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

// ClaudeSettings represents .claude/settings.json structure
type ClaudeSettings struct {
	Hooks map[string][]HookConfig `json:"hooks,omitempty"`
}

type HookConfig struct {
	Hooks []Hook `json:"hooks"`
}

type Hook struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

func runInit(cmd *cobra.Command, args []string) error {
	claudeDir := filepath.Join(WorkspaceDir(), ".claude")
	settingsPath := filepath.Join(claudeDir, "settings.json")

	// Ensure .claude directory exists
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return err
	}

	// Load existing settings or create new
	var settings ClaudeSettings
	if data, err := os.ReadFile(settingsPath); err == nil {
		json.Unmarshal(data, &settings)
	}

	if settings.Hooks == nil {
		settings.Hooks = make(map[string][]HookConfig)
	}

	// Define the bearing hook command
	bearingHook := Hook{
		Type:    "command",
		Command: "bearing worktree check --json",
	}

	// Check if hook already exists
	existingHooks := settings.Hooks["UserPromptSubmit"]
	hookExists := false
	for _, hc := range existingHooks {
		for _, h := range hc.Hooks {
			if h.Command == bearingHook.Command {
				hookExists = true
				break
			}
		}
	}

	if hookExists {
		fmt.Println("Hook already configured in .claude/settings.json")
	} else {
		// Add the hook
		settings.Hooks["UserPromptSubmit"] = append(existingHooks, HookConfig{
			Hooks: []Hook{bearingHook},
		})

		// Write back
		data, err := json.MarshalIndent(settings, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(settingsPath, data, 0644); err != nil {
			return err
		}
		fmt.Println("Added bearing hook to .claude/settings.json")
	}

	fmt.Println("\nBearing initialized. The worktree-check hook will run on each prompt.")
	return nil
}
