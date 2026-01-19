---
title: Base Folders
description: Primary repository clones that stay on main
---

# Base Folders

Base folders are your primary clones of each repository. They serve as the source for creating worktrees.

## Rules

1. **Stay on main** - Base folders should always be on the default branch
2. **Never switch branches** - Use worktrees for feature work
3. **One per repo** - Each repository has exactly one base folder

## Why?

Base folders provide:

- **Stable reference point** - Always know where main is
- **Worktree source** - Git creates worktrees relative to the main clone
- **Safe harbor** - If a worktree gets corrupted, the base is untouched

## Identifying Base Folders

In `worktree-list` output, base folders have `BASE=yes`:

```
FOLDER          REPO          BRANCH    BASE
------          ----          ------    ----
myapp           myapp         main      yes    ← Base folder
myapp-feature   myapp         feature   no     ← Worktree
```

## What If I Switched Branches?

If a base folder accidentally switched branches:

```bash
# Check status
./bearing/scripts/worktree-check

# Fix it
git -C myapp checkout main
```

The `worktree-check` command will catch this violation.
