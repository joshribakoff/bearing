# TUI Development Skill

Development workflow for the Bearing TUI (Python/Textual).

## Setup

```bash
cd ~/Projects/bearing/tui
pip install -e ".[dev]"
```

## Commands

| Command | Description |
|---------|-------------|
| `make run` | Run the TUI |
| `make test` | Run all tests |
| `make screenshot` | Generate screenshot.svg |
| `make clean` | Remove build artifacts |

## Development Workflow

**After ANY code or CSS changes**:

```bash
cd ~/Projects/bearing/tui
pip install -e .          # Reinstall (CSS changes require this)
make test                  # Run tests
make screenshot            # Update screenshot
```

Textual doesn't hot-reload external CSS. Always reinstall after styling changes.

## Visual Changes Checklist

When modifying CSS/styling:

1. `pip install -e .` - Reinstall package
2. `make test` - Verify tests pass
3. `make screenshot` - Generate new screenshot
4. **Inspect the screenshot** - Open `assets/screenshot.svg` and verify visually
5. Run `pytest tests/test_visual_regression.py` - Generate comparison screenshots
6. Update README/docs if screenshots changed significantly

## Screenshot Generation

```bash
# Default screenshot
make screenshot

# Custom output path
bearing-tui --screenshot path/to/output.svg

# Visual regression tests (multiple states)
pytest tests/test_visual_regression.py -v
```

Screenshots use mock data and headless rendering via Textual's testing framework.

## Testing

Tests use `App.run_test()` for headless testing:

```python
@pytest.mark.asyncio
async def test_example(workspace):
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        await pilot.press("1")  # Focus panel 1
        assert isinstance(app.focused, WorktreeTable)
```

## Adding Features

1. Add binding in `app.py` BINDINGS list
2. Implement `action_*` method
3. Update help modal in HelpScreen
4. Update footer hint bar
5. Add test in `tests/test_app.py`
6. Run full workflow (test, screenshot, visual check)

## Styling

Styles are in `styles/app.tcss` using Textual CSS (Darcula theme).

Key selectors:
- `ProjectListItem.-highlight` - Project list selection
- `WorktreeTable > .datatable--cursor` - Table cursor
- `.panel-header` - Panel headers with numbers

**Important**: Use exact class names. `ListItem` won't match `ProjectListItem`.

## Color Consistency

Both panels should use same highlight colors:

| State | Color |
|-------|-------|
| Focused highlight | #2d5a8a |
| Unfocused highlight | #264f78 ($bg-selection) |

## Troubleshooting

**CSS changes not visible**: Reinstall with `pip install -e .`

**Highlight not showing**: Check CSS selector matches Python class name exactly

**Tests failing**: Run `make clean && pip install -e ".[dev]"` for fresh install
