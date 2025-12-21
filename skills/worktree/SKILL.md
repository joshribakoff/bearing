# Worktree Workflow

Use git worktrees to isolate tasks. Never switch branches in base folders.

## Rules

1. **Base folders stay on main** - Folders marked `base:true` must never switch branches
2. **Create worktrees for tasks** - Use `worktree-new`
3. **Check the manifest** - Run `worktree-list` to see all folders
4. **Cleanup after merge** - Use `worktree-cleanup`
5. **Commit parent after child** - After committing in a worktree, commit the parent project

## Commands

```bash
worktree-list                              # View all worktrees
worktree-new myrepo feature-branch         # Create worktree
worktree-cleanup myrepo feature-branch     # Remove after merge
worktree-sync                              # Rebuild manifest from git
```

## Unsafe operations (NEVER do these)

- `git checkout <branch>` in a base folder
- `git switch <branch>` in a base folder
- Editing JSONL files directly

## Cross-repo tasks

When a task spans repos, create worktrees in both:

```bash
worktree-new bearing-dev feature-branch
worktree-new fightingwithai.com feature-branch
```

## Recovery

If you accidentally switch a base folder off main:

```bash
git -C myrepo checkout main
```
