"""Plans list widget for displaying plans from ~/Projects/plans/."""

from pathlib import Path
from typing import NamedTuple

from textual.widgets import ListView, ListItem, Static
from textual.message import Message


class PlanEntry(NamedTuple):
    """Represents a plan entry."""
    file_path: Path
    project: str
    title: str
    issue: str | None
    status: str


class PlanListItem(ListItem):
    """A list item representing a plan."""

    def __init__(self, plan: PlanEntry) -> None:
        super().__init__()
        self.plan = plan

    def compose(self):
        status_indicator = {
            "active": "[green]●[/]",
            "in_progress": "[yellow]●[/]",
            "draft": "[dim]●[/]",
            "completed": "[blue]●[/]",
        }.get(self.plan.status, "[dim]○[/]")

        issue_str = f"[cyan]#{self.plan.issue}[/]" if self.plan.issue else "[dim]no issue[/]"

        yield Static(
            f"{status_indicator} {self.plan.title[:40]:<40} "
            f"[dim]{self.plan.project}[/] {issue_str}"
        )


class PlansList(ListView):
    """ListView showing plans from the plans directory."""

    class PlanSelected(Message):
        """Posted when a plan is selected."""
        def __init__(self, plan: PlanEntry) -> None:
            self.plan = plan
            super().__init__()

    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        self._plans: list[PlanEntry] = []

    def set_plans(self, plans: list[PlanEntry]) -> None:
        """Set the plans to display."""
        self._plans = plans
        self.clear()
        for plan in plans:
            self.append(PlanListItem(plan))

    def on_list_view_selected(self, event: ListView.Selected) -> None:
        """Handle plan selection."""
        item = event.item
        if isinstance(item, PlanListItem):
            self.post_message(self.PlanSelected(item.plan))


def parse_plan_frontmatter(file_path: Path) -> dict:
    """Parse YAML frontmatter from a plan file."""
    content = file_path.read_text()
    lines = content.split("\n")

    frontmatter = {}
    in_frontmatter = False

    for i, line in enumerate(lines):
        if line.strip() == "---":
            if not in_frontmatter and i == 0:
                in_frontmatter = True
                continue
            elif in_frontmatter:
                break
        elif in_frontmatter:
            if ":" in line:
                key, value = line.split(":", 1)
                key = key.strip()
                value = value.strip()
                # Remove quotes
                if value.startswith('"') and value.endswith('"'):
                    value = value[1:-1]
                elif value.startswith("'") and value.endswith("'"):
                    value = value[1:-1]
                # Handle null
                if value == "null":
                    value = None
                frontmatter[key] = value
        else:
            # No frontmatter, try to get title from first heading
            break

    # If no title in frontmatter, extract from first heading
    if "title" not in frontmatter:
        for line in lines:
            if line.startswith("# "):
                frontmatter["title"] = line[2:].strip()
                break

    return frontmatter


def load_plans(workspace_dir: Path) -> list[PlanEntry]:
    """Load plans from the plans directory."""
    plans_dir = workspace_dir / "plans"
    if not plans_dir.exists():
        return []

    plans = []
    for project_dir in plans_dir.iterdir():
        if not project_dir.is_dir():
            continue

        project = project_dir.name
        for plan_file in project_dir.glob("*.md"):
            try:
                fm = parse_plan_frontmatter(plan_file)
                plans.append(PlanEntry(
                    file_path=plan_file,
                    project=project,
                    title=fm.get("title", plan_file.stem),
                    issue=fm.get("issue"),
                    status=fm.get("status", "draft"),
                ))
            except Exception:
                continue

    # Sort by project, then status (active first), then title
    status_order = {"active": 0, "in_progress": 1, "draft": 2, "completed": 3}
    plans.sort(key=lambda p: (p.project, status_order.get(p.status, 4), p.title))

    return plans
