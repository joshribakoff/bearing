---
title: worktree-list
description: Display manifest as ASCII table
---

# worktree-list

Display the current workspace state as an ASCII table.

## Usage

```bash
./bearing/scripts/worktree-list
```

## Output

```
=== Local Worktrees ===

FOLDER                    REPO          BRANCH              BASE
------                    ----          ------              ----
myapp                     myapp         main                yes
myapp-feature-auth        myapp         feature-auth        no
myapp-fix-bug-123         myapp         fix-bug-123         no
other-project             other-project main                yes

=== Workflow (Branches) ===

REPO          BRANCH           BASED_ON    STATUS       PURPOSE
----          ------           --------    ------       -------
myapp         feature-auth     main        in_progress  Add authentication
myapp         fix-bug-123      main        in_progress  Fix login redirect
```

## Sections

**Local Worktrees**: Physical folders on disk with their git state

**Workflow**: Branch metadata including purpose and status

## Notes

- `BASE=yes` indicates base folders that should stay on main
- Use this to get an overview before starting new work
- Combine with `worktree-check` to find issues
