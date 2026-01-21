"""Tests for critical keybindings that must always work."""
import pytest
from bearing_tui.app import BearingApp, HelpScreen, PlansScreen


# =============================================================================
# CTRL+C TESTS - Must work from EVERY screen
# =============================================================================

@pytest.mark.asyncio
async def test_ctrl_c_quits_from_main(workspace):
    """Ctrl+C must ALWAYS quit the app from main screen.

    Disabling Ctrl+C is a usability and accessibility issue.
    Users expect Ctrl+C to work in any terminal application.
    """
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        await pilot.pause()
        await pilot.press("ctrl+c")
        assert app._exit is True, "Ctrl+C must quit from main screen"


@pytest.mark.asyncio
async def test_ctrl_c_quits_from_help_modal(workspace):
    """Ctrl+C must quit from help modal."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        await pilot.pause()
        await pilot.press("?")  # Open help
        await pilot.pause()
        assert isinstance(app.screen, HelpScreen)

        await pilot.press("ctrl+c")
        assert app._exit is True, "Ctrl+C must quit from help modal"


@pytest.mark.asyncio
async def test_ctrl_c_quits_from_plans_modal(workspace):
    """Ctrl+C must quit from plans modal."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        await pilot.pause()
        await pilot.press("p")  # Open plans
        await pilot.pause()
        assert isinstance(app.screen, PlansScreen)

        await pilot.press("ctrl+c")
        assert app._exit is True, "Ctrl+C must quit from plans modal"


@pytest.mark.asyncio
async def test_ctrl_c_quits_from_projects_panel(workspace):
    """Ctrl+C must quit when projects panel is focused."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        await pilot.pause()
        await pilot.press("0")  # Focus projects
        await pilot.pause()
        assert app.focused.id == "project-list"

        await pilot.press("ctrl+c")
        assert app._exit is True, "Ctrl+C must quit from projects panel"


@pytest.mark.asyncio
async def test_ctrl_c_quits_from_worktrees_panel(workspace):
    """Ctrl+C must quit when worktrees panel is focused."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        await pilot.pause()
        await pilot.press("1")  # Focus worktrees
        await pilot.pause()
        assert app.focused.id == "worktree-table"

        await pilot.press("ctrl+c")
        assert app._exit is True, "Ctrl+C must quit from worktrees panel"


@pytest.mark.asyncio
async def test_ctrl_c_quits_from_details_panel(workspace):
    """Ctrl+C must quit when details panel is focused."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        await pilot.pause()
        await pilot.press("2")  # Focus details
        await pilot.pause()
        assert app.focused.id == "details-panel"

        await pilot.press("ctrl+c")
        assert app._exit is True, "Ctrl+C must quit from details panel"


# =============================================================================
# Q KEY TESTS
# =============================================================================

@pytest.mark.asyncio
async def test_q_quits_from_main(workspace):
    """q key should quit the app from main screen."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        await pilot.pause()
        await pilot.press("q")
        assert app._exit is True, "q must quit the app"


@pytest.mark.asyncio
async def test_q_quits_from_help_modal(workspace):
    """q key should quit the app from help modal."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        await pilot.pause()
        await pilot.press("?")
        await pilot.pause()
        await pilot.press("q")
        assert app._exit is True, "q must quit from help modal"


@pytest.mark.asyncio
async def test_q_quits_from_plans_modal(workspace):
    """q key should quit the app from plans modal."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        await pilot.pause()
        await pilot.press("p")
        await pilot.pause()
        await pilot.press("q")
        assert app._exit is True, "q must quit from plans modal"


# =============================================================================
# ESCAPE KEY TESTS
# =============================================================================

@pytest.mark.asyncio
async def test_escape_closes_help_modal(workspace):
    """Escape should close help modal."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        await pilot.pause()
        await pilot.press("?")
        await pilot.pause()
        assert isinstance(app.screen, HelpScreen)

        await pilot.press("escape")
        await pilot.pause()
        assert len(app.screen_stack) == 1


@pytest.mark.asyncio
async def test_escape_closes_plans_modal(workspace):
    """Escape should close plans modal."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        await pilot.pause()
        await pilot.press("p")
        await pilot.pause()
        assert isinstance(app.screen, PlansScreen)

        await pilot.press("escape")
        await pilot.pause()
        assert len(app.screen_stack) == 1


@pytest.mark.asyncio
async def test_escape_on_main_does_nothing(workspace):
    """Escape on main screen should not crash or exit."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        await pilot.pause()

        await pilot.press("escape")
        await pilot.pause()

        # Should still be on main screen, not crashed, not exited
        assert len(app.screen_stack) == 1
        assert app._exit is not True


# =============================================================================
# QUESTION MARK KEY TESTS
# =============================================================================

@pytest.mark.asyncio
async def test_question_mark_opens_help(workspace):
    """? key should open help modal."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        await pilot.pause()

        await pilot.press("?")
        await pilot.pause()

        assert isinstance(app.screen, HelpScreen)


@pytest.mark.asyncio
async def test_question_mark_closes_help(workspace):
    """? key should close help modal when pressed again."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        await pilot.pause()

        await pilot.press("?")
        await pilot.pause()
        assert isinstance(app.screen, HelpScreen)

        await pilot.press("?")
        await pilot.pause()
        assert len(app.screen_stack) == 1
