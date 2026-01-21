"""Tests for Bearing TUI using Textual's testing framework."""
import pytest
from pathlib import Path
import tempfile
import json

from bearing_tui.app import BearingApp, HelpScreen
from bearing_tui.widgets import ProjectList, WorktreeTable, DetailsPanel


@pytest.fixture
def workspace(tmp_path):
    """Create a temporary workspace with state files."""
    # Create local.jsonl
    local_entries = [
        {"folder": "myapp", "repo": "myapp", "branch": "main", "base": True},
        {"folder": "myapp-feature", "repo": "myapp", "branch": "feature", "base": False},
        {"folder": "other", "repo": "other", "branch": "main", "base": True},
    ]
    with open(tmp_path / "local.jsonl", "w") as f:
        for entry in local_entries:
            f.write(json.dumps(entry) + "\n")

    # Create workflow.jsonl
    workflow_entries = [
        {
            "repo": "myapp",
            "branch": "feature",
            "basedOn": "main",
            "purpose": "Add new feature",
            "status": "in_progress",
            "created": "2026-01-19T12:00:00Z",
        },
    ]
    with open(tmp_path / "workflow.jsonl", "w") as f:
        for entry in workflow_entries:
            f.write(json.dumps(entry) + "\n")

    # Create health.jsonl
    health_entries = [
        {"folder": "myapp", "dirty": False, "unpushed": 0, "prState": None},
        {"folder": "myapp-feature", "dirty": True, "unpushed": 2, "prState": "open"},
    ]
    with open(tmp_path / "health.jsonl", "w") as f:
        for entry in health_entries:
            f.write(json.dumps(entry) + "\n")

    return tmp_path


@pytest.mark.asyncio
async def test_app_loads(workspace):
    """Test that the app loads without errors."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        # App should have loaded projects
        project_list = app.query_one(ProjectList)
        assert project_list is not None


@pytest.mark.asyncio
async def test_project_list_populated(workspace):
    """Test that project list shows repos from local.jsonl."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        # State should have projects from local.jsonl
        projects = app.state.get_projects()
        assert "myapp" in projects
        assert "other" in projects
        assert len(projects) == 2

        # Project list should be populated after refresh on mount
        project_list = app.query_one(ProjectList)
        assert len(project_list.projects) == 2


@pytest.mark.asyncio
async def test_keyboard_navigation(workspace):
    """Test 0-indexed panel keyboard navigation."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        # Fresh start (no session) focuses project list
        assert isinstance(app.focused, ProjectList)

        # Press 1 to focus worktree table
        await pilot.press("1")
        assert isinstance(app.focused, WorktreeTable)

        # Press 0 to focus project list
        await pilot.press("0")
        assert isinstance(app.focused, ProjectList)

        # Press 2 to focus details panel
        await pilot.press("2")
        assert isinstance(app.focused, DetailsPanel)


@pytest.mark.asyncio
async def test_help_modal(workspace):
    """Test that ? key shows help modal."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        # Press ? to show help
        await pilot.press("question_mark")

        # Help screen should be pushed
        assert len(app.screen_stack) == 2
        assert isinstance(app.screen, HelpScreen)

        # Press escape to close
        await pilot.press("escape")
        assert len(app.screen_stack) == 1


@pytest.mark.asyncio
async def test_refresh_action(workspace):
    """Test that 'r' key refreshes data."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        # Press r to refresh
        await pilot.press("r")

        # Should show notification
        # (Textual's notify creates a Toast widget)


@pytest.mark.asyncio
async def test_tab_navigation(workspace):
    """Test Tab key cycles through panels."""
    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        # Fresh start (no session) focuses project list
        assert app.focused.id == "project-list"

        # Tab to worktree table
        await pilot.press("tab")
        assert app.focused.id == "worktree-table"

        # Tab to details panel
        await pilot.press("tab")
        assert app.focused.id == "details-panel"

        # Tab wraps back to project list
        await pilot.press("tab")
        assert app.focused.id == "project-list"


@pytest.mark.asyncio
async def test_projects_panel_focus_border(workspace):
    """Test that projects panel gets blue border when focused."""
    from textual.color import Color

    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        # Focus the project list
        await pilot.press("0")
        await pilot.pause()

        # Get computed styles for projects panel
        projects_panel = app.query_one("#projects-panel")
        worktrees_panel = app.query_one("#worktrees-panel")

        # Verify focus is on project list
        assert app.focused.id == "project-list"

        # Check border colors - projects should be blue (focus), worktrees should be gray
        # Border is a tuple of ((edge_type, color), ...) for each edge
        # We check the top border color
        projects_border = projects_panel.styles.border_top
        worktrees_border = worktrees_panel.styles.border_top

        # The focused panel should have blue border (#007acc)
        # The unfocused panel should have gray border (#3c3c3c)
        blue_focus = Color.parse("#007acc")
        gray_border = Color.parse("#3c3c3c")

        # Extract border color (border_top is a tuple of (edge_type, color))
        projects_color = projects_border[1] if projects_border else None
        worktrees_color = worktrees_border[1] if worktrees_border else None

        assert projects_color == blue_focus, f"Projects panel should have blue border when focused, got {projects_color}"
        assert worktrees_color == gray_border, f"Worktrees panel should have gray border when not focused, got {worktrees_color}"


@pytest.mark.asyncio
async def test_worktrees_panel_focus_border(workspace):
    """Test that worktrees panel gets blue border when focused."""
    from textual.color import Color

    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        # Focus the worktree table
        await pilot.press("1")
        await pilot.pause()

        # Get computed styles for both panels
        projects_panel = app.query_one("#projects-panel")
        worktrees_panel = app.query_one("#worktrees-panel")

        # Verify focus is on worktree table
        assert app.focused.id == "worktree-table"

        # Check border colors
        projects_border = projects_panel.styles.border_top
        worktrees_border = worktrees_panel.styles.border_top

        blue_focus = Color.parse("#007acc")
        gray_border = Color.parse("#3c3c3c")

        projects_color = projects_border[1] if projects_border else None
        worktrees_color = worktrees_border[1] if worktrees_border else None

        assert projects_color == gray_border, f"Projects panel should have gray border when not focused, got {projects_color}"
        assert worktrees_color == blue_focus, f"Worktrees panel should have blue border when focused, got {worktrees_color}"


@pytest.mark.asyncio
async def test_project_list_item_highlight(workspace):
    """Test that highlighted item in ProjectList has blue background."""
    from textual.color import Color

    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        # Focus the project list
        await pilot.press("0")
        await pilot.pause()

        # Move down to ensure we have a highlighted item
        await pilot.press("j")
        await pilot.pause()

        project_list = app.query_one(ProjectList)
        assert project_list.highlighted_child is not None, "Should have a highlighted item after navigation"

        highlighted_item = project_list.highlighted_child
        bg = highlighted_item.styles.background

        # When focused, should have bright blue (#2d5a8a)
        bright_blue = Color.parse("#2d5a8a")
        assert bg == bright_blue, f"Highlighted item should have bright blue background when focused, got {bg}"


@pytest.mark.asyncio
async def test_project_list_item_highlight_unfocused(workspace):
    """Test that highlighted item in ProjectList retains blue when panel loses focus."""
    from textual.color import Color

    app = BearingApp(workspace=workspace)
    async with app.run_test() as pilot:
        # Focus the project list and move to first item
        await pilot.press("0")
        await pilot.pause()
        # Move down to ensure we have a highlighted item
        await pilot.press("j")
        await pilot.pause()

        # Now move focus to worktrees panel
        await pilot.press("1")
        await pilot.pause()

        # The project list should still have a highlighted item with dimmer blue
        project_list = app.query_one(ProjectList)
        assert project_list.highlighted_child is not None, "Should still have a highlighted item"

        highlighted_item = project_list.highlighted_child
        bg = highlighted_item.styles.background

        # When unfocused, should have selection blue (#264f78)
        selection_blue = Color.parse("#264f78")
        assert bg == selection_blue, f"Highlighted item should retain blue background when unfocused, got {bg}"
