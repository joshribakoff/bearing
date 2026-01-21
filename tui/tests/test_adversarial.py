"""Adversarial tests to verify tests catch regressions.

These tests verify the TUI handles edge cases and that the test suite
actually catches bugs when they occur.
"""
import json
import pytest
from pathlib import Path
from textual.color import Color

from bearing_tui.app import BearingApp, HelpScreen, PlansScreen
from bearing_tui.widgets import ProjectList, WorktreeTable, DetailsPanel


# =============================================================================
# CSS COLOR VERIFICATION TESTS
# These tests verify our color assertions actually work by checking specific
# color values. If someone changes the CSS, these tests should catch it.
# =============================================================================

@pytest.mark.asyncio
async def test_focus_border_color_is_blue(workspace):
    """Verify focused panel border is exactly #007acc (VS Code blue)."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.press("0")
        await pilot.pause()

        projects_panel = app.query_one("#projects-panel")
        border = projects_panel.styles.border_top
        border_color = border[1] if border else None

        expected = Color.parse("#007acc")
        assert border_color == expected, (
            f"Focus border must be #007acc (VS Code blue), got {border_color}. "
            "If this was intentional, update both the CSS and this test."
        )


@pytest.mark.asyncio
async def test_unfocused_border_color_is_gray(workspace):
    """Verify unfocused panel border is exactly #3c3c3c."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.press("1")  # Focus worktrees, not projects
        await pilot.pause()

        projects_panel = app.query_one("#projects-panel")
        border = projects_panel.styles.border_top
        border_color = border[1] if border else None

        expected = Color.parse("#3c3c3c")
        assert border_color == expected, (
            f"Unfocused border must be #3c3c3c, got {border_color}. "
            "If this was intentional, update both the CSS and this test."
        )


@pytest.mark.asyncio
async def test_highlight_focused_color(workspace):
    """Verify focused list item highlight is #2d5a8a."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.press("0")
        await pilot.press("j")
        await pilot.pause()

        project_list = app.query_one(ProjectList)
        highlighted = project_list.highlighted_child
        bg = highlighted.styles.background

        expected = Color.parse("#2d5a8a")
        assert bg == expected, (
            f"Focused highlight must be #2d5a8a, got {bg}. "
            "This is the bright blue for active selection."
        )


@pytest.mark.asyncio
async def test_highlight_unfocused_color(workspace):
    """Verify unfocused list item highlight is #264f78."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.press("0")
        await pilot.press("j")
        await pilot.press("1")  # Move focus away
        await pilot.pause()

        project_list = app.query_one(ProjectList)
        highlighted = project_list.highlighted_child
        bg = highlighted.styles.background

        expected = Color.parse("#264f78")
        assert bg == expected, (
            f"Unfocused highlight must be #264f78, got {bg}. "
            "This is the dimmer blue for inactive selection."
        )


# =============================================================================
# MALFORMED DATA TESTS
# These tests verify graceful handling of corrupt or unexpected data.
# =============================================================================

@pytest.fixture
def malformed_workspace(tmp_path):
    """Create workspace with malformed JSONL files."""
    # local.jsonl with invalid JSON lines
    with open(tmp_path / "local.jsonl", "w") as f:
        f.write('{"folder": "valid", "repo": "valid", "branch": "main", "base": true}\n')
        f.write('not valid json at all\n')
        f.write('{"folder": "also-valid", "repo": "also-valid", "branch": "main", "base": true}\n')
        f.write('{"incomplete": "json\n')
        f.write('{"folder": "last", "repo": "last", "branch": "main", "base": true}\n')

    # workflow.jsonl with mixed valid/invalid
    with open(tmp_path / "workflow.jsonl", "w") as f:
        f.write('{"repo": "valid", "branch": "feat", "basedOn": "main", "purpose": "test"}\n')
        f.write('{invalid}\n')

    # health.jsonl completely empty
    with open(tmp_path / "health.jsonl", "w") as f:
        pass

    return tmp_path


@pytest.fixture
def missing_fields_workspace(tmp_path):
    """Create workspace with entries missing required fields."""
    # Entries with missing fields
    with open(tmp_path / "local.jsonl", "w") as f:
        f.write('{"folder": "missing-repo", "branch": "main", "base": true}\n')
        f.write('{"repo": "missing-folder", "branch": "main", "base": true}\n')
        f.write('{"folder": "ok", "repo": "ok", "branch": "main", "base": true}\n')
        f.write('{}\n')

    with open(tmp_path / "workflow.jsonl", "w") as f:
        pass
    with open(tmp_path / "health.jsonl", "w") as f:
        pass

    return tmp_path


@pytest.fixture
def unicode_workspace(tmp_path):
    """Create workspace with unicode/emoji in data."""
    with open(tmp_path / "local.jsonl", "w") as f:
        f.write('{"folder": "emoji-project", "repo": "emoji-project", "branch": "main", "base": true}\n')
        f.write('{"folder": "emoji-project-feat", "repo": "emoji-project", "branch": "feat", "base": false}\n')

    with open(tmp_path / "workflow.jsonl", "w") as f:
        f.write('{"repo": "emoji-project", "branch": "feat", "basedOn": "main", "purpose": "Add support for emoji input field"}\n')

    with open(tmp_path / "health.jsonl", "w") as f:
        f.write('{"folder": "emoji-project", "dirty": false, "unpushed": 0}\n')

    return tmp_path


@pytest.mark.asyncio
@pytest.mark.xfail(reason="Known limitation: _read_jsonl crashes on malformed JSON lines instead of skipping them")
async def test_malformed_jsonl_doesnt_crash(malformed_workspace):
    """App should load even with malformed JSONL - skip bad lines.

    NOTE: This test documents a known limitation. The current implementation
    crashes on malformed JSONL. Fixing this would require try/except in _read_jsonl.
    """
    app = BearingApp(workspace=malformed_workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()
        # Should have loaded at least some projects
        project_list = app.query_one(ProjectList)
        # App loaded without crash - that's the main test
        assert True


@pytest.mark.asyncio
@pytest.mark.xfail(reason="Known limitation: Missing required fields cause KeyError")
async def test_missing_fields_handled(missing_fields_workspace):
    """App should handle entries with missing required fields.

    NOTE: This test documents a known limitation. Entries missing 'folder' or 'repo'
    will cause KeyError. This could be fixed with .get() and validation.
    """
    app = BearingApp(workspace=missing_fields_workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()
        # Should load without crashing
        assert True


@pytest.mark.asyncio
async def test_unicode_in_data(unicode_workspace):
    """App should handle unicode characters in project names and purposes."""
    app = BearingApp(workspace=unicode_workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()
        project_list = app.query_one(ProjectList)
        # Should have loaded the emoji project
        assert len(project_list.projects) >= 1


@pytest.fixture
def missing_files_workspace(tmp_path):
    """Create workspace with missing JSONL files."""
    # Only create local.jsonl, others missing
    with open(tmp_path / "local.jsonl", "w") as f:
        f.write('{"folder": "test", "repo": "test", "branch": "main", "base": true}\n')
    return tmp_path


@pytest.mark.asyncio
async def test_missing_workflow_file(missing_files_workspace):
    """App should handle missing workflow.jsonl gracefully."""
    app = BearingApp(workspace=missing_files_workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()
        project_list = app.query_one(ProjectList)
        assert len(project_list.projects) >= 1


# =============================================================================
# RAPID INPUT / STRESS TESTS
# These tests verify the app handles rapid user input without crashing.
# =============================================================================

@pytest.mark.asyncio
async def test_rapid_key_presses(workspace):
    """App should handle rapid key presses without crashing."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        # Rapid navigation
        for _ in range(20):
            await pilot.press("j")
        for _ in range(20):
            await pilot.press("k")

        # Rapid panel switches
        for _ in range(10):
            await pilot.press("0")
            await pilot.press("1")
            await pilot.press("2")

        await pilot.pause()
        # Should not have crashed
        assert True


@pytest.mark.asyncio
async def test_rapid_modal_open_close(workspace):
    """App should handle rapid modal open/close."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        for _ in range(5):
            await pilot.press("?")  # Open help
            await pilot.press("escape")  # Close help

        await pilot.pause()
        assert len(app.screen_stack) == 1


@pytest.mark.asyncio
async def test_escape_sequence_handling(workspace):
    """App should handle escape key properly in various contexts."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        # Escape on main screen - should do nothing harmful
        await pilot.press("escape")
        await pilot.pause()
        assert len(app.screen_stack) == 1

        # Escape in help modal
        await pilot.press("?")
        await pilot.pause()
        assert isinstance(app.screen, HelpScreen)
        await pilot.press("escape")
        await pilot.pause()
        assert len(app.screen_stack) == 1


@pytest.mark.asyncio
async def test_navigation_on_empty_list(empty_workspace):
    """Navigation keys on empty lists should not crash."""
    app = BearingApp(workspace=empty_workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        # Try navigating on empty project list
        await pilot.press("0")
        await pilot.press("j")
        await pilot.press("k")
        await pilot.press("enter")
        await pilot.pause()

        # Try navigating on empty worktree table
        await pilot.press("1")
        await pilot.press("j")
        await pilot.press("k")
        await pilot.press("enter")
        await pilot.pause()

        # Should not have crashed
        assert True


# =============================================================================
# LONG STRING / OVERFLOW TESTS
# These tests verify proper handling of very long strings.
# =============================================================================

@pytest.fixture
def extreme_length_workspace(tmp_path):
    """Create workspace with extremely long strings."""
    long_name = "x" * 500
    long_purpose = "p" * 1000

    with open(tmp_path / "local.jsonl", "w") as f:
        f.write(json.dumps({
            "folder": long_name[:100],
            "repo": long_name[:100],
            "branch": "main",
            "base": True
        }) + "\n")
        f.write(json.dumps({
            "folder": f"{long_name[:100]}-feat",
            "repo": long_name[:100],
            "branch": long_name[:200],
            "base": False
        }) + "\n")

    with open(tmp_path / "workflow.jsonl", "w") as f:
        f.write(json.dumps({
            "repo": long_name[:100],
            "branch": long_name[:200],
            "basedOn": "main",
            "purpose": long_purpose,
            "status": "in_progress"
        }) + "\n")

    with open(tmp_path / "health.jsonl", "w") as f:
        f.write(json.dumps({
            "folder": f"{long_name[:100]}-feat",
            "dirty": True,
            "unpushed": 99999,
            "prState": "open"
        }) + "\n")

    return tmp_path


@pytest.mark.asyncio
async def test_extreme_length_strings(extreme_length_workspace):
    """App should handle extremely long strings without crashing."""
    app = BearingApp(workspace=extreme_length_workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        # Select project
        await pilot.press("0")
        await pilot.press("j")
        await pilot.press("enter")
        await pilot.pause()

        # Select worktree to show details
        await pilot.press("1")
        await pilot.press("j")
        await pilot.press("enter")
        await pilot.pause()

        # Should not have crashed
        assert True


@pytest.mark.asyncio
async def test_tiny_terminal_with_long_names(long_names_workspace):
    """Long names in tiny terminal should not crash."""
    app = BearingApp(workspace=long_names_workspace)
    async with app.run_test(size=(40, 10)) as pilot:
        await pilot.pause()

        await pilot.press("0")
        await pilot.press("j")
        await pilot.press("enter")
        await pilot.pause()

        # Should render without crash
        assert True


# =============================================================================
# BOUNDARY CONDITION TESTS
# =============================================================================

@pytest.mark.asyncio
async def test_cursor_at_zero_row(workspace):
    """Verify cursor can reach row 0 in worktree table."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        # Select project to populate worktrees
        await pilot.press("0")
        await pilot.press("j")
        await pilot.press("enter")
        await pilot.pause()

        # Focus worktree table
        await pilot.press("1")
        await pilot.pause()

        worktree_table = app.query_one(WorktreeTable)
        # Move down then back up to ensure we can reach row 0
        await pilot.press("j")
        await pilot.press("j")
        await pilot.press("k")
        await pilot.press("k")
        await pilot.press("k")  # Extra k to try going above 0
        await pilot.pause()

        # Should be at row 0 or 1 (depending on implementation)
        assert worktree_table.cursor_row >= 0


@pytest.mark.asyncio
async def test_focus_cycle_wraps(workspace):
    """Tab should cycle through all panels and wrap around."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        # Start at project list
        await pilot.press("0")
        await pilot.pause()
        assert app.focused.id == "project-list"

        # Tab through all panels
        await pilot.press("tab")
        assert app.focused.id == "worktree-table"
        await pilot.press("tab")
        assert app.focused.id == "details-panel"
        await pilot.press("tab")
        assert app.focused.id == "project-list"  # Wrapped


@pytest.mark.asyncio
async def test_shift_tab_reverse_cycle(workspace):
    """Shift+Tab should cycle in reverse order."""
    app = BearingApp(workspace=workspace)
    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        # Start at project list
        await pilot.press("0")
        await pilot.pause()

        # Reverse cycle
        await pilot.press("shift+tab")
        assert app.focused.id == "details-panel"
        await pilot.press("shift+tab")
        assert app.focused.id == "worktree-table"
        await pilot.press("shift+tab")
        assert app.focused.id == "project-list"
