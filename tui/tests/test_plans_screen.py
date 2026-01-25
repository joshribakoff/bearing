"""Tests for plans loading and frontmatter parsing."""
import json
import pytest
from pathlib import Path

from bearing_tui.widgets.plans import load_plans, parse_plan_frontmatter


# =============================================================================
# FIXTURES
# =============================================================================

@pytest.fixture
def workspace_with_plans(tmp_path):
    """Create workspace with plans directory."""
    # Create JSONL files
    with open(tmp_path / "local.jsonl", "w") as f:
        f.write('{"folder": "test", "repo": "test", "branch": "main", "base": true}\n')
    with open(tmp_path / "workflow.jsonl", "w") as f:
        pass
    with open(tmp_path / "health.jsonl", "w") as f:
        pass

    # Create plans directory
    plans_dir = tmp_path / "plans" / "bearing"
    plans_dir.mkdir(parents=True)

    # Create plan files with frontmatter
    plan1 = plans_dir / "001-tui-tests.md"
    plan1.write_text("""---
title: TUI Test Foundation
status: active
issue: 17
---

# TUI Test Foundation

This plan covers testing for the TUI.
""")

    plan2 = plans_dir / "002-daemon.md"
    plan2.write_text("""---
title: Daemon Health Monitoring
status: in_progress
issue: 14
---

# Daemon Health Monitoring

Monitoring worktree health.
""")

    plan3 = plans_dir / "003-no-issue.md"
    plan3.write_text("""---
title: Future Feature
status: draft
---

# Future Feature

No issue linked yet.
""")

    # Create another project's plans
    project2_dir = tmp_path / "plans" / "sailkit"
    project2_dir.mkdir(parents=True)

    plan4 = project2_dir / "001-compass.md"
    plan4.write_text("""---
title: Compass Refactor
status: completed
issue: 42
---

# Compass Refactor
""")

    return tmp_path


@pytest.fixture
def workspace_no_plans(tmp_path):
    """Create workspace without plans directory."""
    with open(tmp_path / "local.jsonl", "w") as f:
        f.write('{"folder": "test", "repo": "test", "branch": "main", "base": true}\n')
    with open(tmp_path / "workflow.jsonl", "w") as f:
        pass
    with open(tmp_path / "health.jsonl", "w") as f:
        pass
    return tmp_path


@pytest.fixture
def workspace_empty_plans(tmp_path):
    """Create workspace with empty plans directory."""
    with open(tmp_path / "local.jsonl", "w") as f:
        f.write('{"folder": "test", "repo": "test", "branch": "main", "base": true}\n')
    with open(tmp_path / "workflow.jsonl", "w") as f:
        pass
    with open(tmp_path / "health.jsonl", "w") as f:
        pass
    (tmp_path / "plans").mkdir()
    return tmp_path


# =============================================================================
# PLAN LOADING TESTS (Modal tests removed - fullscreen view now used)
# =============================================================================

def test_load_plans_from_workspace(workspace_with_plans):
    """load_plans returns all plans from plans directory."""
    plans = load_plans(workspace_with_plans)

    # Should have 4 plans total (3 bearing + 1 sailkit)
    assert len(plans) == 4

    # Check bearing plans
    bearing_plans = [p for p in plans if p.project == "bearing"]
    assert len(bearing_plans) == 3

    # Check sailkit plans
    sailkit_plans = [p for p in plans if p.project == "sailkit"]
    assert len(sailkit_plans) == 1


def test_load_plans_parses_frontmatter(workspace_with_plans):
    """Plans should have parsed frontmatter fields."""
    plans = load_plans(workspace_with_plans)

    # Find the active plan
    active_plan = next((p for p in plans if p.status == "active"), None)
    assert active_plan is not None
    assert active_plan.title == "TUI Test Foundation"
    assert active_plan.issue == "17"


def test_load_plans_handles_missing_issue(workspace_with_plans):
    """Plans without issue field should have None."""
    plans = load_plans(workspace_with_plans)

    draft_plan = next((p for p in plans if p.title == "Future Feature"), None)
    assert draft_plan is not None
    assert draft_plan.issue is None


def test_load_plans_empty_workspace(workspace_no_plans):
    """load_plans returns empty list for workspace without plans."""
    plans = load_plans(workspace_no_plans)
    assert plans == []


def test_load_plans_sorts_by_project_and_status(workspace_with_plans):
    """Plans should be sorted by project, then status, then title."""
    plans = load_plans(workspace_with_plans)

    # First should be bearing plans (sorted by status)
    # active < in_progress < draft < completed
    bearing_plans = [p for p in plans if p.project == "bearing"]

    # Check status order
    status_order = {"active": 0, "in_progress": 1, "draft": 2, "completed": 3}
    for i in range(len(bearing_plans) - 1):
        current_status = status_order.get(bearing_plans[i].status, 4)
        next_status = status_order.get(bearing_plans[i + 1].status, 4)
        assert current_status <= next_status


# =============================================================================
# FRONTMATTER PARSING TESTS
# =============================================================================

def test_parse_frontmatter_basic(tmp_path):
    """Parse basic YAML frontmatter."""
    md_file = tmp_path / "test.md"
    md_file.write_text("""---
title: Test Plan
status: active
issue: 123
---

# Test Plan

Content here.
""")

    fm = parse_plan_frontmatter(md_file)
    assert fm["title"] == "Test Plan"
    assert fm["status"] == "active"
    assert fm["issue"] == "123"


def test_parse_frontmatter_quoted_values(tmp_path):
    """Parse frontmatter with quoted values."""
    md_file = tmp_path / "test.md"
    md_file.write_text("""---
title: "Quoted Title"
status: 'single quoted'
---

# Content
""")

    fm = parse_plan_frontmatter(md_file)
    assert fm["title"] == "Quoted Title"
    assert fm["status"] == "single quoted"


def test_parse_frontmatter_null_value(tmp_path):
    """Parse frontmatter with null value."""
    md_file = tmp_path / "test.md"
    md_file.write_text("""---
title: Test
issue: null
---

# Content
""")

    fm = parse_plan_frontmatter(md_file)
    assert fm["issue"] is None


def test_parse_frontmatter_no_frontmatter(tmp_path):
    """Extract title from first heading when no frontmatter."""
    md_file = tmp_path / "test.md"
    md_file.write_text("""# My Plan Title

Some content here.
""")

    fm = parse_plan_frontmatter(md_file)
    assert fm["title"] == "My Plan Title"


def test_parse_frontmatter_extracts_title_from_heading(tmp_path):
    """When no title in frontmatter, extract from first heading."""
    md_file = tmp_path / "test.md"
    md_file.write_text("""---
status: draft
---

# Extracted Title

Content.
""")

    fm = parse_plan_frontmatter(md_file)
    assert fm["title"] == "Extracted Title"
