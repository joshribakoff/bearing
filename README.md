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
| `./bearing/scripts/worktree-status` | Show health of all worktrees |
| `./bearing/scripts/worktree-recover <base-folder>` | Fix base folder on wrong branch |
| `./bearing/scripts/worktree-sync` | Rebuild manifest from git state |
| `./bearing/scripts/worktree-list` | Display manifest as ASCII table |
| `./bearing/scripts/worktree-register <folder>` | Register existing folder as base |
| `./bearing/scripts/worktree-check` | Validate invariants, show health warnings |
| `./bearing/scripts/plan-sync` | Sync plans with issue trackers |
| `./bearing/scripts/plan-push <file>` | Push plan to issue tracker |
| `./bearing/scripts/plan-pull <repo> <issue>` | Pull issue to local plan |

### Options

```bash
# Create worktree with metadata
./bearing/scripts/worktree-new myrepo feature-x --based-on develop --purpose "Add login"
```

## State Files

Bearing uses three state files in the workspace root:

**workflow.jsonl** (committable - portable across machines):
```jsonl
{"repo":"myrepo","branch":"feature","basedOn":"main","purpose":"Add login","status":"in_progress","created":"2024-12-20T12:00:00Z"}
```

**local.jsonl** (not committed - local worktree paths):
```jsonl
{"folder":"myrepo","repo":"myrepo","branch":"main","base":true}
{"folder":"myrepo-feature","repo":"myrepo","branch":"feature","base":false}
```

**health.jsonl** (not committed - cached health status):
```jsonl
{"folder":"myrepo-feature","dirty":true,"unpushed":2,"prState":"OPEN","lastCheck":"2026-01-19T10:00:00Z"}
```

Agents should interact via scripts, never edit these files directly.

## Health Monitoring

`worktree-status` shows the health of all worktrees:
- **Dirty**: Uncommitted changes
- **Unpushed**: Commits not pushed to remote
- **Stale**: PR merged but worktree still exists
- **Base violations**: Base folder not on main

`worktree-recover` fixes base folders that accidentally switched off main, preserving uncommitted work.

## Plan Sync

Sync local markdown plans with issue trackers (GitHub, Linear, Jira).

Plans live in `~/Projects/plans/<project>/` with frontmatter:
```yaml
---
title: Feature name
github_repo: user/repo
github_issue: 42
---
```

Uses adapter pattern - implement `adapters/github.sh`, `adapters/linear.sh`, etc.

## AI Features (Opt-in)

Bearing can use Claude CLI (haiku model) for classification and summarization. **Disabled by default.**

### Enable

```bash
# Option 1: Environment variable
export BEARING_AI_ENABLED=1

# Option 2: User config (~/.bearing)
echo "ai_enabled: true" >> ~/.bearing

# Option 3: Workspace config (.bearing.yaml)
# ai:
#   enabled: true
```

### Auto-Generated Purpose

When creating worktrees, Bearing can auto-generate a purpose description from the branch name:

```bash
worktree-new myrepo feature-add-auth
# Purpose auto-generated: "Add authentication flow"
```

Override with explicit `--purpose`:
```bash
worktree-new myrepo feature-add-auth --purpose "OAuth2 login"
```

### Available Commands

| Command | Description |
|---------|-------------|
| `bearing-ai summarize` | Summarize input text |
| `bearing-ai branch-purpose` | Generate purpose from branch name |
| `bearing-ai classify-priority` | Classify as P0/P1/P2/P3 |
| `bearing-ai suggest-fix` | Suggest commands to fix issues |

### Requirements

- Claude CLI (`claude`) installed and authenticated
- Opt-in enabled via config or env var
- Uses haiku model for cost efficiency (~$0.0001 per call)

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
