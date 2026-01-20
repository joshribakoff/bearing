# TUI Development Skill

Development workflow for the Bearing TUI.

## Setup

```bash
cd ~/Projects/bearing-tui/tui
make install-dev
```

## Commands

| Command | Description |
|---------|-------------|
| `make run` | Run the TUI |
| `make test` | Run tests |
| `make screenshot` | Generate screenshot.svg |
| `make clean` | Remove build artifacts |

## Screenshot Automation

Generate screenshots for documentation:

```bash
# Generate default screenshot
make screenshot

# Custom path
bearing-tui --screenshot path/to/output.svg
```

Screenshots are SVG format, rendered headlessly via Textual's testing framework.

## Testing

Tests use Textual's `App.run_test()` for headless testing:

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

## Styling

Styles are in `styles/app.tcss` using Textual CSS (Darcula theme).

Key selectors:
- `ProjectList ListItem.--highlight` - List selection
- `WorktreeTable > .datatable--cursor` - Table cursor
- `.panel-header` - Panel headers with numbers
