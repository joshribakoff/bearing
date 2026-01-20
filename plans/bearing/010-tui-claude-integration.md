# TUI/Claude Session Integration

Switch between Bearing TUI and Claude without nesting endlessly. Preserve selection state across switches.

## Status: Future

## Problem

Currently, if you want to use both Bearing TUI and Claude:
- Nesting Claude inside TUI (or vice versa) creates confusing layered terminals
- Switching between them loses context (what project/worktree was selected)
- No way to launch Claude scoped to a specific worktree

## Goals

1. **Seamless switching**: Exit TUI to Claude, return to TUI with state preserved
2. **Scoped launches**: Launch Claude in the context of a selected worktree
3. **Simple UX**: Single keybinding to switch, intuitive flow

## Proposed Approach

### Session State File

Save TUI state to a temp file on exit:

```json
{
  "selectedProject": "bearing",
  "selectedWorktree": "bearing-go-rewrite",
  "cursorPosition": {"panel": 1, "row": 3},
  "timestamp": "2024-01-19T22:30:00Z"
}
```

Location: `~/.bearing/tui-session.json`

### Launch Claude from TUI

New keybinding (`Enter` or `e` on worktree):
1. Save session state
2. Exit TUI
3. Print command to launch Claude in worktree dir
4. User runs command (or TUI execs it)

```bash
# TUI prints on exit:
cd ~/Projects/bearing-go-rewrite && claude

# Or user can paste:
bearing claude  # launches Claude in selected worktree
```

### Return to TUI

When launching `bearing-tui`:
1. Check for recent session file (< 1 hour old)
2. Restore selection state
3. Resume where you left off

### Alternative: Background TUI

Instead of exit/restart, keep TUI in background:
- `Ctrl+Z` to suspend TUI (shell job control)
- `fg` to resume
- Downside: Requires shell job control knowledge

## Keybindings

| Key | Action |
|-----|--------|
| `e` / `Enter` | Launch Claude in selected worktree |
| `Ctrl+C` / `q` | Quit TUI (saves state) |

## Implementation Steps

1. Add session state persistence to `BearingApp`
2. Save state on `action_quit()`
3. Restore state in `on_mount()` if session file exists
4. Add `action_launch_claude()` method
5. Print launch command on exit

## Open Questions

- Should TUI exec Claude directly or print command?
- How to handle Claude's own exit (back to shell, not TUI)?
- Should we clear session state after successful restore?
- Integration with tmux/screen for better session management?

## Related

- `bearing worktree cd <folder>` - prints cd command for shell integration
- Similar to lazygit's approach (external editor launches)
