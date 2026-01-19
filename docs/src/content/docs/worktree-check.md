---
title: worktree-check
description: Validate workspace invariants
---

# worktree-check

Validate that the workspace follows Bearing's invariants.

## Usage

```bash
./bearing/scripts/worktree-check [options]
```

## Options

| Option | Description |
|--------|-------------|
| `--json` | Output JSON for Claude Code hooks (always exits 0) |
| `--quiet` | Suppress human-readable output on success |

## Checks Performed

1. **Base folders on main** - Base folders should be on their default branch
2. **Manifest consistency** - Entries match actual git state
3. **No orphaned worktrees** - All worktrees have manifest entries

## Example Output

```
✓ All base folders are on their default branch
✓ Manifest is consistent with git state
✓ No orphaned worktrees found
```

Or with violations:

```
✗ Base folder 'myapp' is on branch 'feature-x', expected 'main'
  Fix: git -C myapp checkout main
```

## Hook Integration

Use with Claude Code hooks to check invariants before each action:

```json
{
  "hooks": {
    "UserPromptSubmit": [{
      "hooks": [{
        "type": "command",
        "command": "\"$CLAUDE_PROJECT_DIR\"/bearing/scripts/worktree-check --json"
      }]
    }]
  }
}
```

See [Claude Code Hooks](/claude-code-hooks/) for details.
