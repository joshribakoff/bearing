# Worktree Workflow

Use git worktrees to isolate tasks. Never switch branches in base folders.

## Rules

1. **Base folders stay on main** - `fightingwithai.com/`, `bearing-dev/`, etc. must always be on `main`
2. **Create worktrees for tasks** - Use `worktree-new` or manual `git worktree add`
3. **Update the manifest** - Add entries to `WORKTREES.md` when creating worktrees
4. **Cleanup after merge** - Use `worktree-cleanup` to remove worktrees and manifest entries

## Creating a worktree

```bash
# From the Projects folder
worktree-new fightingwithai.com feature-branch

# Or manually
git -C fightingwithai.com worktree add ../fightingwithai.com-feature-branch feature-branch
```

## Unsafe operations (NEVER do these)

- `git checkout <branch>` in a base folder
- `git switch <branch>` in a base folder
- Renaming or moving worktree directories
- Working in another agent's worktree

## Cross-repo tasks

When a task spans repos (e.g., library + consuming site), create worktrees in both:

```bash
worktree-new bearing-dev feature-branch
worktree-new fightingwithai.com feature-branch
```

## Validation

If you accidentally switch a base folder off main:

```bash
git -C fightingwithai.com checkout main
```
