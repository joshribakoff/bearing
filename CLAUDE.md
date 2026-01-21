# Bearing

Worktree management CLI for parallel AI-assisted development.

## Dogfooding

**Always run the daemon**. If not running, start it:
```bash
go run ./cmd/bearing daemon start
```

Check status:
```bash
go run ./cmd/bearing daemon status
```

The TUI auto-starts daemon on launch if `bearing` is in PATH.

## Project Structure

```
bearing/
├── cmd/bearing/       # Go CLI entry point
├── internal/          # Go packages
├── scripts/           # Bash scripts (worktree-*, plan-*)
├── tui/               # Python TUI (Textual)
└── docs/              # Documentation site
```

## Development

### Go CLI
```bash
go build -o bearing ./cmd/bearing
go test ./...
make integration
```

### Python TUI
```bash
cd tui
pip install -e ".[dev]"
make test
make screenshot
```

**After TUI CSS/code changes**: `pip install -e .` (Textual doesn't hot-reload)

## Testing

### TUI Visual Regression
```bash
cd tui
pytest tests/test_visual_regression.py -v
```

Generates screenshots for multiple scenarios:
- Normal, empty, overflow, long names, single item
- Focused/unfocused states
- Different terminal sizes

### Adversarial Testing Protocol
After tests pass, verify they catch bugs:
1. Introduce intentional bug
2. Run tests - should fail
3. If passes, fix the test
4. Revert bug

## Skills

See `.claude/skills/` for:
- `tui-dev/` - TUI development workflow

## Plans

Active plans in `~/Projects/plans/bearing/`:
- 017 - TUI test foundation (current priority)
- 013 - Multi-view TUI
- 014 - Daemon plans indexing
- 015 - Unified build/CLI
- 016 - Haiku auto-descriptions

## Worktree Workflow

**CRITICAL:** Never work directly in base folders. Use worktrees.

```bash
./scripts/worktree-list              # View all
./scripts/worktree-new repo branch   # Create
./scripts/worktree-cleanup repo branch  # Remove
```
