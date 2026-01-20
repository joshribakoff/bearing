---
title: Introduction
description: Worktree-based workflow for parallel AI-assisted development
---

# Bearing

:::caution[Experimental]
Bearing is experimental software. Expect breaking changes.
:::

Worktree-based workflow for parallel AI-assisted development.

## The Problem

Working across multiple repos and branches gets messy:
- AI agents step on each other when switching branches in shared folders
- You lose context switching between directories
- Easy to accidentally work in the wrong folder (like the parent mono-workspace)
- Hard to track what's in progress across dozens of repos

## The Solution

Bearing enforces a **worktree-per-task** pattern with a flat mono-workspace layout. Each task gets its own isolated directory. No branch switching, no conflicts, no getting lost.

## Install

```bash
git clone https://github.com/joshribakoff/bearing ~/Projects/bearing
cd ~/Projects/bearing
go build -o bearing ./cmd/bearing
# Add to PATH or move to /usr/local/bin
```

## Example Commands

```bash
bearing worktree new myapp feature-auth
bearing worktree list
bearing worktree cleanup myapp feature-auth
bearing worktree check
bearing worktree sync
```

## Workspace Structure

```
~/Projects/
├── bearing/              # Bearing itself
├── myapp/                # Base folder (stays on main)
├── myapp-feature-auth/   # Worktree for feature
├── myapp-fix-bug/        # Worktree for bug fix
├── other-repo/           # Another project
├── other-repo-refactor/  # Its worktree
└── workflow.jsonl        # Tracks active work across all repos
```

All projects live in one flat workspace. Base folders stay on `main`. Worktrees are created for each task. This scales to thousands of worktrees across hundreds of repos.
