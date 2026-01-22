"""Visual regression tests - generates screenshots for different TUI states and scenarios."""
import pytest
from pathlib import Path


@pytest.fixture
def screenshots_dir(tmp_path):
    """Create screenshots directory."""
    d = tmp_path / "screenshots"
    d.mkdir()
    return d


# =============================================================================
# NORMAL SCENARIO - Standard usage patterns
# =============================================================================

@pytest.mark.asyncio
async def test_visual_projects_focused(workspace, screenshots_dir):
    """Screenshot with projects panel focused."""
    from bearing_tui.app import BearingApp

    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()
        await pilot.press("0")
        await pilot.pause()
        app.save_screenshot(str(screenshots_dir / "normal-01-projects-focused.svg"))


@pytest.mark.asyncio
async def test_visual_worktrees_focused(workspace, screenshots_dir):
    """Screenshot with worktrees panel focused."""
    from bearing_tui.app import BearingApp

    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()
        await pilot.press("1")
        await pilot.pause()
        await pilot.press("j")
        await pilot.pause()
        app.save_screenshot(str(screenshots_dir / "normal-02-worktrees-focused.svg"))


@pytest.mark.asyncio
async def test_visual_help_modal(workspace, screenshots_dir):
    """Screenshot with help modal open."""
    from bearing_tui.app import BearingApp

    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()
        await pilot.press("?")
        await pilot.pause()
        app.save_screenshot(str(screenshots_dir / "normal-03-help-modal.svg"))


@pytest.mark.asyncio
async def test_visual_highlight_panel0_focused(workspace, screenshots_dir):
    """Panel 0 highlight when focused."""
    from bearing_tui.app import BearingApp

    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()
        await pilot.press("0")
        await pilot.press("j")
        await pilot.pause()
        app.save_screenshot(str(screenshots_dir / "normal-04a-panel0-focused.svg"))


@pytest.mark.asyncio
async def test_visual_highlight_panel0_unfocused(workspace, screenshots_dir):
    """Panel 0 highlight when unfocused (panel 1 has focus)."""
    from bearing_tui.app import BearingApp

    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()
        await pilot.press("0")
        await pilot.press("j")
        await pilot.press("1")  # Move focus to panel 1
        await pilot.pause()
        app.save_screenshot(str(screenshots_dir / "normal-04b-panel0-unfocused.svg"))


@pytest.mark.asyncio
async def test_visual_highlight_panel1_focused(workspace, screenshots_dir):
    """Panel 1 cursor when focused."""
    from bearing_tui.app import BearingApp

    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()
        await pilot.press("1")
        await pilot.press("j")
        await pilot.pause()
        app.save_screenshot(str(screenshots_dir / "normal-04c-panel1-focused.svg"))


@pytest.mark.asyncio
async def test_visual_highlight_panel1_unfocused(workspace, screenshots_dir):
    """Panel 1 cursor when unfocused (panel 0 has focus)."""
    from bearing_tui.app import BearingApp

    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()
        await pilot.press("1")
        await pilot.press("j")
        await pilot.press("0")  # Move focus to panel 0
        await pilot.pause()
        app.save_screenshot(str(screenshots_dir / "normal-04d-panel1-unfocused.svg"))


# =============================================================================
# EMPTY SCENARIO - No data
# =============================================================================

@pytest.mark.asyncio
async def test_visual_empty_workspace(empty_workspace, screenshots_dir):
    """Screenshot of empty workspace state."""
    from bearing_tui.app import BearingApp

    app = BearingApp(workspace=empty_workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()
        app.save_screenshot(str(screenshots_dir / "empty-01-no-projects.svg"))


# =============================================================================
# OVERFLOW SCENARIO - Many items (scroll testing)
# =============================================================================

@pytest.mark.asyncio
async def test_visual_overflow_initial(overflow_workspace, screenshots_dir):
    """Screenshot with many projects (scrollable)."""
    from bearing_tui.app import BearingApp

    app = BearingApp(workspace=overflow_workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()
        app.save_screenshot(str(screenshots_dir / "overflow-01-initial.svg"))


@pytest.mark.asyncio
async def test_visual_overflow_scrolled(overflow_workspace, screenshots_dir):
    """Screenshot after scrolling down in project list."""
    from bearing_tui.app import BearingApp

    app = BearingApp(workspace=overflow_workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()
        # Navigate down many times to trigger scroll
        for _ in range(15):
            await pilot.press("j")
        await pilot.pause()
        app.save_screenshot(str(screenshots_dir / "overflow-02-scrolled.svg"))


@pytest.mark.asyncio
async def test_visual_overflow_many_worktrees(overflow_workspace, screenshots_dir):
    """Screenshot with many worktrees for a project."""
    from bearing_tui.app import BearingApp

    app = BearingApp(workspace=overflow_workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()
        await pilot.press("1")  # Focus worktree panel
        await pilot.pause()
        app.save_screenshot(str(screenshots_dir / "overflow-03-many-worktrees.svg"))


# =============================================================================
# LONG NAMES SCENARIO - Truncation testing
# =============================================================================

@pytest.mark.asyncio
async def test_visual_long_names(long_names_workspace, screenshots_dir):
    """Screenshot with very long project/branch names."""
    from bearing_tui.app import BearingApp

    app = BearingApp(workspace=long_names_workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()
        app.save_screenshot(str(screenshots_dir / "longnames-01-truncation.svg"))


@pytest.mark.asyncio
async def test_visual_long_names_details(long_names_workspace, screenshots_dir):
    """Screenshot of details panel with long names."""
    from bearing_tui.app import BearingApp

    app = BearingApp(workspace=long_names_workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()
        await pilot.press("1")
        await pilot.press("j")
        await pilot.pause()
        app.save_screenshot(str(screenshots_dir / "longnames-02-details.svg"))


# =============================================================================
# SINGLE ITEM SCENARIO - Edge case
# =============================================================================

@pytest.mark.asyncio
async def test_visual_single_project(single_workspace, screenshots_dir):
    """Screenshot with only one project and one worktree."""
    from bearing_tui.app import BearingApp

    app = BearingApp(workspace=single_workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()
        app.save_screenshot(str(screenshots_dir / "single-01-one-item.svg"))


# =============================================================================
# RESIZE SCENARIOS - Different terminal sizes
# =============================================================================

@pytest.mark.asyncio
async def test_visual_narrow_terminal(workspace, screenshots_dir):
    """Screenshot in narrow terminal (< 80 cols)."""
    from bearing_tui.app import BearingApp

    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(60, 25)) as pilot:
        await pilot.pause()
        app.save_screenshot(str(screenshots_dir / "resize-01-narrow.svg"))


@pytest.mark.asyncio
async def test_visual_short_terminal(workspace, screenshots_dir):
    """Screenshot in short terminal (< 20 rows)."""
    from bearing_tui.app import BearingApp

    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 15)) as pilot:
        await pilot.pause()
        app.save_screenshot(str(screenshots_dir / "resize-02-short.svg"))


@pytest.mark.asyncio
async def test_visual_tiny_terminal(workspace, screenshots_dir):
    """Screenshot in very small terminal."""
    from bearing_tui.app import BearingApp

    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(50, 12)) as pilot:
        await pilot.pause()
        app.save_screenshot(str(screenshots_dir / "resize-03-tiny.svg"))
