package main

import (
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/bearing-dev/bearing/pkg/worktree"
)

const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
  <title>Sailkit Worktrees</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #1a1a2e; color: #eee; padding: 2rem; }
    h1 { margin-bottom: 1.5rem; color: #00d9ff; }
    table { width: 100%; border-collapse: collapse; background: #16213e; border-radius: 8px; overflow: hidden; }
    th, td { padding: 0.75rem 1rem; text-align: left; border-bottom: 1px solid #1a1a2e; }
    th { background: #0f3460; color: #00d9ff; font-weight: 600; }
    tr:hover { background: #1a1a4e; }
    .base { color: #00ff88; font-weight: 600; }
    .worktree { color: #ffaa00; }
    .branch { font-family: monospace; background: #0f3460; padding: 0.2rem 0.5rem; border-radius: 4px; }
    .empty { color: #666; font-style: italic; }
    .stats { margin-bottom: 1.5rem; display: flex; gap: 2rem; }
    .stat { background: #16213e; padding: 1rem 1.5rem; border-radius: 8px; }
    .stat-value { font-size: 2rem; color: #00d9ff; font-weight: bold; }
    .stat-label { color: #888; font-size: 0.875rem; }
  </style>
</head>
<body>
  <h1>Sailkit Worktrees</h1>
  <div class="stats">
    <div class="stat"><div class="stat-value">{{.BaseCount}}</div><div class="stat-label">Base Repos</div></div>
    <div class="stat"><div class="stat-value">{{.WorktreeCount}}</div><div class="stat-label">Worktrees</div></div>
  </div>
  <table>
    <thead>
      <tr><th>Folder</th><th>Repo</th><th>Branch</th><th>Type</th></tr>
    </thead>
    <tbody>
      {{range .Entries}}
      <tr>
        <td>{{.Folder}}</td>
        <td>{{.Repo}}</td>
        <td>{{if .Branch}}<span class="branch">{{.Branch}}</span>{{else}}<span class="empty">unknown</span>{{end}}</td>
        <td>{{if .Base}}<span class="base">BASE</span>{{else}}<span class="worktree">worktree</span>{{end}}</td>
      </tr>
      {{end}}
    </tbody>
  </table>
</body>
</html>`

type viewData struct {
	Entries       []worktree.Entry
	BaseCount     int
	WorktreeCount int
}

func main() {
	root := worktree.FindRoot()
	state := worktree.NewState(root)

	entries, err := state.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading state: %v\n", err)
		os.Exit(1)
	}

	var baseCount, wtCount int
	for _, e := range entries {
		if e.Base {
			baseCount++
		} else {
			wtCount++
		}
	}

	data := viewData{
		Entries:       entries,
		BaseCount:     baseCount,
		WorktreeCount: wtCount,
	}

	tmpl, err := template.New("view").Parse(htmlTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Template error: %v\n", err)
		os.Exit(1)
	}

	outPath := filepath.Join(root, ".sailkit-view.html")
	f, err := os.Create(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated: %s\n", outPath)
	openBrowser("file://" + outPath)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}
