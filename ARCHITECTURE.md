# Bearing Architecture Review

This document provides a comprehensive architecture review of the Bearing codebase, identifying code smells, duplication, and technical debt accumulated during vibe-coding.

## System Overview

Bearing manages git worktrees with metadata tracking, consisting of:

```
                     +-----------------------+
                     |      User Interface   |
                     +-----------+-----------+
                                 |
          +----------------------+----------------------+
          |                      |                      |
          v                      v                      v
    +-----------+         +-------------+         +-----------+
    |  Go CLI   |         | Python TUI  |         |  Daemon   |
    | (bearing) |         | (textual)   |         | (Go bg)   |
    +-----------+         +-------------+         +-----------+
          |                      |                      |
          +----------------------+----------------------+
                                 |
                                 v
                     +-----------------------+
                     |    JSONL State Files  |
                     | local.jsonl           |
                     | workflow.jsonl        |
                     | health.jsonl          |
                     | projects.jsonl        |
                     +-----------------------+
                                 |
          +----------------------+----------------------+
          |                      |                      |
          v                      v                      v
    +-----------+         +-----------+         +-----------+
    | git CLI   |         |  gh CLI   |         | claude CLI|
    +-----------+         +-----------+         +-----------+
```

### Components

1. **Go CLI** (`cmd/bearing/`, `internal/`): Core worktree management commands
2. **Python TUI** (`tui/`): Textual-based terminal UI for browsing state
3. **Daemon** (`internal/daemon/`): Background health monitoring process
4. **Bash Scripts** (`scripts/`): Legacy implementation (pre-Go rewrite)

## Directory Structure Analysis

```
bearing/
├── cmd/bearing/main.go      # Entry point (clean, minimal)
├── internal/
│   ├── cli/                 # 18 files - Command handlers
│   │   ├── root.go          # Global state, WorkspaceDir()
│   │   ├── worktree*.go     # 9 files for worktree subcommands
│   │   ├── plan*.go         # 4 files for plan sync
│   │   ├── daemon.go        # Daemon management
│   │   └── projects.go      # Project lookup helper
│   ├── daemon/              # 3 files
│   │   ├── daemon.go        # Lifecycle, health check loop
│   │   ├── health.go        # Health assessment helpers
│   │   └── watcher.go       # File watcher (unused?)
│   ├── git/                 # Git CLI wrapper
│   │   └── repo.go          # WorktreeAdd, IsDirty, etc
│   ├── gh/                  # GitHub CLI wrapper
│   │   └── client.go        # PR/Issue operations
│   ├── jsonl/               # JSONL storage
│   │   ├── store.go         # Read/Write operations
│   │   ├── types.go         # Entry structs
│   │   └── lock.go          # File locking
│   └── ai/                  # Claude CLI wrapper
│       └── client.go        # AI summarization
├── tui/                     # Python Textual TUI
│   └── bearing_tui/
│       ├── app.py           # Main application (727 lines!)
│       ├── state.py         # JSONL reader (duplicate of Go)
│       └── widgets/         # UI components
├── scripts/                 # Bash scripts (legacy)
│   ├── worktree-new         # 84 lines
│   ├── worktree-list        # 68 lines
│   ├── worktree-cleanup     # 49 lines
│   ├── worktree-sync        # 74 lines
│   ├── worktree-status      # 264 lines (largest!)
│   └── ...
└── docs/                    # Astro documentation site
```

## Code Smells and Issues

### 1. CRITICAL: Duplicate Implementations (Bash AND Go)

**Every worktree command exists in both bash AND Go:**

| Functionality | Bash Script | Go Implementation |
|---------------|-------------|-------------------|
| worktree new | `scripts/worktree-new` (84 lines) | `internal/cli/worktree_new.go` |
| worktree list | `scripts/worktree-list` (68 lines) | `internal/cli/worktree_list.go` |
| worktree cleanup | `scripts/worktree-cleanup` (49 lines) | `internal/cli/worktree_cleanup.go` |
| worktree sync | `scripts/worktree-sync` (74 lines) | `internal/cli/worktree_sync.go` |
| worktree status | `scripts/worktree-status` (264 lines) | `internal/cli/worktree_status.go` |

**Impact**: Maintenance burden, potential behavior drift, confusion about which to use.

**Recommendation**: Delete bash scripts; they were the original implementation before Go rewrite. Keep only for reference if needed.

### 2. CRITICAL: Duplicate Data Structures (Python AND Go)

Both Python and Go define the same data structures:

**Go types** (`internal/jsonl/types.go`):
```go
type LocalEntry struct {
    Folder string `json:"folder"`
    Repo   string `json:"repo"`
    Branch string `json:"branch"`
    Base   bool   `json:"base"`
}
```

**Python types** (`tui/bearing_tui/state.py`):
```python
@dataclass
class LocalEntry:
    folder: str
    repo: str
    branch: str
    base: bool
```

**AND AGAIN** in `tui/bearing_tui/widgets/details.py`:
```python
@dataclass
class LocalEntry:
    folder: str
    repo: str
    branch: str
    base: bool = False
```

**AND AGAIN** in `tui/bearing_tui/widgets/worktrees.py`:
```python
@dataclass
class WorktreeEntry:
    folder: str
    repo: str
    branch: str
    base: bool = False
```

**Impact**: 4+ definitions of the same concept. Changes must be made in multiple places.

**Recommendation**: Python TUI should import from a single location. Consider generating Python types from Go or sharing a JSON schema.

### 3. Weird Abstractions: Empty Widget Files

Three widget files exist but are empty (0 bytes):
- `tui/bearing_tui/widgets/project_list.py` (0 bytes)
- `tui/bearing_tui/widgets/worktree_table.py` (0 bytes)
- `tui/bearing_tui/widgets/details_panel.py` (0 bytes)

These are dead files. The actual implementations are in:
- `projects.py` (ProjectList)
- `worktrees.py` (WorktreeTable)
- `details.py` (DetailsPanel)

**Recommendation**: Delete the empty files.

### 4. Giant God File: `app.py` (727 lines)

`tui/bearing_tui/app.py` contains:
- BearingApp class
- HelpScreen modal
- PlansScreen modal
- Mock data generation (_create_mock_workspace)
- Main entry point
- Session management
- All keybinding handlers

**Issues**:
- Hard to navigate
- Multiple responsibilities
- Mock data mixed with production code
- Screenshot generation embedded in main app

**Recommendation**: Split into:
- `app.py` - Just BearingApp
- `screens/help.py` - HelpScreen
- `screens/plans.py` - PlansScreen
- `mock_data.py` - Move to tests
- `session.py` - Session persistence

### 5. Inconsistent JSONL Field Names

Go uses camelCase in JSON tags, but Python parses both:

```go
// Go types.go
BasedOn string `json:"basedOn,omitempty"`
PRState *string `json:"prState,omitempty"`
LastCheck time.Time `json:"lastCheck"`
```

```python
# Python state.py
based_on=e.get("basedOn") if e.get("basedOn") != "unknown" else None,
pr_state=e.get("prState"),
last_check=_parse_datetime(e.get("lastCheck")),
```

This works but is fragile. No schema validation.

### 6. Unused Code: daemon/watcher.go

`internal/daemon/watcher.go` implements file watching with fsnotify, but it's never actually called anywhere in the daemon lifecycle. The daemon only uses timer-based polling.

**Recommendation**: Either wire up file watching or delete the unused code.

### 7. Hardcoded GitHub Username

In `tui/bearing_tui/app.py`:
```python
url = f"https://github.com/joshribakoff/{plan.project}/issues/{plan.issue}"
```

This hardcodes the GitHub owner instead of using projects.jsonl.

### 8. Global State in cli/root.go

```go
var (
    workspaceDir string
    rootCmd      = &cobra.Command{...}
)
```

Package-level variables make testing harder and create hidden dependencies.

### 9. Duplicate YAML Frontmatter Parsing

`internal/cli/plan_push.go` has frontmatter parsing:
```go
func parsePlanFile(path string) (*planFrontmatter, string, error) {...}
```

`tui/bearing_tui/widgets/plans.py` has its own:
```python
def parse_plan_frontmatter(file_path: Path) -> dict: {...}
```

**Recommendation**: Either share via JSON intermediate or consolidate.

### 10. Inconsistent Error Handling

Go code silently ignores some errors:
```go
// internal/cli/worktree_sync.go
branch, err := repo.CurrentBranch()
if err != nil {
    continue  // Silently skips
}
```

```go
// internal/daemon/daemon.go
h.Dirty, _ = repo.IsDirty()  // Ignores error
h.Unpushed, _ = repo.UnpushedCount(e.Branch)  // Ignores error
```

### 11. Magic Strings Throughout

Status values aren't constants:
```go
Status: "active"    // worktree_new.go
Status: "in_progress"  // scripts/worktree-new
status = "cleaned"     // worktree_cleanup.go
status = "merged"      // worktree_cleanup.go
```

Bash uses "in_progress", Go uses "active" for the same concept!

### 12. SAILKIT_DIR Reference in Scripts

Scripts reference old project name:
```bash
SAILKIT_DIR="$(dirname "$SCRIPT_DIR")"
```

This was before the project was renamed to "bearing".

## Data Flow Issues

### Daemon Health Check Duplicates CLI Logic

`internal/daemon/daemon.go` runHealthCheck():
```go
h.Dirty, _ = repo.IsDirty()
h.Unpushed, _ = repo.UnpushedCount(e.Branch)
ghClient := gh.NewClient(folderPath)
if pr, _ := ghClient.GetPR(e.Branch); pr != nil {
    h.PRState = &pr.State
}
```

`internal/cli/worktree_status.go` runWorktreeStatus():
```go
s.Dirty, _ = repo.IsDirty()
s.Unpushed, _ = repo.UnpushedCount(e.Branch)
ghClient := gh.NewClient(folderPath)
if pr, _ := ghClient.GetPR(e.Branch); pr != nil {
    s.PRState = &pr.State
}
```

Identical code in two places.

**Recommendation**: Extract to `internal/daemon/health.go` and use from both.

## JSONL Files Analysis

| File | Purpose | Fields |
|------|---------|--------|
| `local.jsonl` | Worktree paths | folder, repo, branch, base |
| `workflow.jsonl` | Branch metadata | repo, branch, basedOn, purpose, status, created |
| `health.jsonl` | Cached health | folder, dirty, unpushed, prState, lastCheck |
| `projects.jsonl` | Project config | name, github_repo, path |

**Issues**:
- `projects.jsonl` is only used by plan sync, not worktree commands
- No schema validation
- Status field values inconsistent between bash/Go

## Cleanup Plan

### Priority 1: Remove Dead Code
1. Delete empty widget files: `project_list.py`, `worktree_table.py`, `details_panel.py`
2. Delete unused `internal/daemon/watcher.go`
3. Archive or delete bash scripts (keep Git history)

### Priority 2: Consolidate Duplicates
1. Extract health check logic to shared function (daemon + CLI)
2. Consolidate Python dataclasses to single location
3. Define status constants

### Priority 3: Refactoring
1. Split `app.py` into smaller modules
2. Move mock data to test fixtures
3. Fix hardcoded GitHub username
4. Replace SAILKIT_DIR references

### Priority 4: Schema/Types
1. Define JSONL schemas (or use JSON Schema)
2. Consider code generation for Python types
3. Add validation

## Recommendations Summary

| Issue | Severity | Effort | Recommendation |
|-------|----------|--------|----------------|
| Bash+Go duplication | Critical | Low | Delete bash scripts |
| Empty widget files | Low | Low | Delete 3 files |
| Python type duplication | High | Medium | Consolidate to one module |
| Giant app.py | Medium | Medium | Split into modules |
| Unused watcher.go | Low | Low | Delete or wire up |
| Hardcoded username | Medium | Low | Use projects.jsonl |
| Duplicate health check | Medium | Low | Extract to shared function |
| Inconsistent status values | Medium | Low | Define constants |
| SAILKIT_DIR reference | Low | Low | Rename to BEARING_DIR |
