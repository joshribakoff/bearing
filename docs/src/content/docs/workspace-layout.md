---
title: Workspace Layout
description: Understanding Bearing's flat folder structure
---

# Workspace Layout

Bearing assumes a flat workspace folder containing all your projects and worktrees:

```
~/Projects/                    # Your workspace root
├── bearing/                   # Bearing itself (cloned here)
├── myapp/                     # Base folder (stays on main)
├── myapp-feature-auth/        # Worktree for auth feature
├── myapp-fix-bug-123/         # Worktree for bug fix
├── other-project/             # Another base folder
├── other-project-refactor/    # Its worktree
├── workflow.jsonl             # Workflow state (committable)
└── local.jsonl                # Local worktree state
```

## Why Flat?

The flat structure:

- **Scales well** - Workspaces with 100+ worktrees work fine
- **Easy to navigate** - See everything at a glance
- **Agent-friendly** - Simple paths, no nested confusion
- **Git-native** - Matches how git worktrees work

## Naming Convention

Worktrees follow the pattern `{repo}-{branch}`:

| Base Folder | Branch | Worktree Folder |
|-------------|--------|-----------------|
| `myapp` | `feature-auth` | `myapp-feature-auth` |
| `myapp` | `fix/bug-123` | `myapp-fix-bug-123` |
| `api-server` | `refactor` | `api-server-refactor` |

## Base Folders vs Worktrees

**Base folders** are your primary clones:
- Stay on `main` (or default branch)
- Never switch branches directly
- Used as the source for creating worktrees

**Worktrees** are task-specific:
- One per branch/task
- Deleted after merging
- Isolated from other work

## State Files

Two files track workspace state:

| File | Contents | Committed? |
|------|----------|------------|
| `workflow.jsonl` | Branches, purposes, relationships | Yes |
| `local.jsonl` | Local folder paths | No |

See [State Files](/docs/state-files/) for details.
