---
title: worktree-sync
description: Rebuild manifest from git state
---

# worktree-sync

Rebuild the manifest files from git state.

## Usage

```bash
./bearing/scripts/worktree-sync
```

## When to Use

Run this after:

- Fresh clone of your workspace
- Manual git operations outside Bearing
- Recovering from corrupted state files

## What It Does

1. Scans all folders in the workspace
2. Detects git repositories and worktrees
3. Rebuilds `local.jsonl` from discovered state
4. Preserves workflow metadata from `workflow.jsonl`

## Example

```bash
# After cloning your workspace on a new machine
git clone --recurse-submodules https://github.com/user/workspace.git
cd workspace
./bearing/scripts/worktree-sync
```

## Notes

- Git is the source of truth—this command just rebuilds the cache
- Existing workflow metadata (purposes, status) is preserved
- New worktrees discovered are added to the manifest
