"""Tests for critical keybindings that must always work."""
import pytest
from bearing_tui.app import BearingApp


@pytest.mark.asyncio
async def test_ctrl_c_quits(workspace):
    """Ctrl+C must ALWAYS quit the app - this is a hard requirement.

    Disabling Ctrl+C is a usability and accessibility issue.
    Users expect Ctrl+C to work in any terminal application.
    """
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        await pilot.press("ctrl+c")
        assert app._exit is True, "Ctrl+C must quit the app"


@pytest.mark.asyncio
async def test_q_quits(workspace):
    """q key should quit the app."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        await pilot.press("q")
        assert app._exit is True, "q must quit the app"


@pytest.mark.asyncio
async def test_escape_closes_modal(workspace):
    """Escape should close modal screens."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        await pilot.press("?")  # Open help
        await pilot.press("escape")
        # Should be back to main screen, not crashed
        assert len(app.screen_stack) == 1
