---
title: worktree-register
description: Register existing folder as base
---

# worktree-register

Register an existing git repository folder as a base folder.

## Usage

```bash
bearing worktree register <folder>
```

## Arguments

| Argument | Description |
|----------|-------------|
| `folder` | Path to an existing git repository folder |

## Example

```bash
# Register a newly cloned repo
git clone https://github.com/org/new-project.git
bearing worktree register new-project
```

## What It Does

1. Verifies the folder is a git repository
2. Adds it to `local.jsonl` as a base folder
3. Records its current branch and remote

## When to Use

- After manually cloning a new repository
- When adding an existing project to a Bearing-managed workspace
- When migrating to Bearing from another workflow

## Notes

- The folder must already exist and be a git repository
- Typically used for base folders, not worktrees
- Worktrees created with `bearing worktree new` are registered automatically
