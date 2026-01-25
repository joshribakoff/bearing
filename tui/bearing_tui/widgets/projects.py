"""Project list widget for the left panel."""
import re
from textual.widgets import ListView, ListItem, Static
from textual.message import Message
from rich.text import Text


class ProjectListItem(ListItem):
    """ListItem that stores the original project name and count."""

    def __init__(self, project: str, count: int = 0, **kwargs) -> None:
        super().__init__(**kwargs)
        self.project = project
        self.count = count

    def compose(self):
        text = Text()
        text.append(self.project)
        if self.count > 0:
            text.append(f" ({self.count})", style="dim cyan")
        yield Static(text, markup=False)

    def on_click(self, event) -> None:
        """Focus parent list when item is clicked."""
        if self.parent:
            self.parent.focus()


class ProjectList(ListView):
    """Left panel showing list of repos/projects."""

    class ProjectSelected(Message):
        """Emitted when a project is selected."""
        def __init__(self, project: str) -> None:
            self.project = project
            super().__init__()

    def __init__(self, projects: list[str] | None = None, **kwargs) -> None:
        super().__init__(**kwargs)
        self.projects = projects or []
        self._counts: dict[str, int] = {}

    def compose(self):
        return
        yield

    def on_list_view_selected(self, event: ListView.Selected) -> None:
        """Handle selection and emit ProjectSelected message."""
        if isinstance(event.item, ProjectListItem):
            self.post_message(self.ProjectSelected(event.item.project))

    def on_click(self, event) -> None:
        """Focus the list when clicked."""
        self.focus()

    def set_projects(self, projects: list[str], counts: dict[str, int] | None = None, preserve_selection: str | None = None) -> None:
        """Update the project list with optional worktree counts.

        Args:
            projects: List of project names
            counts: Optional dict of project -> worktree count
            preserve_selection: Project name to re-select after update
        """
        self.projects = projects
        self._counts = counts or {}
        self.clear()
        if not projects:
            self.append(ListItem(Static("No projects found")))
        else:
            for i, project in enumerate(projects):
                count = self._counts.get(project, 0)
                self.append(ProjectListItem(project, count))
                # Restore selection if this is the preserved project
                if preserve_selection and project == preserve_selection:
                    self.index = i
