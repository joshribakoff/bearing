# Bearing

Worktree-based workflow for parallel AI-assisted development.

## Why

When multiple AI agents work on the same codebase, they can step on each other if they switch branches in shared folders. Bearing enforces a worktree-per-task pattern that keeps agents isolated.

## Install

Clone bearing into your workspace folder alongside your other projects:

```bash
git clone https://github.com/bearing-dev/bearing ~/Projects/bearing
~/Projects/bearing/install.sh
```

The installer prompts for scope (project-level or global) and creates symlinks to Bearing's skills.

## Workspace Layout

Bearing assumes a flat workspace folder containing all your projects and worktrees:

```
~/Projects/                    # Your workspace root
├── bearing/                   # Bearing itself (cloned here)
├── myapp/                     # Base folder (stays on main)
├── myapp-feature-auth/        # Worktree for auth feature
├── myapp-fix-bug-123/         # Worktree for bug fix
├── other-project/             # Another base folder
├── other-project-refactor/    # Its worktree
├── workflow.jsonl             # Workflow state (committable)
└── local.jsonl                # Local worktree state
```

This scales well—workspaces with 100+ worktrees work fine. The flat structure makes it easy to see everything at a glance and lets multiple AI agents work in parallel without conflicts.

## Architecture

Bearing leverages git's existing capabilities rather than duplicating them:

| Layer | Responsibility | Storage |
|-------|---------------|---------|
| **Git submodules** | Commit pointers, remotes, branch refs | `.gitmodules`, `.git/` |
| **Manifest** | Workflow metadata cache (purposes, status) | `workflow.jsonl`, `local.jsonl` |
| **Scripts** | Orchestration, guardrails | `bearing/scripts/` |

**Design principle:** Git is the source of truth. The manifest is a cache of computed state plus workflow metadata. After a fresh clone, run `worktree-sync` to rebuild the manifest from git state.

**Fresh clone workflow:**
```bash
git clone --recurse-submodules https://github.com/user/projects.git
cd projects
./bearing/scripts/worktree-sync  # Rebuild manifest from git
```

## Concepts

- **Base folders** (e.g., `fightingwithai.com/`) stay on `main`
- **Worktrees** (e.g., `fightingwithai.com-feature/`) are created for tasks
- **Workflow state** (`workflow.jsonl`) tracks branches, purposes, relationships (committable)
- **Local state** (`local.jsonl`) tracks worktree folders (not committed)
- **Config** (`.bearing.yaml`) defines workspace repos and settings

## Commands

Run from your Projects folder:

| Command | Description |
|---------|-------------|
| `./bearing/scripts/worktree-new <repo> <branch>` | Create worktree for branch |
| `./bearing/scripts/worktree-cleanup <repo> <branch>` | Remove worktree after merge |
| `./bearing/scripts/worktree-sync` | Rebuild manifest from git state |
| `./bearing/scripts/worktree-list` | Display manifest as ASCII table |
| `./bearing/scripts/worktree-register <folder>` | Register existing folder as base |
| `./bearing/scripts/worktree-check` | Validate invariants (base folders on main) |

### Options

```bash
# Create worktree with metadata
./bearing/scripts/worktree-new myrepo feature-x --based-on develop --purpose "Add login"
```

## State Files

Bearing uses two state files in the workspace root:

**workflow.jsonl** (committable - portable across machines):
```jsonl
{"repo":"myrepo","branch":"feature","basedOn":"main","purpose":"Add login","status":"in_progress","created":"2024-12-20T12:00:00Z"}
```

**local.jsonl** (not committed - local worktree paths):
```jsonl
{"folder":"myrepo","repo":"myrepo","branch":"main","base":true}
{"folder":"myrepo-feature","repo":"myrepo","branch":"feature","base":false}
```

Agents should interact via scripts, never edit these files directly.

## Config

`.bearing.yaml` defines the workspace:
```yaml
repos:
  - name: myrepo
    remote: https://github.com/org/myrepo.git
    defaultBranch: main

state:
  workflow: workflow.jsonl
  local: local.jsonl
```

## Testing

```bash
# Bash smoke tests (quick validation)
./test/smoke-test.sh

# Python integration tests (full coverage with mocking)
cd test && python -m pytest test_integration.py -v
```

The Python harness provides:
- Subprocess mocking for stdin/stdout (test like a user)
- Isolated temp directory per test
- Parallel test execution support (`pytest -n auto`)
- Structured assertions on JSONL state files

## Hooks

Bearing integrates with Claude Code's hook system to check invariants before each action.

Add to `.claude/settings.json`:

```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "\"$CLAUDE_PROJECT_DIR\"/bearing/scripts/worktree-check --json"
          }
        ]
      }
    ]
  }
}
```

**How it works:**
- Hook runs on every user prompt (before Claude responds)
- `--json` outputs Claude Code hook format with `continue: true`
- On violations, `systemMessage` tells Claude to ask the user whether to fix or proceed
- Claude sees the context and can offer to run `git -C <folder> checkout main`

**Flags:**
- `--json`: Output JSON for Claude Code hooks (always exits 0)
- `--quiet`: Suppress human-readable output on success (for manual runs)

## Slash Commands

After install, these slash commands are available:

| Command | Description |
|---------|-------------|
| `/worktree-status` | Check invariants and display worktree table |

## Plans (Lightweight Planning)

Bearing includes a lightweight planning system for tracking work tied to GitHub issues. Plans are Markdown files stored in a top-level `plans/` folder and committed to version control.

**Philosophy:**
- Plans live alongside code in version control
- Each plan links to a GitHub issue via frontmatter
- Plans are human-readable and agent-readable
- Committing plans creates an audit trail of decisions

**Workspace layout:**
```
~/Projects/
├── plans/
│   ├── myrepo/
│   │   ├── 123-add-auth.md      # Plan for issue #123
│   │   └── 145-fix-perf.md      # Plan for issue #145
│   └── other-repo/
│       └── 42-refactor.md
├── myrepo/
├── myrepo-add-auth/             # Worktree for the plan
└── workflow.jsonl
```

**Plan file format:**
```markdown
---
issue: 123
repo: myrepo
status: in_progress
---

# Add Authentication

## Context
User needs to log in before accessing dashboard.

## Approach
1. Add OAuth provider
2. Create session middleware
3. Protect dashboard routes

## Tasks
- [x] Research OAuth libraries
- [ ] Implement login flow
- [ ] Add tests
```

**Commands:**
```bash
bearing plan pull myrepo 123      # Create plan from GitHub issue
bearing plan push plans/myrepo/123.md  # Update issue from plan
bearing plan sync --project myrepo     # Sync all plans for project
```

**Benefits:**
- Plans persist across sessions (committed to git)
- Multiple agents can reference the same plan
- GitHub issue stays updated with progress
- Easy to review plan changes in PRs

## Daemon (Background Health Monitoring)

Bearing includes an optional background daemon that periodically checks worktree health and caches the results.

**What it does:**
- Runs health checks every N seconds (default: 300s / 5 min)
- Checks each worktree for: dirty state, unpushed commits, PR status
- Writes results to `health.jsonl` for fast `worktree status --cached` queries
- Respects rate limits (GitHub API calls are spaced out)

**Files:**
```
~/.bearing/
├── bearing.pid       # PID file (prevents duplicate daemons)
└── daemon.log        # Daemon output log
```

**Commands:**
```bash
bearing daemon start              # Start in background
bearing daemon start --foreground # Run in foreground (for debugging)
bearing daemon start --interval 60  # Check every 60 seconds
bearing daemon status             # Check if running
bearing daemon status --json      # {"running": true, "pid": 12345}
bearing daemon stop               # Send SIGTERM to stop
```

**PID file management:**
- On start: checks if PID file exists and process is alive (refuses to double-start)
- On run: writes current PID to file
- On stop: removed automatically via `defer`
- Stale PID files (process died) are detected and overwritten

**Use cases:**
- `worktree status --cached` returns instant results from `health.jsonl`
- `worktree status --refresh` forces immediate check
- Dashboard/notification integrations can watch `health.jsonl`

## Future Ideas

Documented for future consideration:

### Agent Wrapper Script
A wrapper script that runs pre-flight checks before launching any AI agent:
```bash
#!/bin/bash
./bearing/scripts/worktree-check || { echo "Fix violations first"; exit 1; }
exec "${BEARING_AGENT:-claude}" "$@"
```
Benefits: True blocking (refuses to start), portable across agents (Claude, Cursor, Aider), clear error display. Current hook approach works within Claude Code's system but cannot truly block session start.

### Workflow Automation
- **worktree-push**: Push branch and optionally create PR
- **worktree-finish**: Push + PR + mark complete in workflow.jsonl
- **worktree-checkout**: Recreate local worktree from workflow.jsonl entry (for switching machines)
- **Auto-PR templates**: Configure PR body template in .bearing.yaml

### Validation & Safety
- **Auto-fix with --force**: `worktree-check --fix` to checkout main on violating base folders
- **Git hooks in repos**: Pre-checkout hooks to block unsafe branch switches
- **Cross-repo coordination**: Track which agent owns which worktree

### Configuration
- **Local overrides**: `~/.bearing.yaml` for user-specific settings
- **Per-repo config**: `.bearing.yaml` in each repo for repo-specific behavior
- **Workflow config**: autoPush, autoPR, cleanupOnMerge settings

### Platform Support
- **Windows**: Path separator handling, PowerShell scripts
- **Shell compatibility**: Fish, zsh, bash differences
- **CI integration**: GitHub Actions for smoke tests
