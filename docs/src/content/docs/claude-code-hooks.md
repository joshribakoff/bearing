---
title: Claude Code Hooks
description: Automatic validation before each prompt
---

# Claude Code Hooks

The installer sets up hooks that validate your workspace before each prompt.

## What It Does

Before Claude responds to any message, Bearing checks:

- Base folders are on their default branch
- Manifest is consistent with git state
- No orphaned worktrees

## When Violations Occur

If something's wrong, Claude sees a message like:

> Base folder 'myapp' is on branch 'feature-x', expected 'main'.

Claude will ask if you want to fix it. No manual intervention needed—just approve the fix.
