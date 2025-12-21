# Worktree Status

Check the status of all worktrees and validate invariants.

## What This Does

1. Syncs the manifest with current git state
2. Displays all worktrees in a table
3. Checks for invariant violations (base folders not on main)

## Run

```bash
worktree-check    # Check for violations
worktree-list     # View all worktrees
```

## Fixing Violations

If a base folder is on the wrong branch:

```bash
git -C <folder> checkout main
```

## Output

Shows tables with:
- FOLDER: Directory name
- REPO: Parent repository
- BRANCH: Current branch
- BASE: Whether this is a base folder
