---
title: Commands
description: Reference for all Bearing scripts
---

# Commands

All commands are run from your workspace root (e.g., `~/Projects/`).

| Command | Description |
|---------|-------------|
| `worktree-new` | Create a worktree for a branch |
| `worktree-cleanup` | Remove a worktree after merge |
| `worktree-sync` | Rebuild manifest from git state |
| `worktree-list` | Display manifest as ASCII table |
| `worktree-register` | Register existing folder as base |
| `worktree-check` | Validate invariants |

## Common Workflows

### Starting a new task

```bash
./bearing/scripts/worktree-new myapp feature-auth
cd myapp-feature-auth
# ... work on the feature ...
```

### Finishing a task

```bash
# After merging the PR
./bearing/scripts/worktree-cleanup myapp feature-auth
```

### Checking workspace health

```bash
./bearing/scripts/worktree-check
./bearing/scripts/worktree-list
```
