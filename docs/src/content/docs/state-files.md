---
title: State Files
description: JSONL files tracking workspace state
---

# State Files

Bearing uses two JSONL files in the workspace root to track state.

## workflow.jsonl (Committable)

Tracks branch metadata that's portable across machines:

```jsonl
{"repo":"myapp","branch":"feature-auth","basedOn":"main","purpose":"Add authentication","status":"in_progress","created":"2024-12-20T12:00:00Z"}
{"repo":"myapp","branch":"fix-bug-123","basedOn":"main","purpose":"Fix login redirect","status":"in_progress","created":"2024-12-20T14:30:00Z"}
```

| Field | Description |
|-------|-------------|
| `repo` | Repository name |
| `branch` | Branch name |
| `basedOn` | Parent branch |
| `purpose` | Human-readable description |
| `status` | `in_progress`, `merged`, `abandoned` |
| `created` | ISO timestamp |

**Commit this file** - It's useful for sharing context across machines or with teammates.

## local.jsonl (Not Committed)

Tracks local worktree folder paths:

```jsonl
{"folder":"myapp","repo":"myapp","branch":"main","base":true}
{"folder":"myapp-feature-auth","repo":"myapp","branch":"feature-auth","base":false}
```

| Field | Description |
|-------|-------------|
| `folder` | Local folder name |
| `repo` | Repository name |
| `branch` | Current branch |
| `base` | Is this a base folder? |

**Don't commit this file** - It's machine-specific.

## Rebuilding State

If state files get corrupted or out of sync:

```bash
bearing worktree sync
```

This rebuilds `local.jsonl` from git state while preserving `workflow.jsonl` metadata.

## Direct Editing

Don't edit these files manually. Use the CLI instead—it handles consistency.
