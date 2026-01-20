---
title: Introduction
description: Worktree-based workflow for parallel AI-assisted development
---

# Bearing

:::caution[Experimental]
Bearing is experimental software being developed live on [YouTube](https://youtube.com/@joshribakoff). Expect breaking changes.
:::

Worktree-based workflow for parallel AI-assisted development.

## The Problem

Multiple AI agents working on the same codebase step on each other when they switch branches in shared folders.

## The Solution

Bearing enforces a **worktree-per-task** pattern. Each task gets its own isolated directory. No branch switching, no conflicts.

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
└── workflow.jsonl        # Tracks active work
```

Base folders stay on `main`. Worktrees are created for each task. This scales to 100+ worktrees.
