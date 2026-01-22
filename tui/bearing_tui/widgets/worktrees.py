"""Worktree table widget for the right panel."""
from dataclasses import dataclass
from typing import Any

from textual.widgets import DataTable
from textual.message import Message


@dataclass
class WorktreeEntry:
    """Represents a worktree from local.jsonl."""
    folder: str
    repo: str
    branch: str
    base: bool = False
    purpose: str | None = None


@dataclass
class HealthEntry:
    """Represents health status for a worktree."""
    folder: str
    dirty: bool = False
    unpushed: int = 0
    pr_state: str | None = None


class WorktreeTable(DataTable):
    """Right panel showing worktrees for selected project."""

    class WorktreeSelected(Message):
        """Emitted when a worktree row is selected."""
        def __init__(self, folder: str) -> None:
            self.folder = folder
            super().__init__()

    def __init__(self, **kwargs) -> None:
        super().__init__(**kwargs)
        self.cursor_type = "row"

    def _setup_columns(self) -> None:
        """Add table columns."""
        self.add_columns("Branch", "Purpose", "Dirty", "Unpushed", "PR", "Base")

    def on_mount(self) -> None:
        """Set up table when mounted."""
        self._setup_columns()

    def set_worktrees(
        self,
        worktrees: list[WorktreeEntry],
        health_map: dict[str, HealthEntry] | None = None
    ) -> None:
        """Update table with worktrees and health data."""
        health_map = health_map or {}
        self.clear()

        if not worktrees:
            self.add_row("No worktrees", "", "", "", "", "", key="empty")
            return

        for wt in worktrees:
            health = health_map.get(wt.folder)
            purpose = (wt.purpose[:30] + "...") if wt.purpose and len(wt.purpose) > 30 else (wt.purpose or "-")
            dirty = "\u25cf" if health and health.dirty else ""
            unpushed = str(health.unpushed) if health and health.unpushed else "-"
            pr = health.pr_state if health and health.pr_state else "-"
            base = "\u2605" if wt.base else ""
            self.add_row(wt.branch, purpose, dirty, unpushed, pr, base, key=wt.folder)

    def on_data_table_row_selected(self, event: DataTable.RowSelected) -> None:
        """Handle row selection and emit WorktreeSelected message."""
        if event.row_key and event.row_key.value != "empty":
            self.post_message(self.WorktreeSelected(str(event.row_key.value)))

    def clear_worktrees(self) -> None:
        """Clear the table and show empty state."""
        self.clear()
        self.add_row("Select a project", "", "", "", "", "", key="empty")
