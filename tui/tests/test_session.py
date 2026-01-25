"""Tests for session persistence functionality."""
import json
import pytest
from pathlib import Path
from datetime import datetime, timedelta
from unittest.mock import patch, PropertyMock

from bearing_tui.app import BearingApp
from bearing_tui.widgets import ProjectList, WorktreeTable, DetailsPanel


# =============================================================================
# SESSION FILE TESTS
# =============================================================================

@pytest.mark.asyncio
async def test_session_file_created_on_quit(workspace, tmp_path):
    """Session file should be created when quitting the app."""
    session_dir = tmp_path / ".bearing"
    session_dir.mkdir(parents=True)
    session_file = session_dir / "tui-session.json"

    with patch.object(BearingApp, '_session_file', new_callable=PropertyMock) as mock_prop:
        mock_prop.return_value = session_file

        app = BearingApp(workspace=workspace)

        async with app.run_test(size=(100, 25)) as pilot:
            await pilot.pause()

            # Select a project
            await pilot.press("0")
            await pilot.press("j")
            await pilot.press("enter")
            await pilot.pause()

            # Quit (which saves session)
            await pilot.press("q")

    # Check session file exists and has correct structure
    assert session_file.exists(), "Session file should be created on quit"

    session = json.loads(session_file.read_text())
    assert "selected_project" in session
    assert "timestamp" in session


@pytest.mark.asyncio
async def test_session_saves_selected_project(workspace, tmp_path):
    """Session should save the currently selected project."""
    session_dir = tmp_path / ".bearing"
    session_dir.mkdir(parents=True)
    session_file = session_dir / "tui-session.json"

    with patch.object(BearingApp, '_session_file', new_callable=PropertyMock) as mock_prop:
        mock_prop.return_value = session_file

        app = BearingApp(workspace=workspace)

        async with app.run_test(size=(100, 25)) as pilot:
            await pilot.pause()

            # Select a project
            await pilot.press("0")
            await pilot.press("j")
            await pilot.press("enter")
            await pilot.pause()

            # Get the selected project
            selected_project = app._current_project
            assert selected_project is not None

            await pilot.press("q")

    session = json.loads(session_file.read_text())
    assert session["selected_project"] == selected_project


@pytest.mark.asyncio
async def test_session_saves_focused_panel(workspace, tmp_path):
    """Session should save which panel has focus."""
    session_dir = tmp_path / ".bearing"
    session_dir.mkdir(parents=True)
    session_file = session_dir / "tui-session.json"

    with patch.object(BearingApp, '_session_file', new_callable=PropertyMock) as mock_prop:
        mock_prop.return_value = session_file

        app = BearingApp(workspace=workspace)

        async with app.run_test(size=(100, 25)) as pilot:
            await pilot.pause()

            # Focus the worktree panel
            await pilot.press("1")
            await pilot.pause()

            await pilot.press("q")

    session = json.loads(session_file.read_text())
    assert session["focused_panel"] == "worktree-table"


@pytest.mark.asyncio
async def test_session_saves_worktree_cursor(workspace, tmp_path):
    """Session should save worktree table cursor position."""
    session_dir = tmp_path / ".bearing"
    session_dir.mkdir(parents=True)
    session_file = session_dir / "tui-session.json"

    with patch.object(BearingApp, '_session_file', new_callable=PropertyMock) as mock_prop:
        mock_prop.return_value = session_file

        app = BearingApp(workspace=workspace)

        async with app.run_test(size=(100, 25)) as pilot:
            await pilot.pause()

            # Select project to populate worktrees
            await pilot.press("0")
            await pilot.press("j")
            await pilot.press("enter")
            await pilot.pause()

            # Move cursor in worktree table
            await pilot.press("1")
            await pilot.press("j")
            await pilot.press("j")
            await pilot.pause()

            worktree_table = app.query_one(WorktreeTable)
            cursor_row = worktree_table.cursor_row

            await pilot.press("q")

    session = json.loads(session_file.read_text())
    assert session["worktree_cursor"] == cursor_row


# =============================================================================
# SESSION RESTORE TESTS
# =============================================================================

@pytest.mark.asyncio
async def test_session_restores_focused_panel(workspace, tmp_path):
    """Session should restore focus to the saved panel."""
    session_dir = tmp_path / ".bearing"
    session_dir.mkdir(parents=True)
    session_file = session_dir / "tui-session.json"

    # Create a recent session file
    session = {
        "selected_project": None,
        "project_index": None,
        "worktree_cursor": None,
        "focused_panel": "worktree-table",
        "timestamp": datetime.now().isoformat(),
    }
    session_file.write_text(json.dumps(session))

    with patch.object(BearingApp, '_session_file', new_callable=PropertyMock) as mock_prop:
        mock_prop.return_value = session_file

        app = BearingApp(workspace=workspace)

        async with app.run_test(size=(100, 25)) as pilot:
            await pilot.pause()
            await pilot.pause()  # Extra pause for session restore

            # Focus should be on worktree table
            assert app.focused.id == "worktree-table"


@pytest.mark.asyncio
async def test_session_expired_not_restored(workspace, tmp_path):
    """Session older than 24 hours should not be restored."""
    session_dir = tmp_path / ".bearing"
    session_dir.mkdir(parents=True)
    session_file = session_dir / "tui-session.json"

    # Create an old session file (>24 hours)
    old_time = datetime.now() - timedelta(hours=25)
    session = {
        "selected_project": None,
        "project_index": None,
        "worktree_cursor": None,
        "focused_panel": "worktree-table",
        "timestamp": old_time.isoformat(),
    }
    session_file.write_text(json.dumps(session))

    with patch.object(BearingApp, '_session_file', new_callable=PropertyMock) as mock_prop:
        mock_prop.return_value = session_file

        app = BearingApp(workspace=workspace)

        async with app.run_test(size=(100, 25)) as pilot:
            await pilot.pause()

            # Focus should default to project-list, not worktree-table
            assert app.focused.id == "project-list"


@pytest.mark.asyncio
async def test_session_missing_file_handles_gracefully(workspace, tmp_path):
    """App should start normally when session file doesn't exist."""
    session_dir = tmp_path / ".bearing"
    session_dir.mkdir(parents=True)
    session_file = session_dir / "tui-session.json"
    # Don't create the file

    with patch.object(BearingApp, '_session_file', new_callable=PropertyMock) as mock_prop:
        mock_prop.return_value = session_file

        app = BearingApp(workspace=workspace)

        async with app.run_test(size=(100, 25)) as pilot:
            await pilot.pause()

            # Should start with default focus
            assert app.focused.id == "project-list"


@pytest.mark.asyncio
async def test_session_corrupt_file_handles_gracefully(workspace, tmp_path):
    """App should start normally when session file is corrupt."""
    session_dir = tmp_path / ".bearing"
    session_dir.mkdir(parents=True)
    session_file = session_dir / "tui-session.json"
    session_file.write_text("not valid json {{{")

    with patch.object(BearingApp, '_session_file', new_callable=PropertyMock) as mock_prop:
        mock_prop.return_value = session_file

        app = BearingApp(workspace=workspace)

        async with app.run_test(size=(100, 25)) as pilot:
            await pilot.pause()

            # Should start with default focus, not crash
            assert app.focused.id == "project-list"


# =============================================================================
# SESSION TIMESTAMP TESTS
# =============================================================================

@pytest.mark.asyncio
async def test_session_has_valid_timestamp(workspace, tmp_path):
    """Session file should have ISO format timestamp."""
    session_dir = tmp_path / ".bearing"
    session_dir.mkdir(parents=True)
    session_file = session_dir / "tui-session.json"

    with patch.object(BearingApp, '_session_file', new_callable=PropertyMock) as mock_prop:
        mock_prop.return_value = session_file

        app = BearingApp(workspace=workspace)

        async with app.run_test(size=(100, 25)) as pilot:
            await pilot.pause()
            await pilot.press("q")

    session = json.loads(session_file.read_text())
    timestamp_str = session["timestamp"]

    # Should be parseable as ISO format
    timestamp = datetime.fromisoformat(timestamp_str)
    # Should be recent (within last minute)
    assert datetime.now() - timestamp < timedelta(minutes=1)


# =============================================================================
# DAEMON AUTO-START TESTS
# =============================================================================

@pytest.mark.asyncio
async def test_daemon_auto_start_called_on_mount(workspace):
    """_ensure_daemon_running should be called when app mounts."""
    daemon_started = []

    def mock_ensure_daemon(self):
        daemon_started.append(True)
        return

    with patch.object(BearingApp, '_ensure_daemon_running', mock_ensure_daemon):
        app = BearingApp(workspace=workspace)

        async with app.run_test(size=(100, 25)) as pilot:
            await pilot.pause()

    assert len(daemon_started) > 0, "_ensure_daemon_running should be called on mount"


@pytest.mark.asyncio
async def test_daemon_action_doesnt_crash_when_not_found(workspace):
    """Press 'd' should not crash when daemon not found."""
    import subprocess as sp

    def mock_run(*args, **kwargs):
        raise FileNotFoundError("bearing not found")

    app = BearingApp(workspace=workspace)

    async with app.run_test(size=(100, 25)) as pilot:
        await pilot.pause()

        # Patch subprocess.run for this specific action
        with patch("bearing_tui.app.subprocess.run", side_effect=FileNotFoundError):
            await pilot.press("d")
            await pilot.pause()

        # Should not crash even if daemon not found
        assert True


# =============================================================================
# PROJECT INDEX PERSISTENCE TESTS
# =============================================================================

@pytest.mark.asyncio
async def test_session_saves_project_index(workspace, tmp_path):
    """Session should save the project list index."""
    session_dir = tmp_path / ".bearing"
    session_dir.mkdir(parents=True)
    session_file = session_dir / "tui-session.json"

    with patch.object(BearingApp, '_session_file', new_callable=PropertyMock) as mock_prop:
        mock_prop.return_value = session_file

        app = BearingApp(workspace=workspace)

        async with app.run_test(size=(100, 25)) as pilot:
            await pilot.pause()

            # Navigate to second item
            await pilot.press("0")
            await pilot.press("j")
            await pilot.press("j")  # Move to second item
            await pilot.pause()

            project_list = app.query_one(ProjectList)
            saved_index = project_list.index

            await pilot.press("q")

    session = json.loads(session_file.read_text())
    assert session["project_index"] == saved_index
