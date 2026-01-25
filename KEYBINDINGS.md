# Bearing Keybindings

Standard keybindings for TUI and Web dashboard.

## Navigation

| Key | Action | TUI | Web |
|-----|--------|-----|-----|
| j / Down | Move down | Yes | Yes |
| k / Up | Move up | Yes | Yes |
| h / Left | Focus left panel / projects | Yes | Yes |
| l / Right | Focus right panel / worktrees | Yes | Yes |
| Tab | Next panel | Yes | Yes (cycles views) |
| Shift+Tab | Previous panel | Yes | No |
| Enter | Select item | Yes | Yes |
| 0 | Focus projects panel | Yes | No |
| 1 | Focus worktrees panel | Yes | Switch to worktrees view |
| 2 | Focus details panel | Yes | Switch to issues view |
| 3 | - | No | Switch to PRs view |

## Actions

| Key | Action | TUI | Web |
|-----|--------|-----|-----|
| r | Refresh data | Yes | Yes |
| R | Force refresh (daemon) | Yes | No |
| o | Open PR in browser | Yes | Yes |
| p | View plans modal | Yes | Yes |
| n | New worktree | Yes (placeholder) | No |
| c | Cleanup worktree | Yes (placeholder) | No |
| d | Daemon health check | Yes | No |
| x | Toggle closed PRs | Yes | No |
| q | Quit | Yes | No (browser) |
| Ctrl+C | Quit | Yes | No (browser) |

## Modals

| Key | Action | TUI | Web |
|-----|--------|-----|-----|
| ? | Open/close help | Yes | Yes (open only) |
| Escape | Close modal | Yes | Yes |
| p | Close plans modal (when open) | Yes | Yes |

### Plans Modal Navigation

| Key | Action | TUI | Web |
|-----|--------|-----|-----|
| j / Down | Move down | Yes | Yes |
| k / Up | Move up | Yes | Yes |
| o | Open issue in browser | Yes | Yes |
| Escape | Close | Yes | Yes |
| p | Close | Yes | Yes |

## Discrepancies

### TUI-only features
- **Number keys (0/1/2)**: TUI uses these for direct panel focus; Web uses 1/2/3 for view switching
- **Shift+Tab**: TUI supports reverse panel cycling; Web does not
- **R (uppercase)**: Force refresh via daemon
- **n/c**: Worktree creation and cleanup (placeholders)
- **d**: Daemon health check
- **x**: Toggle visibility of closed/merged PRs
- **q/Ctrl+C**: App quit (not applicable in browser)

### Web-only features
- **Tab**: Cycles through views (worktrees/issues/prs) rather than panels
- **View switching (1/2/3)**: Switches between worktrees, issues, and PRs views

### Behavioral differences
- **h/l**: In TUI, these focus specific panels. In Web, `h` always goes to project list, `l` moves from project list to worktree table only.
- **Help modal close**: TUI supports `?` to toggle help; Web only supports `Escape` or `?` to close.

## Recommended Alignment

To achieve consistency, consider:

1. **Web should add Shift+Tab** for reverse panel navigation
2. **TUI should add view switching** when issues/PRs views are implemented
3. **Standardize number keys**: Either both use 0/1/2 for panels or both use 1/2/3 for views
4. **Web should add x key** for toggle closed PRs when filtering is implemented
