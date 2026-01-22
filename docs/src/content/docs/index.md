---
title: Introduction
description: Infrastructure for agentic workflows with Claude
---

# Bearing

:::caution[Experimental]
Bearing is experimental software. Read the [introductory blog post](https://www.joshribakoff.com/blog/deliberate-ai-use/). Expect breaking changes.
:::

**The best orchestrator for Claude is Claude.**

Bearing is not an orchestration framework. It's *infrastructure* that enables Claude to orchestrate itself.

## What Bearing Provides

- **Worktree management** — Isolated directories so parallel agents don't conflict
- **Plan visualization** — TUI to see all plans across repos
- **State sync** — JSONL files synced to GitHub issues for persistence
- **Query tools** — CLI commands agents can use to understand workspace state
- **Hooks** — Feed context to Claude Code agents automatically

## The Problem

Multiple AI agents working on the same codebase step on each other when they switch branches in shared folders.

## The Solution

Bearing enforces a **worktree-per-task** pattern. Each task gets its own isolated directory. Claude orchestrates the work; Bearing provides the infrastructure.

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
