package ai

import (
	"bytes"
	"os"
	"os/exec"
)

// IsEnabled returns true if AI features are enabled
func IsEnabled() bool {
	return os.Getenv("BEARING_AI_ENABLED") != "0"
}

// Client wraps Claude CLI operations
type Client struct{}

// NewClient creates a new AI client
func NewClient() *Client {
	return &Client{}
}

// Prompt sends a prompt to Claude and returns the response
func (c *Client) Prompt(prompt string) (string, error) {
	if !IsEnabled() {
		return "", nil
	}

	cmd := exec.Command("claude", "-p", prompt)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return stdout.String(), nil
}
