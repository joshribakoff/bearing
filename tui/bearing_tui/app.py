"""Bearing TUI application."""
import os
from pathlib import Path

from textual.app import App, ComposeResult
from textual.containers import Horizontal, Vertical
from textual.widgets import Static, Footer
from textual.binding import Binding

from bearing_tui.state import BearingState
from bearing_tui.widgets import (
    ProjectList,
    WorktreeTable,
    WorktreeEntry,
    HealthEntry,
    DetailsPanel,
    LocalEntry,
    WorkflowEntry,
)


class BearingApp(App):
    """Bearing worktree management TUI."""

    CSS_PATH = "styles/app.tcss"
    TITLE = "Bearing"

    BINDINGS = [
        Binding("q", "quit", "Quit"),
        Binding("n", "new_worktree", "New"),
        Binding("c", "cleanup", "Cleanup"),
        Binding("r", "refresh", "Refresh"),
        Binding("d", "daemon", "Daemon"),
        Binding("j", "cursor_down", "Down", show=False),
        Binding("k", "cursor_up", "Up", show=False),
        Binding("h", "focus_left", "Left", show=False),
        Binding("l", "focus_right", "Right", show=False),
    ]

    def __init__(self, workspace: Path | None = None):
        super().__init__()
        # Default to BEARING_WORKSPACE env var, or parent of script directory
        if workspace is None:
            env_workspace = os.environ.get("BEARING_WORKSPACE")
            if env_workspace:
                workspace = Path(env_workspace)
            else:
                workspace = Path(__file__).parent.parent.parent.parent.parent
        self.state = BearingState(workspace)
        self._current_project: str | None = None

    def compose(self) -> ComposeResult:
        """Create the app layout."""
        yield Static("Bearing", id="title")
        with Horizontal(id="main-container"):
            with Vertical(id="projects-panel"):
                yield ProjectList(id="project-list")
            with Vertical(id="worktrees-panel"):
                yield WorktreeTable()
        yield DetailsPanel(id="details-panel")
        yield Static("[n]ew  [c]leanup  [r]efresh  [d]aemon  [q]uit", id="footer-bar")

    def on_mount(self) -> None:
        """Load data when app mounts."""
        self.action_refresh()

    def action_refresh(self) -> None:
        """Refresh data from files."""
        projects = self.state.get_projects()
        project_list = self.query_one(ProjectList)
        project_list.set_projects(projects)

        # Clear worktree table and details
        worktree_table = self.query_one(WorktreeTable)
        worktree_table.clear_worktrees()
        details = self.query_one(DetailsPanel)
        details.clear()
        self._current_project = None

        self.notify("Data refreshed")

    def on_project_list_project_selected(self, event: ProjectList.ProjectSelected) -> None:
        """Handle project selection."""
        self._current_project = event.project
        self._update_worktree_table(event.project)

    def _update_worktree_table(self, project: str) -> None:
        """Update worktree table for selected project."""
        worktrees = self.state.get_worktrees_for_project(project)

        # Convert to WorktreeEntry format
        wt_entries = [
            WorktreeEntry(
                folder=w.folder,
                repo=w.repo,
                branch=w.branch,
                base=w.base,
            )
            for w in worktrees
        ]

        # Build health map
        health_map = {}
        for w in worktrees:
            health = self.state.get_health_for_folder(w.folder)
            if health:
                health_map[w.folder] = HealthEntry(
                    folder=health.folder,
                    dirty=health.dirty,
                    unpushed=health.unpushed,
                    pr_state=health.pr_state,
                )

        worktree_table = self.query_one(WorktreeTable)
        worktree_table.set_worktrees(wt_entries, health_map)

    def on_worktree_table_worktree_selected(self, event: WorktreeTable.WorktreeSelected) -> None:
        """Handle worktree selection."""
        self._update_details(event.folder)

    def _update_details(self, folder: str) -> None:
        """Update details panel for selected worktree."""
        # Find the local entry
        local_entries = self.state.read_local()
        local_entry = None
        for e in local_entries:
            if e.folder == folder:
                local_entry = LocalEntry(
                    folder=e.folder,
                    repo=e.repo,
                    branch=e.branch,
                    base=e.base,
                )
                break

        if not local_entry:
            return

        # Get workflow entry
        workflow = self.state.get_workflow_for_branch(local_entry.repo, local_entry.branch)
        workflow_entry = None
        if workflow:
            workflow_entry = WorkflowEntry(
                repo=workflow.repo,
                branch=workflow.branch,
                based_on=workflow.based_on,
                purpose=workflow.purpose,
                status=workflow.status,
                created=str(workflow.created) if workflow.created else None,
            )

        # Get health entry
        health = self.state.get_health_for_folder(folder)
        health_entry = None
        if health:
            from bearing_tui.widgets.details import HealthEntry as DetailsHealthEntry
            health_entry = DetailsHealthEntry(
                folder=health.folder,
                dirty=health.dirty,
                unpushed=health.unpushed,
                pr_state=health.pr_state,
            )

        details = self.query_one(DetailsPanel)
        details.set_worktree(local_entry, workflow_entry, health_entry)

    def action_cursor_down(self) -> None:
        """Move cursor down in focused widget."""
        focused = self.focused
        if isinstance(focused, ProjectList):
            focused.action_cursor_down()
        elif isinstance(focused, WorktreeTable):
            focused.action_cursor_down()

    def action_cursor_up(self) -> None:
        """Move cursor up in focused widget."""
        focused = self.focused
        if isinstance(focused, ProjectList):
            focused.action_cursor_up()
        elif isinstance(focused, WorktreeTable):
            focused.action_cursor_up()

    def action_focus_left(self) -> None:
        """Focus the project list."""
        self.query_one(ProjectList).focus()

    def action_focus_right(self) -> None:
        """Focus the worktree table."""
        self.query_one(WorktreeTable).focus()

    def action_new_worktree(self) -> None:
        """Create a new worktree (placeholder)."""
        self.notify("New worktree: not yet implemented")

    def action_cleanup(self) -> None:
        """Cleanup a worktree (placeholder)."""
        self.notify("Cleanup: not yet implemented")

    def action_daemon(self) -> None:
        """Toggle daemon status (placeholder)."""
        self.notify("Daemon: not yet implemented")


def main():
    """Run the Bearing TUI."""
    app = BearingApp()
    app.run()


if __name__ == "__main__":
    main()
