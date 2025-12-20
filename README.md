# Sailkit

Worktree-based workflow for parallel AI-assisted development.

## Why

When multiple AI agents work on the same codebase, they can step on each other if they switch branches in shared folders. Sailkit enforces a worktree-per-task pattern that keeps agents isolated.

## Install

```bash
git clone https://github.com/sailkit-dev/sailkit-dev ~/Projects/sailkit-dev
~/Projects/sailkit-dev/install.sh
```

The installer prompts for scope (project-level or global) and creates symlinks to Sailkit's skills.

## Concepts

- **Base folders** (`fightingwithai.com/`, `bearing-dev/`) stay on `main`
- **Worktrees** (`fightingwithai.com-feature/`) are created for tasks
- **Manifest** (`WORKTREES.md`) tracks active worktrees across repos

## Commands

| Command | Description |
|---------|-------------|
| `worktree-new <repo> <branch>` | Create worktree for branch |
| `worktree-cleanup <repo> <branch>` | Remove worktree after merge |
| `worktree-sync` | Update manifest from actual state |

## Testing

```bash
./test/smoke-test.sh
```

Runs local smoke tests in a temp directory. Tests cover:
- worktree-new creates worktree and updates manifest
- worktree-cleanup removes worktree and updates manifest
- worktree-sync discovers existing worktrees
- install.sh creates correct symlinks
