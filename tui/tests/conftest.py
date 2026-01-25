"""Pytest fixtures for Bearing TUI tests."""
import shutil
from pathlib import Path

import pytest


@pytest.fixture(autouse=True)
def isolate_session(tmp_path, monkeypatch):
    """Isolate session file to temp directory to avoid affecting real user state.

    pytest's tmp_path fixture creates a UNIQUE temp directory per test, e.g.:
    /tmp/pytest-of-user/pytest-123/test_foo0/
    /tmp/pytest-of-user/pytest-123/test_bar0/
    So tests won't clobber each other.
    """
    # Create temp .bearing directory in this test's unique tmp_path
    temp_bearing = tmp_path / ".bearing"
    temp_bearing.mkdir()

    # Monkeypatch Path.home() to return this test's temp directory
    # This ensures tests don't read/write the user's real ~/.bearing/
    def mock_home():
        return tmp_path

    monkeypatch.setattr(Path, "home", staticmethod(mock_home))
    yield

from tests.mock_data import (
    create_normal_workspace,
    create_empty_workspace,
    create_overflow_workspace,
    create_long_names_workspace,
    create_single_workspace,
    create_prs_workspace,
)


@pytest.fixture
def workspace():
    """Create a normal workspace with standard test data."""
    path = create_normal_workspace()
    yield path
    shutil.rmtree(path, ignore_errors=True)


@pytest.fixture
def empty_workspace():
    """Create an empty workspace with no projects."""
    path = create_empty_workspace()
    yield path
    shutil.rmtree(path, ignore_errors=True)


@pytest.fixture
def overflow_workspace():
    """Create a workspace with many items to test scrolling."""
    path = create_overflow_workspace()
    yield path
    shutil.rmtree(path, ignore_errors=True)


@pytest.fixture
def long_names_workspace():
    """Create a workspace with very long names to test truncation."""
    path = create_long_names_workspace()
    yield path
    shutil.rmtree(path, ignore_errors=True)


@pytest.fixture
def single_workspace():
    """Create a workspace with a single project and worktree."""
    path = create_single_workspace()
    yield path
    shutil.rmtree(path, ignore_errors=True)


@pytest.fixture
def prs_workspace():
    """Create a workspace with PR data for testing."""
    path = create_prs_workspace()
    yield path
    shutil.rmtree(path, ignore_errors=True)


@pytest.fixture
def screenshots_dir(tmp_path):
    """Create a temporary directory for visual regression screenshots."""
    d = tmp_path / "screenshots"
    d.mkdir()
    return d
