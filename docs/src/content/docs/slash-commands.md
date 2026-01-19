---
title: Slash Commands
description: Quick status checks in Claude Code
---

# Slash Commands

After installing Bearing, these slash commands are available in Claude Code.

## Available Commands

| Command | Description |
|---------|-------------|
| `/worktree-status` | Check invariants and display worktree table |

## Usage

In Claude Code, type:

```
/worktree-status
```

Claude will run the status check and show:

- Current worktree state
- Any invariant violations
- Suggested fixes

## How It Works

Slash commands are defined as "skills" in Bearing's `skills/` directory. The installer creates symlinks so Claude Code can find them.

## Adding Custom Commands

Skills are markdown files with embedded prompts. See `bearing/skills/` for examples.
