package gh

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"
)

// Client wraps GitHub CLI operations
type Client struct {
	repoPath string
}

// NewClient creates a Client for the given repo path
func NewClient(repoPath string) *Client {
	return &Client{repoPath: repoPath}
}

// PRInfo contains PR information
type PRInfo struct {
	State  string `json:"state"`
	Number int    `json:"number"`
	URL    string `json:"url"`
	Title  string `json:"title"`
}

// GetPR gets PR info for the given branch
func (c *Client) GetPR(branch string) (*PRInfo, error) {
	cmd := exec.Command("gh", "pr", "view", branch, "--json", "state,number,url,title")
	cmd.Dir = c.repoPath
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// No PR exists
		if strings.Contains(stderr.String(), "no pull requests found") {
			return nil, nil
		}
		return nil, err
	}

	var info PRInfo
	if err := json.Unmarshal(stdout.Bytes(), &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// Issue contains issue information
type Issue struct {
	Number int      `json:"number"`
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	State  string   `json:"state"`
	Labels []string `json:"labels"`
}

// GetIssue fetches an issue by number
func (c *Client) GetIssue(number int) (*Issue, error) {
	cmd := exec.Command("gh", "issue", "view", string(rune(number)), "--json", "number,title,body,state,labels")
	cmd.Dir = c.repoPath
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var issue Issue
	if err := json.Unmarshal(stdout.Bytes(), &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}
