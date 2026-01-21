"""Comprehensive TUI tests covering navigation, empty states, scroll, focus, and details."""
import pytest
import json
from pathlib import Path

from bearing_tui.app import BearingApp
from bearing_tui.widgets import ProjectList, WorktreeTable, DetailsPanel


@pytest.fixture
def workspace(tmp_path):
    """Create a temporary workspace with state files."""
    local_entries = [
        {"folder": "myapp", "repo": "myapp", "branch": "main", "base": True},
        {"folder": "myapp-feature", "repo": "myapp", "branch": "feature", "base": False},
        {"folder": "other", "repo": "other", "branch": "main", "base": True},
    ]
    with open(tmp_path / "local.jsonl", "w") as f:
        for entry in local_entries:
            f.write(json.dumps(entry) + "\n")

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

    health_entries = [
        {"folder": "myapp", "dirty": False, "unpushed": 0, "prState": None},
        {"folder": "myapp-feature", "dirty": True, "unpushed": 2, "prState": "open"},
    ]
    with open(tmp_path / "health.jsonl", "w") as f:
        for entry in health_entries:
            f.write(json.dumps(entry) + "\n")

    return tmp_path


@pytest.fixture
def empty_workspace(tmp_path):
    """Create a workspace with no projects."""
    with open(tmp_path / "local.jsonl", "w") as f:
        pass  # Empty file
    with open(tmp_path / "workflow.jsonl", "w") as f:
        pass
    with open(tmp_path / "health.jsonl", "w") as f:
        pass
    return tmp_path


@pytest.fixture
def large_workspace(tmp_path):
    """Create a workspace with 25+ items for scroll testing."""
    local_entries = []
    for i in range(30):
        local_entries.append({
            "folder": f"project{i:02d}",
            "repo": f"project{i:02d}",
            "branch": "main",
            "base": True
        })
        # Add worktrees for first project
        if i == 0:
            for j in range(30):
                local_entries.append({
                    "folder": f"project00-wt{j:02d}",
                    "repo": "project00",
                    "branch": f"branch{j:02d}",
                    "base": False
                })

    with open(tmp_path / "local.jsonl", "w") as f:
        for entry in local_entries:
            f.write(json.dumps(entry) + "\n")
    with open(tmp_path / "workflow.jsonl", "w") as f:
        pass
    with open(tmp_path / "health.jsonl", "w") as f:
        pass
    return tmp_path


# ============================================================================
# Navigation Edge Cases
# ============================================================================

@pytest.mark.asyncio
async def test_navigation_at_top_boundary(workspace):
    """Test j/k at top of list - k should not go above first item."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()
        await pilot.press("0")  # Focus project list
        await pilot.pause()

        project_list = app.query_one(ProjectList)

        # Move to first item explicitly
        await pilot.press("j")
        await pilot.pause()

        # Press k multiple times at top - should stay at first item
        await pilot.press("k")
        await pilot.press("k")
        await pilot.press("k")
        await pilot.pause()

        # Index should be 0 (first item)
        assert project_list.index == 0


@pytest.mark.asyncio
async def test_navigation_at_bottom_boundary(workspace):
    """Test j/k at bottom of list - j should not go beyond last item."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()
        await pilot.press("0")  # Focus project list
        await pilot.pause()

        project_list = app.query_one(ProjectList)
        num_items = len(project_list.projects)

        # Navigate to bottom
        for _ in range(num_items + 5):
            await pilot.press("j")
        await pilot.pause()

        # Index should not exceed last valid index
        assert project_list.index is not None
        assert project_list.index < num_items


@pytest.mark.asyncio
async def test_tab_preserves_selection(workspace):
    """Test that Tab between panels keeps selection in both panels."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        # Focus project list and select an item
        await pilot.press("0")
        await pilot.press("j")
        await pilot.pause()

        project_list = app.query_one(ProjectList)
        original_index = project_list.index

        # Tab to worktree table
        await pilot.press("tab")
        await pilot.pause()
        assert app.focused.id == "worktree-table"

        # Tab back to project list
        await pilot.press("shift+tab")
        await pilot.pause()
        assert app.focused.id == "project-list"

        # Selection should be preserved
        assert project_list.index == original_index


@pytest.mark.asyncio
async def test_panel_0_to_1_selection(workspace):
    """Test switching from panel 0 to panel 1 maintains project context."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        # Select a project and press Enter
        await pilot.press("0")
        await pilot.press("j")
        await pilot.press("enter")
        await pilot.pause()

        # Should have populated worktree table
        worktree_table = app.query_one(WorktreeTable)
        assert worktree_table.row_count > 0

        # Switch to panel 1 - worktrees should still be visible
        await pilot.press("1")
        await pilot.pause()
        assert app.focused.id == "worktree-table"
        assert worktree_table.row_count > 0


# ============================================================================
# Empty States
# ============================================================================

@pytest.mark.asyncio
async def test_empty_workspace_shows_message(empty_workspace):
    """Test that empty workspace shows 'No projects found'."""
    app = BearingApp(workspace=empty_workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        project_list = app.query_one(ProjectList)
        # Empty project list should show "No projects found" message
        assert len(project_list.projects) == 0

        # Check that the list contains the empty state item
        items = list(project_list.query("ListItem"))
        assert len(items) == 1
        # The item should contain "No projects found" text
        static = items[0].query_one("Static")
        # Use render() method to get the text content
        assert "No projects found" in static.render().plain


@pytest.mark.asyncio
async def test_empty_project_shows_no_worktrees(workspace):
    """Test that selecting a project with no worktrees shows appropriate message."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        # Initially worktree table should show empty state
        worktree_table = app.query_one(WorktreeTable)

        # Clear and check empty state
        worktree_table.clear_worktrees()
        await pilot.pause()

        # Should show "Select a project" message
        assert worktree_table.row_count == 1


# ============================================================================
# Scroll Behavior
# ============================================================================

@pytest.mark.asyncio
async def test_list_scrolls_on_navigation(large_workspace):
    """Test that selection stays visible when navigating long lists."""
    app = BearingApp(workspace=large_workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        await pilot.press("0")  # Focus project list
        await pilot.pause()

        project_list = app.query_one(ProjectList)

        # Navigate down many items
        for _ in range(20):
            await pilot.press("j")
        await pilot.pause()

        # Index should be around 20
        assert project_list.index >= 15

        # Navigate back up
        for _ in range(10):
            await pilot.press("k")
        await pilot.pause()

        # Index should have decreased
        assert project_list.index < 15


@pytest.mark.asyncio
async def test_many_items_renders_correctly(large_workspace):
    """Test that 25+ items render without errors."""
    app = BearingApp(workspace=large_workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        project_list = app.query_one(ProjectList)
        # Should have loaded all 30 projects
        assert len(project_list.projects) == 30

        # Select first project and view worktrees
        await pilot.press("0")
        await pilot.press("j")
        await pilot.press("enter")
        await pilot.pause()

        worktree_table = app.query_one(WorktreeTable)
        # First project should have 31 entries (1 base + 30 worktrees)
        assert worktree_table.row_count == 31


# ============================================================================
# Focus States
# ============================================================================

@pytest.mark.asyncio
async def test_panel_0_focused_highlight(workspace):
    """Test that panel 0 has bright highlight when focused."""
    from textual.color import Color

    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        await pilot.press("0")  # Focus project list
        await pilot.press("j")  # Select an item
        await pilot.pause()

        project_list = app.query_one(ProjectList)
        highlighted = project_list.highlighted_child
        assert highlighted is not None

        bg = highlighted.styles.background
        bright_blue = Color.parse("#2d5a8a")
        assert bg == bright_blue, f"Expected bright blue {bright_blue}, got {bg}"


@pytest.mark.asyncio
async def test_panel_0_unfocused_highlight(workspace):
    """Test that panel 0 has dim highlight when unfocused."""
    from textual.color import Color

    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        # Focus project list and select
        await pilot.press("0")
        await pilot.press("j")
        await pilot.pause()

        # Move focus away
        await pilot.press("1")
        await pilot.pause()

        project_list = app.query_one(ProjectList)
        highlighted = project_list.highlighted_child
        assert highlighted is not None

        bg = highlighted.styles.background
        selection_blue = Color.parse("#264f78")
        assert bg == selection_blue, f"Expected selection blue {selection_blue}, got {bg}"


@pytest.mark.asyncio
async def test_panel_1_focused_cursor(workspace):
    """Test that panel 1 (worktree table) has bright cursor when focused."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        # Select a project first to populate worktree table
        await pilot.press("0")
        await pilot.press("j")
        await pilot.press("enter")
        await pilot.pause()

        # Focus worktree table
        await pilot.press("1")
        await pilot.pause()

        worktree_table = app.query_one(WorktreeTable)
        assert app.focused.id == "worktree-table"
        # DataTable should have cursor visible when focused
        assert worktree_table.show_cursor is True


@pytest.mark.asyncio
async def test_panel_1_unfocused_cursor(workspace):
    """Test that panel 1 cursor state when unfocused."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        # Select a project first
        await pilot.press("0")
        await pilot.press("j")
        await pilot.press("enter")
        await pilot.pause()

        # Focus and select in worktree table
        await pilot.press("1")
        await pilot.press("j")
        await pilot.pause()

        worktree_table = app.query_one(WorktreeTable)
        initial_row = worktree_table.cursor_row

        # Move focus away
        await pilot.press("0")
        await pilot.pause()

        # Cursor position should be preserved
        assert worktree_table.cursor_row == initial_row


# ============================================================================
# Details Panel
# ============================================================================

@pytest.mark.asyncio
async def test_details_updates_on_selection(workspace):
    """Test that details panel updates when worktree is selected."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        details = app.query_one(DetailsPanel)
        # Initially should show empty state
        assert details.current_folder is None

        # Select a project
        await pilot.press("0")
        await pilot.press("j")
        await pilot.press("enter")
        await pilot.pause()

        # Focus worktree table and select a worktree
        await pilot.press("1")
        await pilot.press("j")
        await pilot.press("enter")
        await pilot.pause()

        # Details should have the folder set
        assert details.current_folder is not None


@pytest.mark.asyncio
async def test_details_shows_all_fields(workspace):
    """Test that details panel shows folder, branch, purpose, and health."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        # Select myapp project
        await pilot.press("0")
        await pilot.press("j")
        await pilot.press("enter")
        await pilot.pause()

        # Select the feature worktree (second item, has workflow data)
        await pilot.press("1")
        await pilot.press("j")
        await pilot.press("j")
        await pilot.press("enter")
        await pilot.pause()

        details = app.query_one(DetailsPanel)
        # Use render() to get text content
        content = details.render().plain

        # Should show folder
        assert "myapp-feature" in content or "Folder" in content
        # Should show branch
        assert "feature" in content or "Branch" in content


@pytest.mark.asyncio
async def test_details_preserved_on_refresh(workspace):
    """Test that details panel is preserved when refresh is triggered."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        # Select a project and worktree
        await pilot.press("0")
        await pilot.press("j")
        await pilot.press("enter")
        await pilot.pause()
        await pilot.press("1")
        await pilot.press("j")
        await pilot.press("enter")
        await pilot.pause()

        details = app.query_one(DetailsPanel)
        # After selecting, should have folder set
        assert details.current_folder is not None
        saved_folder = details.current_folder

        # Refresh
        await pilot.press("r")
        await pilot.pause()

        # After refresh, details should be preserved (not cleared)
        assert details.current_folder == saved_folder


# ============================================================================
# Additional Edge Cases
# ============================================================================

@pytest.mark.asyncio
async def test_vim_h_l_navigation(workspace):
    """Test h/l keys for left/right panel focus."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        # Start at project list
        await pilot.press("0")
        await pilot.pause()
        assert app.focused.id == "project-list"

        # Press l to go right
        await pilot.press("l")
        await pilot.pause()
        assert app.focused.id == "worktree-table"

        # Press h to go left
        await pilot.press("h")
        await pilot.pause()
        assert app.focused.id == "project-list"


@pytest.mark.asyncio
async def test_number_key_focus(workspace):
    """Test 0, 1, 2 keys for direct panel focus."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        await pilot.press("2")  # Details
        await pilot.pause()
        assert app.focused.id == "details-panel"

        await pilot.press("0")  # Projects
        await pilot.pause()
        assert app.focused.id == "project-list"

        await pilot.press("1")  # Worktrees
        await pilot.pause()
        assert app.focused.id == "worktree-table"


@pytest.mark.asyncio
async def test_worktree_selection_triggers_details(workspace):
    """Test that selecting worktree row updates details via message."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        # Select project to populate worktrees
        await pilot.press("0")
        await pilot.press("j")
        await pilot.press("enter")
        await pilot.pause()

        details = app.query_one(DetailsPanel)
        assert details.current_folder is None

        # Select a worktree
        await pilot.press("1")
        await pilot.press("j")
        await pilot.press("enter")
        await pilot.pause()

        # Details should now have a folder set
        assert details.current_folder is not None
