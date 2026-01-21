package cli

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/joshribakoff/bearing/internal/jsonl"
)

var projectsCache map[string]*jsonl.ProjectEntry

// LoadProjects loads and caches projects from projects.jsonl
func LoadProjects() (map[string]*jsonl.ProjectEntry, error) {
	if projectsCache != nil {
		return projectsCache, nil
	}

	projectsFile := filepath.Join(WorkspaceDir(), "projects.jsonl")
	f, err := os.Open(projectsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No projects.jsonl yet
		}
		return nil, err
	}
	defer f.Close()

	projectsCache = make(map[string]*jsonl.ProjectEntry)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry jsonl.ProjectEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}
		projectsCache[entry.Name] = &entry
	}
	return projectsCache, scanner.Err()
}

// LookupGitHubRepo returns the GitHub repo (owner/repo) for a project name
func LookupGitHubRepo(projectName string) string {
	projects, err := LoadProjects()
	if err != nil || projects == nil {
		return ""
	}
	if p, ok := projects[projectName]; ok {
		return p.GitHubRepo
	}
	return ""
}

// GetRepoPath returns the local path for running gh commands
func GetRepoPath(projectName string) string {
	projects, err := LoadProjects()
	if err != nil || projects == nil {
		return filepath.Join(WorkspaceDir(), projectName)
	}
	if p, ok := projects[projectName]; ok {
		return filepath.Join(WorkspaceDir(), p.Path)
	}
	return filepath.Join(WorkspaceDir(), projectName)
}
