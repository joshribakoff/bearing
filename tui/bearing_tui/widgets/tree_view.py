"""Tree view widget showing projects and worktrees in a hierarchical structure."""
from dataclasses import dataclass
from typing import Any

from textual.widgets import Tree
from textual.widgets.tree import TreeNode
from textual.message import Message
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


@dataclass
class HealthEntry:
    """Represents health status for a worktree."""
    folder: str
    dirty: bool = False
    unpushed: int = 0
    pr_state: str | None = None


@dataclass
class WorktreeData:
    """Combined data for a worktree node."""
    folder: str
    branch: str
    base: bool = False
    dirty: bool = False
    pr_state: str | None = None
    purpose: str | None = None


class WorktreeTreeView(Tree[WorktreeData | str]):
    """Tree view showing projects with worktree children.

    Structure:
        ▼ bearing
          ├── main (base)
          ├── go-rewrite          ●  OPEN
          └── tui-improvements       DRAFT
        ▼ sailkit
          ├── main (base)
          └── atlas-refactor      ●
        ▶ portfolio (no worktrees)
    """

    class TreeNodeSelected(Message):
        """Emitted when a tree node is selected."""
        def __init__(self, node_type: str, value: str, data: WorktreeData | None = None) -> None:
            self.node_type = node_type  # "project" or "worktree"
            self.value = value  # project name or folder path
            self.data = data  # WorktreeData if worktree
            super().__init__()

    def __init__(self, **kwargs) -> None:
        super().__init__("Worktrees", **kwargs)
        self.show_root = False
        self._local_entries: list[LocalEntry] = []
        self._workflow_map: dict[tuple[str, str], WorkflowEntry] = {}
        self._health_map: dict[str, HealthEntry] = {}

    def set_data(
        self,
        local_entries: list[LocalEntry],
        workflow_entries: list[WorkflowEntry] | None = None,
        health_entries: list[HealthEntry] | None = None,
    ) -> None:
        """Update tree with data from bearing state files.

        Args:
            local_entries: Worktree folder mappings from local.jsonl
            workflow_entries: Branch metadata from workflow.jsonl
            health_entries: Health status from health.jsonl
        """
        self._local_entries = local_entries
        self._workflow_map = {}
        self._health_map = {}

        if workflow_entries:
            for w in workflow_entries:
                self._workflow_map[(w.repo, w.branch)] = w

        if health_entries:
            for h in health_entries:
                self._health_map[h.folder] = h

        self._rebuild_tree()

    def _rebuild_tree(self) -> None:
        """Rebuild the tree structure from current data."""
        self.clear()

        # Group worktrees by project (repo)
        projects: dict[str, list[LocalEntry]] = {}
        for entry in self._local_entries:
            if entry.repo not in projects:
                projects[entry.repo] = []
            projects[entry.repo].append(entry)

        # Sort projects alphabetically
        for project_name in sorted(projects.keys()):
            worktrees = projects[project_name]
            project_label = self._format_project_label(project_name, worktrees)
            project_node = self.root.add(project_label, data=project_name)

            # Sort worktrees: base first, then by branch name
            sorted_wts = sorted(worktrees, key=lambda w: (not w.base, w.branch))

            for wt in sorted_wts:
                wt_data = self._create_worktree_data(wt)
                wt_label = self._format_worktree_label(wt_data)
                project_node.add_leaf(wt_label, data=wt_data)

            # Expand projects with worktrees by default
            if worktrees:
                project_node.expand()

    def _format_project_label(self, project_name: str, worktrees: list[LocalEntry]) -> Text:
        """Format the label for a project node."""
        text = Text()
        text.append(project_name, style="bold cyan")
        if not worktrees:
            text.append(" (no worktrees)", style="dim")
        return text

    def _format_worktree_label(self, data: WorktreeData) -> Text:
        """Format the label for a worktree node.

        Format: branch_name (base)  ●  PR_STATE
        """
        text = Text()

        # Branch name
        if data.base:
            text.append(data.branch, style="yellow")
            text.append(" (base)", style="dim yellow")
        else:
            text.append(data.branch, style="white")

        # Dirty indicator and PR state - right-aligned appearance via padding
        indicators = []

        if data.dirty:
            indicators.append(("●", "bold yellow"))

        if data.pr_state:
            pr_style = {
                "open": "bold green",
                "merged": "bold magenta",
                "closed": "bold red",
                "draft": "dim cyan",
            }.get(data.pr_state.lower(), "white")
            indicators.append((data.pr_state.upper(), pr_style))

        if indicators:
            # Add spacing before indicators
            text.append("  ", style="")
            for i, (indicator, style) in enumerate(indicators):
                if i > 0:
                    text.append("  ", style="")
                text.append(indicator, style=style)

        return text

    def _create_worktree_data(self, entry: LocalEntry) -> WorktreeData:
        """Create WorktreeData by combining local, workflow, and health data."""
        health = self._health_map.get(entry.folder)
        workflow = self._workflow_map.get((entry.repo, entry.branch))

        return WorktreeData(
            folder=entry.folder,
            branch=entry.branch,
            base=entry.base,
            dirty=health.dirty if health else False,
            pr_state=health.pr_state if health else None,
            purpose=workflow.purpose if workflow else None,
        )

    def on_tree_node_selected(self, event: Tree.NodeSelected) -> None:
        """Handle node selection and emit TreeNodeSelected message."""
        node = event.node
        data = node.data

        if isinstance(data, WorktreeData):
            # Worktree node selected
            self.post_message(self.TreeNodeSelected(
                node_type="worktree",
                value=data.folder,
                data=data,
            ))
        elif isinstance(data, str):
            # Project node selected
            self.post_message(self.TreeNodeSelected(
                node_type="project",
                value=data,
                data=None,
            ))

    def expand_all(self) -> None:
        """Expand all project nodes."""
        for node in self.root.children:
            node.expand()

    def collapse_all(self) -> None:
        """Collapse all project nodes."""
        for node in self.root.children:
            node.collapse()

    def select_project(self, project_name: str) -> None:
        """Select and expand a specific project by name."""
        for node in self.root.children:
            if node.data == project_name:
                node.expand()
                self.select_node(node)
                break

    def select_worktree(self, folder: str) -> None:
        """Select a specific worktree by folder name."""
        for project_node in self.root.children:
            for wt_node in project_node.children:
                if isinstance(wt_node.data, WorktreeData) and wt_node.data.folder == folder:
                    project_node.expand()
                    self.select_node(wt_node)
                    return

    def refresh_health(self, health_entries: list[HealthEntry]) -> None:
        """Update health data and refresh labels without rebuilding entire tree."""
        self._health_map = {h.folder: h for h in health_entries}

        for project_node in self.root.children:
            for wt_node in project_node.children:
                if isinstance(wt_node.data, WorktreeData):
                    # Find corresponding local entry
                    for entry in self._local_entries:
                        if entry.folder == wt_node.data.folder:
                            new_data = self._create_worktree_data(entry)
                            wt_node.data = new_data
                            wt_node.set_label(self._format_worktree_label(new_data))
                            break
