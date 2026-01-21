"""Worktree table widget for the right panel."""
from dataclasses import dataclass
from typing import Any

from rich.text import Text
from textual.widgets import DataTable
from textual.message import Message


# PR status styling: icon, color (DRAFT is gray like GitHub shows)
PR_STYLES = {
    "OPEN": ("●", "green"),
    "DRAFT": ("○", "dim"),
    "MERGED": ("✓", "dim cyan"),
    "CLOSED": ("✗", "dim red"),
}


@dataclass
class WorktreeEntry:
    """Represents a worktree from local.jsonl."""
    folder: str
    repo: str
    branch: str
    base: bool = False
    plan: str | None = None
    issue: str | None = None


@dataclass
class HealthEntry:
    """Represents health status for a worktree."""
    folder: str
    dirty: bool = False
    unpushed: int = 0
    pr_state: str | None = None
    pr_title: str | None = None


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
        self._hide_closed = False  # Toggle to hide merged/closed PRs
        self._all_entries: list[tuple] = []  # Store all entries for filtering

    def toggle_hide_closed(self) -> bool:
        """Toggle hiding of closed/merged PRs. Returns new state."""
        self._hide_closed = not self._hide_closed
        self._apply_filter()
        return self._hide_closed

    def _setup_columns(self) -> None:
        """Add table columns - ordered by actionability, branch last."""
        # PR status and title first (most actionable), then plan/issue, then branch last
        self.add_columns("PR", "Title", "Plan", "Issue", "Dirty", "Branch")

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

        if not worktrees:
            self.clear()
            self.add_row("-", "No worktrees", "-", "-", "", "", key="empty")
            self._all_entries = []
            return

        # Build all entries with styling
        self._all_entries = []
        for wt in worktrees:
            health = health_map.get(wt.folder)
            pr_state = health.pr_state if health and health.pr_state else None
            pr_title = health.pr_title if health and health.pr_title else None

            # Style PR status with icon and color
            if pr_state and pr_state in PR_STYLES:
                icon, color = PR_STYLES[pr_state]
                pr_display = Text(f"{icon} {pr_state}", style=color)
                is_closed = pr_state in ("MERGED", "CLOSED")
            else:
                pr_display = Text("-", style="dim")
                is_closed = False

            # Gray out entire row if closed/merged
            row_style = "dim" if is_closed else ""

            # PR title (truncate if too long)
            title_text = pr_title[:30] + "…" if pr_title and len(pr_title) > 30 else (pr_title or "-")
            title = Text(title_text, style=row_style)

            plan = Text(wt.plan or "-", style=row_style)
            issue = Text(f"#{wt.issue}" if wt.issue else "-", style=row_style)

            # Branch with star for base (replaces separate base column)
            branch_text = f"★ {wt.branch}" if wt.base else wt.branch
            branch = Text(branch_text, style="cyan" if wt.base else row_style)

            dirty = Text("●", style="yellow") if health and health.dirty else Text("")

            self._all_entries.append((pr_display, title, plan, issue, dirty, branch, wt.folder, is_closed))

        self._apply_filter()

    def _apply_filter(self) -> None:
        """Apply current filter settings to display."""
        self.clear()
        for pr, title, plan, issue, dirty, branch, folder, is_closed in self._all_entries:
            if self._hide_closed and is_closed:
                continue
            self.add_row(pr, title, plan, issue, dirty, branch, key=folder)

    def on_data_table_row_selected(self, event: DataTable.RowSelected) -> None:
        """Handle row selection and emit WorktreeSelected message."""
        if event.row_key and event.row_key.value != "empty":
            self.post_message(self.WorktreeSelected(str(event.row_key.value)))

    def clear_worktrees(self) -> None:
        """Clear the table and show empty state."""
        self.clear()
        self.add_row("-", "Select a project", "-", "-", "", "-", key="empty")
