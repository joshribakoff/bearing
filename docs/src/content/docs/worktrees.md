---
title: Worktrees
description: Task-specific isolated directories
---

# Worktrees

Worktrees are git's built-in feature for having multiple working directories from a single repository. Bearing uses them to isolate work.

## How They Work

A worktree is a separate working directory with its own checked-out branch, but sharing the same git history:

```
myapp/                    # Base folder (main branch)
├── .git/                 # Main git directory
├── src/
└── package.json

myapp-feature-auth/       # Worktree (feature-auth branch)
├── .git                  # File linking to main .git
├── src/                  # Separate working copy
└── package.json
```

## Benefits

- **Full isolation** - Each worktree has its own files
- **No stashing** - Keep changes in progress, switch tasks freely
- **Parallel work** - Multiple agents work simultaneously
- **Shared history** - All worktrees share commits, branches, remotes

## Lifecycle

```bash
# Create worktree for a task
./bearing/scripts/worktree-new myapp feature-auth

# Work in it
cd myapp-feature-auth
git add . && git commit -m "Add auth"
git push -u origin feature-auth

# Clean up after merge
./bearing/scripts/worktree-cleanup myapp feature-auth
```

## Git Commands (Under the Hood)

Bearing wraps these git commands:

```bash
# Creating a worktree
git -C myapp worktree add ../myapp-feature-auth -b feature-auth

# Listing worktrees
git -C myapp worktree list

# Removing a worktree
git -C myapp worktree remove ../myapp-feature-auth
git -C myapp worktree prune
```

You can use git directly if needed—Bearing won't break.
