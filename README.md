# Bearing

Worktree-based workflow for parallel AI-assisted development.

## Why

When multiple AI agents work on the same codebase, they can step on each other if they switch branches in shared folders. Bearing enforces a worktree-per-task pattern that keeps agents isolated.

## Install

```bash
git clone https://github.com/bearing-dev/bearing ~/Projects/bearing
~/Projects/bearing/install.sh
```

The installer prompts for scope (project-level or global) and creates symlinks to Bearing's skills.

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
