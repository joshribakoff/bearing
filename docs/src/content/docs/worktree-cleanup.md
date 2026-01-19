---
title: worktree-cleanup
description: Remove a worktree after merge
---

# worktree-cleanup

Remove a worktree after its branch has been merged.

## Usage

```bash
./bearing/scripts/worktree-cleanup <repo> <branch>
```

## Arguments

| Argument | Description |
|----------|-------------|
| `repo` | Name of the repository |
| `branch` | Branch name of the worktree to remove |

## Example

```bash
./bearing/scripts/worktree-cleanup myapp feature-auth
```

## What It Does

1. Removes the worktree folder (`{repo}-{branch}`)
2. Runs `git worktree prune` to clean up git metadata
3. Removes the entry from `local.jsonl`
4. Optionally updates `workflow.jsonl` status to `merged`

## Notes

- Run this after your PR has been merged
- Does not delete the remote branch (do that via GitHub/GitLab)
- Safe to run even if the folder was manually deleted
