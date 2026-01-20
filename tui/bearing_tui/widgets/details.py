"""Details panel widget for the bottom panel."""
from dataclasses import dataclass
from typing import Any

from textual.widgets import Static
from rich.text import Text


@dataclass
class LocalEntry:
    """Represents a worktree from local.jsonl."""
    folder: str
    repo: str
    branch: str
    base: bool = False


@dataclass
class WorkflowEntry:
    """Represents workflow metadata from workflow.jsonl."""
    repo: str
    branch: str
    based_on: str | None = None
    purpose: str | None = None
    status: str | None = None
    created: str | None = None


@dataclass
class HealthEntry:
    """Represents health status for a worktree."""
    folder: str
    dirty: bool = False
    unpushed: int = 0
    pr_state: str | None = None
    pr_url: str | None = None


class DetailsPanel(Static, can_focus=True):
    """Bottom panel showing details for selected worktree."""

    def __init__(self, **kwargs) -> None:
        super().__init__(**kwargs)
        self.current_folder: str | None = None

    def on_mount(self) -> None:
        """Show empty state on mount."""
        self._show_empty()

    def _show_empty(self) -> None:
        """Display empty state."""
        text = Text("Select a worktree to view details", style="dim italic")
        self.update(text)

    def set_worktree(
        self,
        local_entry: LocalEntry | None,
        workflow_entry: WorkflowEntry | None = None,
        health_entry: HealthEntry | None = None
    ) -> None:
        """Update details for selected worktree."""
        if not local_entry:
            self._show_empty()
            return

        self.current_folder = local_entry.folder
        text = Text()

        # Folder and branch
        text.append("Folder: ", style="bold blue")
        text.append(f"{local_entry.folder}\n", style="bright_white")
        text.append("Branch: ", style="bold blue")
        text.append(f"{local_entry.branch}", style="bright_white")
        if local_entry.base:
            text.append("  ")
            text.append("\u2605 BASE", style="bold yellow")
        text.append("\n")

        # Workflow metadata
        if workflow_entry:
            if workflow_entry.purpose:
                text.append("Purpose: ", style="bold magenta")
                text.append(f"{workflow_entry.purpose}\n", style="bright_white")
            if workflow_entry.based_on:
                text.append("Based on: ", style="bold blue")
                text.append(f"{workflow_entry.based_on}\n", style="dim")
            if workflow_entry.status:
                status_style = "bold green" if workflow_entry.status == "completed" else "bold cyan"
                text.append("Status: ", style="bold blue")
                text.append(f"{workflow_entry.status}\n", style=status_style)

        # Health status
        if health_entry:
            text.append("Health: ", style="bold cyan")
            parts = []
            if health_entry.dirty:
                parts.append(Text("\u25cf dirty", style="bold yellow"))
            if health_entry.unpushed:
                parts.append(Text(f"\u2191 {health_entry.unpushed} unpushed", style="bold bright_cyan"))
            if health_entry.pr_state:
                pr_style = {
                    "open": "bold green",
                    "merged": "bold magenta",
                    "closed": "bold red"
                }.get(health_entry.pr_state.lower(), "white")
                parts.append(Text(f"PR: {health_entry.pr_state.upper()}", style=pr_style))
            if parts:
                for i, part in enumerate(parts):
                    if i > 0:
                        text.append(" \u2022 ", style="dim")
                    text.append_text(part)
            else:
                text.append("\u2714 clean", style="bold green")

        self.update(text)

    def clear(self) -> None:
        """Clear the details panel."""
        self.current_folder = None
        self._show_empty()
