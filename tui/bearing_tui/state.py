"""Bearing state file reader for JSONL files."""

import json
import os
from pathlib import Path
from dataclasses import dataclass
from typing import Optional
from datetime import datetime


def find_workspace_root(start_path: Path | None = None) -> Path | None:
    """Walk up directory tree to find workspace root (contains local.jsonl or workflow.jsonl)."""
    if start_path is None:
        start_path = Path.cwd()

    current = start_path.resolve()

    # Walk up until we find state files or hit root
    while current != current.parent:
        if (current / "local.jsonl").exists() or (current / "workflow.jsonl").exists():
            return current
        current = current.parent

    # Check root as well
    if (current / "local.jsonl").exists() or (current / "workflow.jsonl").exists():
        return current

    return None


@dataclass
class LocalEntry:
    folder: str
    repo: str
    branch: str
    base: bool


@dataclass
class WorkflowEntry:
    repo: str
    branch: str
    based_on: Optional[str]
    purpose: Optional[str]
    status: str
    created: Optional[datetime]


@dataclass
class HealthEntry:
    folder: str
    dirty: bool
    unpushed: int
    pr_state: Optional[str]
    last_check: Optional[datetime]


def _parse_datetime(value: Optional[str]) -> Optional[datetime]:
    """Parse ISO format datetime, handling 'unknown' and None."""
    if not value or value == "unknown":
        return None
    try:
        # Handle timezone-aware ISO format (e.g., 2026-01-19T21:33:21.035565-08:00)
        return datetime.fromisoformat(value)
    except ValueError:
        return None


def _read_jsonl(path: Path) -> list[dict]:
    """Read JSONL file, returning empty list if missing."""
    if not path.exists():
        return []
    entries = []
    with open(path, "r") as f:
        for line in f:
            line = line.strip()
            if line:
                entries.append(json.loads(line))
    return entries


class BearingState:
    def __init__(self, workspace_dir: Path):
        self.workspace_dir = workspace_dir

    def read_local(self) -> list[LocalEntry]:
        """Read local.jsonl - folder mappings."""
        entries = _read_jsonl(self.workspace_dir / "local.jsonl")
        return [
            LocalEntry(
                folder=e["folder"],
                repo=e["repo"],
                branch=e["branch"],
                base=e.get("base", False),
            )
            for e in entries
        ]

    def read_workflow(self) -> list[WorkflowEntry]:
        """Read workflow.jsonl - branch metadata."""
        entries = _read_jsonl(self.workspace_dir / "workflow.jsonl")
        return [
            WorkflowEntry(
                repo=e["repo"],
                branch=e["branch"],
                based_on=e.get("basedOn") if e.get("basedOn") != "unknown" else None,
                purpose=e.get("purpose") or None,
                status=e.get("status", "unknown"),
                created=_parse_datetime(e.get("created")),
            )
            for e in entries
        ]

    def read_health(self) -> list[HealthEntry]:
        """Read health.jsonl - cached health status."""
        entries = _read_jsonl(self.workspace_dir / "health.jsonl")
        return [
            HealthEntry(
                folder=e["folder"],
                dirty=e.get("dirty", False),
                unpushed=e.get("unpushed", 0),
                pr_state=e.get("prState"),
                last_check=_parse_datetime(e.get("lastCheck")),
            )
            for e in entries
        ]

    def get_projects(self) -> list[str]:
        """Get unique repo names from local entries."""
        repos = set()
        for entry in self.read_local():
            repos.add(entry.repo)
        return sorted(repos)

    def get_worktrees_for_project(self, repo: str) -> list[LocalEntry]:
        """Get worktrees filtered by repo."""
        return [e for e in self.read_local() if e.repo == repo]

    def get_health_for_folder(self, folder: str) -> Optional[HealthEntry]:
        """Get health entry for a specific folder."""
        for entry in self.read_health():
            if entry.folder == folder:
                return entry
        return None

    def get_workflow_for_branch(self, repo: str, branch: str) -> Optional[WorkflowEntry]:
        """Get workflow entry for a repo/branch."""
        for entry in self.read_workflow():
            if entry.repo == repo and entry.branch == branch:
                return entry
        return None


if __name__ == "__main__":
    # Test by reading actual files
    workspace = Path("/Users/joshribakoff/Projects")
    state = BearingState(workspace)

    print("=== Projects ===")
    for p in state.get_projects():
        print(f"  {p}")

    print("\n=== Sample Local Entries ===")
    for e in state.read_local()[:5]:
        print(f"  {e.folder}: {e.repo}/{e.branch} (base={e.base})")

    print("\n=== Sample Workflow Entries ===")
    for e in state.read_workflow()[:5]:
        print(f"  {e.repo}/{e.branch}: {e.status}, created={e.created}")

    print("\n=== Sample Health Entries ===")
    for e in state.read_health()[:5]:
        print(f"  {e.folder}: dirty={e.dirty}, unpushed={e.unpushed}, pr={e.pr_state}")

    print("\n=== Worktrees for 'bearing' ===")
    for e in state.get_worktrees_for_project("bearing"):
        print(f"  {e.folder}: {e.branch}")

    print("\n=== Health for 'bearing' folder ===")
    h = state.get_health_for_folder("bearing")
    if h:
        print(f"  dirty={h.dirty}, unpushed={h.unpushed}, last_check={h.last_check}")

    print("\n=== Workflow for bearing/go-rewrite ===")
    w = state.get_workflow_for_branch("bearing", "go-rewrite")
    if w:
        print(f"  based_on={w.based_on}, purpose={w.purpose}, created={w.created}")
