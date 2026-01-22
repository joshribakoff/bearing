"""Comprehensive mock data scenarios for TUI testing."""
import json
import tempfile
from pathlib import Path
from typing import Callable


def _write_jsonl(path: Path, entries: list[dict]) -> None:
    """Write entries to a JSONL file."""
    with open(path, "w") as f:
        for entry in entries:
            f.write(json.dumps(entry) + "\n")


def create_normal_workspace() -> Path:
    """Create a workspace with 5-6 projects, 2-3 worktrees each, mixed states.

    Simulates typical development workflow with:
    - Some clean base branches
    - Some dirty feature branches with unpushed commits
    - Various PR states (open, merged, draft)
    """
    workspace = Path(tempfile.mkdtemp(prefix="bearing_normal_"))

    local_entries = [
        # sailkit: base + 2 features
        {"folder": "sailkit", "repo": "sailkit", "branch": "main", "base": True},
        {"folder": "sailkit-compass-refactor", "repo": "sailkit", "branch": "compass-refactor", "base": False},
        {"folder": "sailkit-add-lantern-themes", "repo": "sailkit", "branch": "add-lantern-themes", "base": False},
        # bearing: base + 2 features
        {"folder": "bearing", "repo": "bearing", "branch": "main", "base": True},
        {"folder": "bearing-tui-improvements", "repo": "bearing", "branch": "tui-improvements", "base": False},
        {"folder": "bearing-health-checks", "repo": "bearing", "branch": "health-checks", "base": False},
        # fightingwithai: base + 1 feature
        {"folder": "fightingwithai.com", "repo": "fightingwithai.com", "branch": "main", "base": True},
        {"folder": "fightingwithai.com-vim-mode", "repo": "fightingwithai.com", "branch": "vim-mode", "base": False},
        # surfdeeper: base + 3 features
        {"folder": "surfdeeper", "repo": "surfdeeper", "branch": "main", "base": True},
        {"folder": "surfdeeper-wave-forecast", "repo": "surfdeeper", "branch": "wave-forecast", "base": False},
        {"folder": "surfdeeper-spot-search", "repo": "surfdeeper", "branch": "spot-search", "base": False},
        {"folder": "surfdeeper-tide-charts", "repo": "surfdeeper", "branch": "tide-charts", "base": False},
        # portfolio: base only
        {"folder": "portfolio", "repo": "portfolio", "branch": "main", "base": True},
        # api-server: base + 1 feature
        {"folder": "api-server", "repo": "api-server", "branch": "main", "base": True},
        {"folder": "api-server-auth-refactor", "repo": "api-server", "branch": "auth-refactor", "base": False},
    ]

    workflow_entries = [
        {"repo": "sailkit", "branch": "compass-refactor", "basedOn": "main", "purpose": "Refactor compass component for better tree-shaking", "status": "in_progress", "created": "2026-01-15T10:00:00Z"},
        {"repo": "sailkit", "branch": "add-lantern-themes", "basedOn": "main", "purpose": "Add dark/light/system theme support to lantern", "status": "review", "created": "2026-01-18T14:30:00Z"},
        {"repo": "bearing", "branch": "tui-improvements", "basedOn": "main", "purpose": "Improve TUI keyboard navigation and focus handling", "status": "in_progress", "created": "2026-01-19T09:00:00Z"},
        {"repo": "bearing", "branch": "health-checks", "basedOn": "main", "purpose": "Add automated health monitoring for worktrees", "status": "done", "created": "2026-01-10T08:00:00Z"},
        {"repo": "fightingwithai.com", "branch": "vim-mode", "basedOn": "main", "purpose": "Implement vim-style keyboard navigation", "status": "review", "created": "2026-01-17T11:00:00Z"},
        {"repo": "surfdeeper", "branch": "wave-forecast", "basedOn": "main", "purpose": "Integrate NOAA wave forecast data", "status": "in_progress", "created": "2026-01-16T15:00:00Z"},
        {"repo": "surfdeeper", "branch": "spot-search", "basedOn": "main", "purpose": "Add fuzzy search for surf spots", "status": "blocked", "created": "2026-01-12T09:30:00Z"},
        {"repo": "surfdeeper", "branch": "tide-charts", "basedOn": "main", "purpose": "Display tide charts with moon phases", "status": "in_progress", "created": "2026-01-19T16:00:00Z"},
        {"repo": "api-server", "branch": "auth-refactor", "basedOn": "main", "purpose": "Migrate from JWT to session-based auth", "status": "in_progress", "created": "2026-01-14T13:00:00Z"},
    ]

    health_entries = [
        # sailkit
        {"folder": "sailkit", "dirty": False, "unpushed": 0, "prState": None},
        {"folder": "sailkit-compass-refactor", "dirty": True, "unpushed": 3, "prState": "draft"},
        {"folder": "sailkit-add-lantern-themes", "dirty": False, "unpushed": 0, "prState": "open"},
        # bearing
        {"folder": "bearing", "dirty": False, "unpushed": 0, "prState": None},
        {"folder": "bearing-tui-improvements", "dirty": True, "unpushed": 5, "prState": "open"},
        {"folder": "bearing-health-checks", "dirty": False, "unpushed": 0, "prState": "merged"},
        # fightingwithai
        {"folder": "fightingwithai.com", "dirty": False, "unpushed": 0, "prState": None},
        {"folder": "fightingwithai.com-vim-mode", "dirty": False, "unpushed": 1, "prState": "open"},
        # surfdeeper
        {"folder": "surfdeeper", "dirty": True, "unpushed": 0, "prState": None},
        {"folder": "surfdeeper-wave-forecast", "dirty": True, "unpushed": 2, "prState": "draft"},
        {"folder": "surfdeeper-spot-search", "dirty": False, "unpushed": 0, "prState": None},
        {"folder": "surfdeeper-tide-charts", "dirty": True, "unpushed": 1, "prState": None},
        # portfolio
        {"folder": "portfolio", "dirty": False, "unpushed": 0, "prState": None},
        # api-server
        {"folder": "api-server", "dirty": False, "unpushed": 0, "prState": None},
        {"folder": "api-server-auth-refactor", "dirty": True, "unpushed": 8, "prState": "open"},
    ]

    _write_jsonl(workspace / "local.jsonl", local_entries)
    _write_jsonl(workspace / "workflow.jsonl", workflow_entries)
    _write_jsonl(workspace / "health.jsonl", health_entries)

    return workspace


def create_empty_workspace() -> Path:
    """Create workspace with empty JSONL files (no data).

    Tests edge case of brand new workspace with no projects.
    """
    workspace = Path(tempfile.mkdtemp(prefix="bearing_empty_"))

    _write_jsonl(workspace / "local.jsonl", [])
    _write_jsonl(workspace / "workflow.jsonl", [])
    _write_jsonl(workspace / "health.jsonl", [])

    return workspace


def create_overflow_workspace() -> Path:
    """Create workspace with 25 projects, 15 worktrees each for scroll testing.

    Tests:
    - Vertical scrolling in project list
    - Vertical scrolling in worktree table
    - Performance with large datasets
    """
    workspace = Path(tempfile.mkdtemp(prefix="bearing_overflow_"))

    local_entries = []
    workflow_entries = []
    health_entries = []

    for i in range(25):
        project = f"project-{i:02d}"
        # Add base
        local_entries.append({
            "folder": project,
            "repo": project,
            "branch": "main",
            "base": True,
        })
        health_entries.append({
            "folder": project,
            "dirty": i % 7 == 0,  # Some dirty
            "unpushed": 0,
            "prState": None,
        })

        # Add 14 worktrees per project
        for j in range(14):
            branch = f"feature-{j:02d}"
            folder = f"{project}-{branch}"
            local_entries.append({
                "folder": folder,
                "repo": project,
                "branch": branch,
                "base": False,
            })
            workflow_entries.append({
                "repo": project,
                "branch": branch,
                "basedOn": "main",
                "purpose": f"Feature {j} for {project}",
                "status": ["in_progress", "review", "done", "blocked"][j % 4],
                "created": f"2026-01-{(j % 28) + 1:02d}T{(i % 24):02d}:00:00Z",
            })
            health_entries.append({
                "folder": folder,
                "dirty": (i + j) % 3 == 0,
                "unpushed": (i + j) % 5,
                "prState": [None, "draft", "open", "merged"][(i + j) % 4],
            })

    _write_jsonl(workspace / "local.jsonl", local_entries)
    _write_jsonl(workspace / "workflow.jsonl", workflow_entries)
    _write_jsonl(workspace / "health.jsonl", health_entries)

    return workspace


def create_long_names_workspace() -> Path:
    """Create workspace with very long names for truncation testing.

    Tests:
    - Project name truncation (50+ chars)
    - Branch name truncation
    - Purpose text truncation
    - Table column width handling
    """
    workspace = Path(tempfile.mkdtemp(prefix="bearing_longnames_"))

    long_projects = [
        "this-is-an-extremely-long-project-name-that-should-be-truncated",
        "another-ridiculously-long-project-name-for-testing-purposes",
        "super-duper-extra-long-project-name-to-test-ui-boundaries",
    ]

    local_entries = []
    workflow_entries = []
    health_entries = []

    for project in long_projects:
        # Base
        local_entries.append({
            "folder": project,
            "repo": project,
            "branch": "main",
            "base": True,
        })
        health_entries.append({
            "folder": project,
            "dirty": False,
            "unpushed": 0,
            "prState": None,
        })

        # Long branch names
        long_branches = [
            "feature-with-an-incredibly-long-branch-name-that-exceeds-normal-limits",
            "bugfix-for-critical-issue-in-authentication-flow-with-oauth2-provider",
            "refactor-legacy-code-to-use-modern-patterns-and-best-practices",
        ]

        for branch in long_branches:
            folder = f"{project}-{branch}"[:100]  # Limit folder name
            local_entries.append({
                "folder": folder,
                "repo": project,
                "branch": branch,
                "base": False,
            })
            workflow_entries.append({
                "repo": project,
                "branch": branch,
                "basedOn": "main",
                "purpose": (
                    "This is an extremely long purpose description that goes into great detail "
                    "about what this particular feature branch is trying to accomplish and why "
                    "it is important for the overall project goals and user experience improvements"
                ),
                "status": "in_progress",
                "created": "2026-01-15T10:00:00Z",
            })
            health_entries.append({
                "folder": folder,
                "dirty": True,
                "unpushed": 99,  # High number to test display
                "prState": "open",
            })

    _write_jsonl(workspace / "local.jsonl", local_entries)
    _write_jsonl(workspace / "workflow.jsonl", workflow_entries)
    _write_jsonl(workspace / "health.jsonl", health_entries)

    return workspace


def create_single_workspace() -> Path:
    """Create workspace with 1 project, 1 worktree (edge case).

    Tests:
    - Minimal data display
    - Empty state handling when no features exist
    - Single selection behavior
    """
    workspace = Path(tempfile.mkdtemp(prefix="bearing_single_"))

    local_entries = [
        {"folder": "solo-project", "repo": "solo-project", "branch": "main", "base": True},
        {"folder": "solo-project-first-feature", "repo": "solo-project", "branch": "first-feature", "base": False},
    ]

    workflow_entries = [
        {
            "repo": "solo-project",
            "branch": "first-feature",
            "basedOn": "main",
            "purpose": "The very first feature for this project",
            "status": "in_progress",
            "created": "2026-01-20T12:00:00Z",
        },
    ]

    health_entries = [
        {"folder": "solo-project", "dirty": False, "unpushed": 0, "prState": None},
        {"folder": "solo-project-first-feature", "dirty": True, "unpushed": 1, "prState": "draft"},
    ]

    _write_jsonl(workspace / "local.jsonl", local_entries)
    _write_jsonl(workspace / "workflow.jsonl", workflow_entries)
    _write_jsonl(workspace / "health.jsonl", health_entries)

    return workspace


# Map scenario names to factory functions
SCENARIOS: dict[str, Callable[[], Path]] = {
    "normal": create_normal_workspace,
    "empty": create_empty_workspace,
    "overflow": create_overflow_workspace,
    "long_names": create_long_names_workspace,
    "single": create_single_workspace,
}
