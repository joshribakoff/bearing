"""Project list widget for the left panel."""
import re
from textual.widgets import ListView, ListItem, Label
from textual.message import Message


def _sanitize_id(name: str) -> str:
    """Sanitize a name for use as a Textual ID."""
    return re.sub(r"[^a-zA-Z0-9_-]", "_", name)


class ProjectListItem(ListItem):
    """ListItem that stores the original project name."""

    def __init__(self, project: str, **kwargs) -> None:
        super().__init__(**kwargs)
        self.project = project


class ProjectList(ListView):
    """Left panel showing list of repos/projects."""

    DEFAULT_CSS = """
    ProjectList {
        width: 30;
        border: solid $primary;
    }
    ProjectList > ListItem {
        padding: 0 1;
    }
    ProjectList > ListItem.--highlight {
        background: $accent;
    }
    """

    class ProjectSelected(Message):
        """Emitted when a project is selected."""
        def __init__(self, project: str) -> None:
            self.project = project
            super().__init__()

    def __init__(self, projects: list[str] | None = None, **kwargs) -> None:
        super().__init__(**kwargs)
        self.projects = projects or []

    def compose(self):
        if not self.projects:
            yield ListItem(Label("No projects found"), id="empty-state")
        else:
            for project in self.projects:
                item = ProjectListItem(project, id=f"project-{_sanitize_id(project)}")
                item.compose_add_child(Label(project))
                yield item

    def on_list_view_selected(self, event: ListView.Selected) -> None:
        """Handle selection and emit ProjectSelected message."""
        if isinstance(event.item, ProjectListItem):
            self.post_message(self.ProjectSelected(event.item.project))

    def set_projects(self, projects: list[str]) -> None:
        """Update the project list."""
        self.projects = projects
        self.clear()
        if not projects:
            self.append(ListItem(Label("No projects found"), id="empty-state"))
        else:
            for project in projects:
                item = ProjectListItem(project, id=f"project-{_sanitize_id(project)}")
                item.compose_add_child(Label(project))
                self.append(item)
