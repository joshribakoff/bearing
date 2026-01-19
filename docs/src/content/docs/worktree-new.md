---
title: worktree-new
description: Create a worktree for a branch
---

# worktree-new

Create a new worktree for a branch.

## Usage

```bash
./bearing/scripts/worktree-new <repo> <branch> [options]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `repo` | Name of the repository (base folder name) |
| `branch` | Branch name to create/checkout |

## Options

| Option | Description |
|--------|-------------|
| `--based-on <branch>` | Branch to base the new branch on (default: main) |
| `--purpose "<text>"` | Description of what this worktree is for |

## Examples

```bash
# Basic usage
./bearing/scripts/worktree-new myapp feature-auth

# With metadata
./bearing/scripts/worktree-new myapp feature-auth \
  --based-on develop \
  --purpose "Add user authentication flow"
```

## What It Does

1. Creates a new git worktree at `{repo}-{branch}`
2. Checks out (or creates) the specified branch
3. Records the worktree in `local.jsonl`
4. Records the branch in `workflow.jsonl`

## Notes

- If the branch doesn't exist, it's created from `--based-on` (default: main)
- The worktree folder is created at the workspace root level
- Multiple worktrees for the same repo can exist simultaneously
