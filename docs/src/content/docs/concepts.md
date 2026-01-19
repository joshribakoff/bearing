---
title: Concepts
description: Core concepts behind Bearing's design
---

# Concepts

Bearing is built on a few simple concepts that work together.

## Core Ideas

| Concept | Description |
|---------|-------------|
| [Base Folders](/base-folders/) | Primary clones that stay on main |
| [Worktrees](/worktrees/) | Task-specific isolated directories |
| [State Files](/state-files/) | JSONL files tracking workspace state |

## Design Principles

**Git is the source of truth.** Bearing's manifest files are a cache of computed state plus workflow metadata. You can always rebuild from git state.

**Isolation over coordination.** Instead of complex locking or coordination between agents, each agent gets its own isolated worktree.

**Simple tools over complex systems.** Bash scripts, JSONL files, standard git commands. No daemons, no databases, no magic.
