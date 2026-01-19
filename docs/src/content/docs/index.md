---
title: Introduction
description: Worktree-based workflow for parallel AI-assisted development
---

# Bearing

Worktree-based workflow for parallel AI-assisted development.

## The Problem

Multiple AI agents working on the same codebase step on each other when they switch branches in shared folders.

## The Solution

Bearing enforces a **worktree-per-task** pattern. Each task gets its own isolated directory. No branch switching, no conflicts.

## Install

```bash
git clone https://github.com/joshribakoff/bearing ~/Projects/bearing
~/Projects/bearing/install.sh
```

The installer adds skills and hooks to Claude Code. After that, just ask:

- "Create a worktree for the auth feature"
- "What worktrees do I have?"
- "Clean up the feature-auth worktree"

Claude handles the rest.

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
