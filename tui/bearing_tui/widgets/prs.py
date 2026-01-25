"""PR table widget for displaying GitHub PRs."""
from dataclasses import dataclass
from typing import NamedTuple

from textual.widgets import DataTable
from textual.message import Message


@dataclass
class PRDisplayEntry:
    """PR data for display in the table."""
    repo: str
    number: int
    title: str
    state: str
    branch: str
    checks: str | None
    worktree: str | None  # Linked worktree folder if branch matches


class PRsTable(DataTable):
    """Table showing GitHub PRs."""

    class PRSelected(Message):
        """Emitted when a PR row is selected."""
        def __init__(self, repo: str, number: int) -> None:
            self.repo = repo
            self.number = number
            super().__init__()

    def __init__(self, **kwargs) -> None:
        super().__init__(**kwargs)
        self.cursor_type = "row"
        self._prs: list[PRDisplayEntry] = []

    def _setup_columns(self) -> None:
        """Add table columns."""
        self.add_columns("#", "Title", "State", "Branch", "Checks", "Worktree")

    def on_mount(self) -> None:
        """Set up table when mounted."""
        self._setup_columns()

    def set_prs(self, prs: list[PRDisplayEntry]) -> None:
        """Update table with PR data."""
        self._prs = prs
        self.clear()

        if not prs:
            self.add_row("No PRs", "", "", "", "", "", key="empty")
            return

        for pr in prs:
            # Format state with color indicators
            state_display = {
                "OPEN": "[green]OPEN[/]",
                "DRAFT": "[dim]DRAFT[/]",
                "MERGED": "[magenta]MERGED[/]",
                "CLOSED": "[red]CLOSED[/]",
            }.get(pr.state, pr.state)

            # Format checks
            checks_display = {
                "SUCCESS": "[green]\u2713[/]",
                "FAILURE": "[red]\u2717[/]",
                "PENDING": "[yellow]\u25cf[/]",
            }.get(pr.checks or "", "-")

            # Truncate title
            title = (pr.title[:35] + "...") if len(pr.title) > 38 else pr.title

            # Worktree indicator
            wt = pr.worktree or "-"

            self.add_row(
                str(pr.number),
                title,
                state_display,
                pr.branch[:20],
                checks_display,
                wt,
                key=f"{pr.repo}:{pr.number}"
            )

    def on_data_table_row_selected(self, event: DataTable.RowSelected) -> None:
        """Handle row selection."""
        if event.row_key and str(event.row_key.value) != "empty":
            key = str(event.row_key.value)
            if ":" in key:
                repo, num = key.rsplit(":", 1)
                self.post_message(self.PRSelected(repo, int(num)))

    def clear_prs(self) -> None:
        """Clear the table."""
        self.clear()
        self._prs = []
        self.add_row("Press 'r' to view PRs", "", "", "", "", "", key="empty")

    def get_selected_pr(self) -> PRDisplayEntry | None:
        """Get the currently selected PR."""
        if not self._prs or self.row_count == 0:
            return None
        try:
            from textual.coordinate import Coordinate
            cell_key = self.coordinate_to_cell_key(Coordinate(self.cursor_row, 0))
            key = str(cell_key.row_key.value)
            if key == "empty":
                return None
            for pr in self._prs:
                if f"{pr.repo}:{pr.number}" == key:
                    return pr
        except Exception:
            pass
        return None
