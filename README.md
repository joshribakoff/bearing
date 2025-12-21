# Sailkit

Worktree-based workflow for parallel AI-assisted development.

## Why

When multiple AI agents work on the same codebase, they can step on each other if they switch branches in shared folders. Sailkit enforces a worktree-per-task pattern that keeps agents isolated.

## Install

```bash
git clone https://github.com/sailkit-dev/sailkit-dev ~/Projects/sailkit-dev
~/Projects/sailkit-dev/install.sh
```

The installer builds Go binaries and creates skill symlinks.

## Commands

```bash
worktree-new <repo> <branch>       # Create worktree
worktree-cleanup <repo> <branch>   # Remove worktree after merge
worktree-list                      # Display state tables
worktree-sync                      # Rebuild manifest from git
worktree-register <folder>         # Register existing folder as base
worktree-check                     # Validate invariants
worktree-check --json              # JSON output for hooks
```

## Concepts

- **Base folders** (e.g., `myrepo/`) stay on `main`
- **Worktrees** (e.g., `myrepo-feature/`) are created for tasks

## State Files

**workflow.jsonl** (committable):
```jsonl
{"repo":"myrepo","branch":"feature","status":"in_progress"}
```

**local.jsonl** (not committed):
```jsonl
{"folder":"myrepo","repo":"myrepo","branch":"main","base":true}
{"folder":"myrepo-feature","repo":"myrepo","branch":"feature","base":false}
```

## Hooks

Add to `.claude/settings.json`:

```json
{
  "hooks": {
    "UserPromptSubmit": [{
      "hooks": [{
        "type": "command",
        "command": "\"$CLAUDE_PROJECT_DIR\"/sailkit-dev/bin/worktree-check --json"
      }]
    }]
  }
}
```

## Testing

```bash
go test ./pkg/worktree/...
```

## Development

```bash
go build ./cmd/...      # Build all commands
go test ./...           # Run all tests
```
