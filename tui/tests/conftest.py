"""Pytest fixtures for Bearing TUI tests."""
import shutil

import pytest

from tests.mock_data import (
    create_normal_workspace,
    create_empty_workspace,
    create_overflow_workspace,
    create_long_names_workspace,
    create_single_workspace,
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
def screenshots_dir(tmp_path):
    """Create a temporary directory for visual regression screenshots."""
    d = tmp_path / "screenshots"
    d.mkdir()
    return d
