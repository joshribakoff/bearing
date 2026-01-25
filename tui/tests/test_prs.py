"""Tests for PR browser functionality."""
import pytest
from bearing_tui.app import BearingApp, PRsScreen
from bearing_tui.widgets import PRsTable


@pytest.fixture
def prs_app(prs_workspace):
    """Create app with PR data."""
    return BearingApp(workspace=prs_workspace)


async def test_prs_modal_opens(prs_app):
    """Test that R key opens the PRs modal."""
    async with prs_app.run_test() as pilot:
        await pilot.pause()
        await pilot.press("R")
        await pilot.pause()
        # Check modal is shown
        assert len(prs_app.screen_stack) > 1
        assert isinstance(prs_app.screen, PRsScreen)


async def test_prs_modal_shows_prs(prs_app):
    """Test that PRs are displayed in the modal."""
    async with prs_app.run_test() as pilot:
        await pilot.pause()
        await pilot.press("R")
        await pilot.pause()
        prs_table = prs_app.screen.query_one(PRsTable)
        # Should have multiple rows (PRs)
        assert prs_table.row_count > 0


async def test_prs_modal_closes_on_escape(prs_app):
    """Test that Escape closes the PRs modal."""
    async with prs_app.run_test() as pilot:
        await pilot.pause()
        await pilot.press("R")
        await pilot.pause()
        await pilot.press("escape")
        await pilot.pause()
        # Should be back to main screen
        assert len(prs_app.screen_stack) == 1


async def test_prs_modal_navigation(prs_app):
    """Test j/k navigation in PRs modal."""
    async with prs_app.run_test() as pilot:
        await pilot.pause()
        await pilot.press("R")
        await pilot.pause()
        prs_table = prs_app.screen.query_one(PRsTable)
        initial_cursor = prs_table.cursor_row
        await pilot.press("j")
        await pilot.pause()
        assert prs_table.cursor_row == initial_cursor + 1
        await pilot.press("k")
        await pilot.pause()
        assert prs_table.cursor_row == initial_cursor


async def test_prs_sorted_open_first(prs_app):
    """Test that OPEN PRs are sorted before MERGED."""
    async with prs_app.run_test() as pilot:
        await pilot.pause()
        await pilot.press("R")
        await pilot.pause()
        prs_table = prs_app.screen.query_one(PRsTable)
        # First PR should be OPEN (check display entry state)
        if prs_table._prs:
            # OPEN and DRAFT states come before MERGED/CLOSED
            open_states = {"OPEN", "DRAFT"}
            first_pr = prs_table._prs[0]
            assert first_pr.state in open_states


async def test_prs_shows_linked_worktree(prs_app):
    """Test that PRs show linked worktree folder."""
    async with prs_app.run_test() as pilot:
        await pilot.pause()
        await pilot.press("R")
        await pilot.pause()
        prs_table = prs_app.screen.query_one(PRsTable)
        # Check that at least one PR has a linked worktree
        has_worktree = any(pr.worktree for pr in prs_table._prs)
        assert has_worktree, "Should have at least one PR with a linked worktree"
