"""Plans table widget for the plans view."""
from dataclasses import dataclass
from pathlib import Path
from typing import NamedTuple

from textual.widgets import DataTable
from textual.message import Message


class PlanEntry(NamedTuple):
    """Represents a plan entry."""
    file_path: Path
    project: str
    title: str
    issue: str | None
    status: str
    pr: str | None = None


class PlansTable(DataTable):
    """Table showing plans for selected project."""

    class PlanSelected(Message):
        """Emitted when a plan row is selected."""
        def __init__(self, plan: PlanEntry) -> None:
            self.plan = plan
            super().__init__()

    def __init__(self, **kwargs) -> None:
        super().__init__(**kwargs)
        self.cursor_type = "row"
        self._plans: list[PlanEntry] = []

    def _setup_columns(self) -> None:
        """Add table columns."""
        self.add_columns("Title", "Status", "PR", "Issue")

    def on_mount(self) -> None:
        """Set up table when mounted."""
        self._setup_columns()

    def set_plans(self, plans: list[PlanEntry]) -> None:
        """Update table with plans."""
        self._plans = plans
        self.clear()

        if not plans:
            self.add_row("No plans", "", "", "", key="empty")
            return

        for plan in plans:
            status_display = self._format_status(plan.status)
            pr_display = plan.pr if plan.pr else "-"
            issue_display = f"#{plan.issue}" if plan.issue else "-"
            title = (plan.title[:35] + "...") if len(plan.title) > 35 else plan.title
            self.add_row(title, status_display, pr_display, issue_display, key=str(plan.file_path))

    def _format_status(self, status: str) -> str:
        """Format status with color indicator."""
        indicators = {
            "active": "[green]\u25cf[/] active",
            "in_progress": "[yellow]\u25cf[/] in_progress",
            "draft": "[dim]\u25cf[/] draft",
            "completed": "[blue]\u25cf[/] completed",
        }
        return indicators.get(status, f"[dim]\u25cb[/] {status}")

    def on_data_table_row_selected(self, event: DataTable.RowSelected) -> None:
        """Handle row selection and emit PlanSelected message."""
        if event.row_key and event.row_key.value != "empty":
            # Find the plan by file path
            for plan in self._plans:
                if str(plan.file_path) == event.row_key.value:
                    self.post_message(self.PlanSelected(plan))
                    break

    def clear_plans(self) -> None:
        """Clear the table and show empty state."""
        self._plans = []
        self.clear()
        self.add_row("Select a project", "", "", "", key="empty")

    def get_selected_plan(self) -> PlanEntry | None:
        """Get the currently selected plan."""
        if self.row_count == 0:
            return None
        try:
            from textual.coordinate import Coordinate
            cell_key = self.coordinate_to_cell_key(Coordinate(self.cursor_row, 0))
            key = str(cell_key.row_key.value)
            if key == "empty":
                return None
            for plan in self._plans:
                if str(plan.file_path) == key:
                    return plan
        except Exception:
            pass
        return None
