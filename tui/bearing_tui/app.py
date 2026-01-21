"""Bearing TUI application."""
import json
import os
import subprocess
import webbrowser
from datetime import datetime, timedelta
from pathlib import Path

from textual.app import App, ComposeResult
from textual.containers import Horizontal, Vertical, Container
from textual.widgets import Static, Footer, Label
from textual.binding import Binding
from textual.screen import ModalScreen

from bearing_tui.state import BearingState, find_workspace_root
from bearing_tui.widgets import (
    ProjectList,
    WorktreeTable,
    WorktreeEntry,
    HealthEntry,
    DetailsPanel,
    LocalEntry,
    WorkflowEntry,
    PlansList,
    load_plans,
)


class HelpScreen(ModalScreen):
    """Modal screen showing keybindings."""

    BINDINGS = [
        Binding("escape", "dismiss", "Close"),
        Binding("q", "dismiss", "Close"),
        Binding("question_mark", "dismiss", "Close"),
    ]

    def compose(self) -> ComposeResult:
        yield Container(
            Static(
                "[b cyan]Bearing TUI - Keybindings[/]\n\n"
                "[b]Navigation[/]\n"
                "  [yellow]0[/]      Focus projects panel\n"
                "  [yellow]1[/]      Focus worktrees panel\n"
                "  [yellow]2[/]      Focus details panel\n"
                "  [yellow]h / \u2190[/]  Focus left panel\n"
                "  [yellow]l / \u2192[/]  Focus right panel\n"
                "  [yellow]j / \u2193[/]  Move down\n"
                "  [yellow]k / \u2191[/]  Move up\n"
                "  [yellow]Tab[/]    Next panel\n"
                "  [yellow]Enter[/]  Select item\n\n"
                "[b]Actions[/]\n"
                "  [yellow]n[/]      New worktree\n"
                "  [yellow]c[/]      Cleanup worktree\n"
                "  [yellow]r[/]      Refresh data\n"
                "  [yellow]R[/]      Force refresh (daemon)\n"
                "  [yellow]d[/]      Daemon health check\n"
                "  [yellow]o[/]      Open PR in browser\n"
                "  [yellow]p[/]      View plans\n"
                "  [yellow]?[/]      Show this help\n"
                "  [yellow]q[/]      Quit\n",
                id="help-content",
            ),
            id="help-modal",
        )


class PlansScreen(ModalScreen):
    """Modal screen showing plans list."""

    BINDINGS = [
        Binding("escape", "dismiss", "Close"),
        Binding("q", "dismiss", "Close"),
        Binding("p", "dismiss", "Close"),
        Binding("j", "cursor_down", "Down", show=False),
        Binding("k", "cursor_up", "Up", show=False),
        Binding("o", "open_issue", "Open Issue", show=False),
    ]

    def __init__(self, workspace: Path) -> None:
        super().__init__()
        self.workspace = workspace

    def compose(self) -> ComposeResult:
        yield Container(
            Static("[b cyan]Plans[/] [dim](press o to open issue, Esc to close)[/]", id="plans-header"),
            PlansList(id="plans-list"),
            id="plans-modal",
        )

    def on_mount(self) -> None:
        plans = load_plans(self.workspace)
        plans_list = self.query_one("#plans-list", PlansList)
        plans_list.set_plans(plans)
        plans_list.focus()

    def action_cursor_down(self) -> None:
        self.query_one(PlansList).action_cursor_down()

    def action_cursor_up(self) -> None:
        self.query_one(PlansList).action_cursor_up()

    def action_open_issue(self) -> None:
        """Open the selected plan's issue in browser."""
        plans_list = self.query_one(PlansList)
        if plans_list.index is not None and plans_list.index < len(plans_list._plans):
            plan = plans_list._plans[plans_list.index]
            if plan.issue:
                # Get repo from the plan's project name
                url = f"https://github.com/joshribakoff/{plan.project}/issues/{plan.issue}"
                webbrowser.open(url)
                self.app.notify(f"Opened issue #{plan.issue}", timeout=2)


class BearingApp(App):
    """Bearing worktree management TUI."""

    CSS_PATH = "styles/app.tcss"
    TITLE = "Bearing"

    BINDINGS = [
        Binding("q", "quit", "Quit"),
        Binding("ctrl+c", "quit", "Quit", show=False),
        Binding("question_mark", "show_help", "Help"),
        Binding("n", "new_worktree", "New"),
        Binding("c", "cleanup", "Cleanup"),
        Binding("r", "refresh", "Refresh"),
        Binding("R", "force_refresh", "Force Refresh", show=False),
        Binding("d", "daemon", "Daemon"),
        Binding("o", "open_pr", "Open PR", show=False),
        Binding("p", "show_plans", "Plans"),
        # Panel navigation by number (0-indexed)
        Binding("0", "focus_panel_0", "Projects", show=False),
        Binding("1", "focus_panel_1", "Worktrees", show=False),
        Binding("2", "focus_panel_2", "Details", show=False),
        # Vim-style navigation
        Binding("j", "cursor_down", "Down", show=False),
        Binding("k", "cursor_up", "Up", show=False),
        Binding("h", "focus_left", "Left", show=False),
        Binding("l", "focus_right", "Right", show=False),
        Binding("tab", "focus_next_panel", "Next", show=False),
        Binding("shift+tab", "focus_prev_panel", "Prev", show=False),
    ]

    def __init__(self, workspace: Path | None = None):
        super().__init__()
        if workspace is None:
            # Try environment variable first
            env_workspace = os.environ.get("BEARING_WORKSPACE")
            if env_workspace:
                workspace = Path(env_workspace)
            else:
                # Walk up directory tree to find workspace root
                workspace = find_workspace_root()
                if workspace is None:
                    # Fallback to parent of bearing-tui
                    workspace = Path(__file__).parent.parent.parent.parent.parent
        self.workspace = workspace
        self.state = BearingState(workspace)
        self._current_project: str | None = None
        self._panel_order = ["project-list", "worktree-table", "details-panel"]

    def compose(self) -> ComposeResult:
        """Create the app layout."""
        yield Static("\u2693 Bearing", id="title")
        with Horizontal(id="main-container"):
            with Vertical(id="projects-panel"):
                yield Label("[0] Projects", classes="panel-header")
                yield ProjectList(id="project-list")
            with Vertical(id="worktrees-panel"):
                yield Label("[1] Worktrees", classes="panel-header")
                yield WorktreeTable(id="worktree-table")
        yield Label("[2] Details", classes="panel-header details-header")
        yield DetailsPanel(id="details-panel")
        yield Static(
            "[yellow]0[/]-[yellow]2[/] panels  "
            "[yellow]j/k[/] nav  "
            "[yellow]n[/]ew  "
            "[yellow]c[/]leanup  "
            "[yellow]r[/]efresh  "
            "[yellow]o[/]pen PR  "
            "[yellow]p[/]lans  "
            "[yellow]?[/] help  "
            "[yellow]q[/]uit",
            id="footer-bar"
        )

    @property
    def _session_file(self) -> Path:
        """Path to session state file."""
        bearing_dir = Path.home() / ".bearing"
        bearing_dir.mkdir(exist_ok=True)
        return bearing_dir / "tui-session.json"

    def _save_session(self) -> None:
        """Save full UI state to session file."""
        try:
            # Get focused panel
            focused = self.focused
            focused_panel = focused.id if focused and focused.id in self._panel_order else None

            # Get worktree table cursor position
            worktree_table = self.query_one(WorktreeTable)
            worktree_cursor = worktree_table.cursor_row if worktree_table.row_count > 0 else None

            # Get project list index
            project_list = self.query_one(ProjectList)
            project_index = project_list.index

            session = {
                "selected_project": self._current_project,
                "project_index": project_index,
                "worktree_cursor": worktree_cursor,
                "focused_panel": focused_panel,
                "timestamp": datetime.now().isoformat(),
            }
            self._session_file.write_text(json.dumps(session, indent=2))
        except Exception:
            pass  # Don't fail on session save errors

    def _restore_session(self) -> bool:
        """Restore full UI state from session file if recent. Returns True if restored."""
        try:
            if not self._session_file.exists():
                return False
            session = json.loads(self._session_file.read_text())
            timestamp = datetime.fromisoformat(session.get("timestamp", ""))
            # Only restore if < 24 hours old
            if datetime.now() - timestamp > timedelta(hours=24):
                return False

            # Restore project selection
            project = session.get("selected_project")
            if project:
                self._select_project(project)

            # Restore worktree cursor
            worktree_cursor = session.get("worktree_cursor")
            if worktree_cursor is not None:
                worktree_table = self.query_one(WorktreeTable)
                if worktree_table.row_count > worktree_cursor:
                    worktree_table.cursor_coordinate = (worktree_cursor, 0)
                    # Trigger details update for selected worktree
                    self._update_details_for_cursor(worktree_cursor)

            # Restore focused panel
            focused_panel = session.get("focused_panel")
            if focused_panel:
                self.query_one(f"#{focused_panel}").focus()
                return True
            return False
        except Exception:
            return False  # Don't fail on session restore errors

    def _update_details_for_cursor(self, cursor_row: int) -> None:
        """Update details panel for worktree at cursor position."""
        from textual.coordinate import Coordinate
        worktree_table = self.query_one(WorktreeTable)
        try:
            cell_key = worktree_table.coordinate_to_cell_key(Coordinate(cursor_row, 0))
            folder = str(cell_key.row_key.value)
            if folder and folder != "empty":
                self._update_details(folder)
        except Exception:
            pass

    def _select_project(self, project: str) -> None:
        """Select a project by name."""
        project_list = self.query_one("#project-list", ProjectList)
        for i, item in enumerate(project_list.children):
            if hasattr(item, "project") and item.project == project:
                project_list.index = i
                self._current_project = project
                self._update_worktree_table(project)
                break

    def _ensure_daemon_running(self) -> None:
        """Start daemon if not already running."""
        try:
            result = subprocess.run(
                ["bearing", "daemon", "status"],
                capture_output=True,
                text=True,
                timeout=2,
            )
            if "running" not in result.stdout.lower():
                # Start daemon in background
                subprocess.Popen(
                    ["bearing", "daemon", "start"],
                    stdout=subprocess.DEVNULL,
                    stderr=subprocess.DEVNULL,
                )
        except (subprocess.TimeoutExpired, FileNotFoundError):
            # bearing CLI not found or timeout - skip daemon
            pass

    def on_mount(self) -> None:
        """Load data when app mounts."""
        self._ensure_daemon_running()
        self.action_refresh()
        # Restore session (includes focus) or default to project list
        if not self._restore_session():
            self.query_one("#project-list", ProjectList).focus()

    def action_show_help(self) -> None:
        """Show the help modal."""
        self.push_screen(HelpScreen())

    def action_quit(self) -> None:
        """Save session and quit."""
        self._save_session()
        self.exit()

    def action_show_plans(self) -> None:
        """Show the plans modal."""
        self.push_screen(PlansScreen(self.workspace))

    def action_refresh(self) -> None:
        """Refresh data from files."""
        projects = self.state.get_projects()

        # Count worktrees per project
        local_entries = self.state.read_local()
        counts: dict[str, int] = {}
        for entry in local_entries:
            counts[entry.repo] = counts.get(entry.repo, 0) + 1

        project_list = self.query_one(ProjectList)
        project_list.set_projects(projects, counts)

        # Clear worktree table and details
        worktree_table = self.query_one(WorktreeTable)
        worktree_table.clear_worktrees()
        details = self.query_one(DetailsPanel)
        details.clear()
        self._current_project = None

        self.notify("Data refreshed", timeout=2)

    def action_focus_panel_0(self) -> None:
        """Focus the projects panel."""
        self.query_one("#project-list", ProjectList).focus()

    def action_focus_panel_1(self) -> None:
        """Focus the worktrees panel."""
        self.query_one("#worktree-table", WorktreeTable).focus()

    def action_focus_panel_2(self) -> None:
        """Focus the details panel."""
        self.query_one("#details-panel", DetailsPanel).focus()

    def action_focus_next_panel(self) -> None:
        """Focus the next panel in order."""
        current = self.focused
        if current is None:
            self.action_focus_panel_1()
            return

        current_id = current.id
        if current_id in self._panel_order:
            idx = self._panel_order.index(current_id)
            next_idx = (idx + 1) % len(self._panel_order)
            next_id = self._panel_order[next_idx]
            self.query_one(f"#{next_id}").focus()
        else:
            self.action_focus_panel_1()

    def action_focus_prev_panel(self) -> None:
        """Focus the previous panel in order."""
        current = self.focused
        if current is None:
            self.action_focus_panel_1()
            return

        current_id = current.id
        if current_id in self._panel_order:
            idx = self._panel_order.index(current_id)
            prev_idx = (idx - 1) % len(self._panel_order)
            prev_id = self._panel_order[prev_idx]
            self.query_one(f"#{prev_id}").focus()
        else:
            self.action_focus_panel_1()

    def on_project_list_project_selected(self, event: ProjectList.ProjectSelected) -> None:
        """Handle project selection."""
        self._current_project = event.project
        self._update_worktree_table(event.project)

    def _update_worktree_table(self, project: str) -> None:
        """Update worktree table for selected project."""
        worktrees = self.state.get_worktrees_for_project(project)

        wt_entries = []
        for w in worktrees:
            workflow = self.state.get_workflow_for_branch(w.repo, w.branch)
            wt_entries.append(WorktreeEntry(
                folder=w.folder,
                repo=w.repo,
                branch=w.branch,
                base=w.base,
                purpose=workflow.purpose if workflow else None,
            ))

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
        self.notify("New worktree: not yet implemented", timeout=2)

    def action_cleanup(self) -> None:
        """Cleanup a worktree (placeholder)."""
        self.notify("Cleanup: not yet implemented", timeout=2)

    def action_daemon(self) -> None:
        """Check daemon status and trigger health refresh if running."""
        try:
            result = subprocess.run(
                ["bearing", "daemon", "status"],
                capture_output=True,
                text=True,
                timeout=5,
            )
            if result.returncode == 0:
                # Daemon is running, trigger refresh in background
                subprocess.Popen(
                    ["bearing", "worktree", "status", "--refresh"],
                    stdout=subprocess.DEVNULL,
                    stderr=subprocess.DEVNULL,
                )
                self.notify("Daemon running, health refresh triggered", timeout=2)
            else:
                self.notify("Daemon not running", timeout=2)
        except FileNotFoundError:
            self.notify("bearing command not found", timeout=2)
        except subprocess.TimeoutExpired:
            self.notify("Daemon status check timed out", timeout=2)

    def action_force_refresh(self) -> None:
        """Force refresh via bearing worktree status --refresh, then reload TUI."""
        self.notify("Force refreshing...", timeout=1)
        try:
            subprocess.run(
                ["bearing", "worktree", "status", "--refresh"],
                capture_output=True,
                timeout=30,
            )
            self.action_refresh()
            self.notify("Force refresh complete", timeout=2)
        except FileNotFoundError:
            self.notify("bearing command not found", timeout=2)
        except subprocess.TimeoutExpired:
            self.notify("Force refresh timed out", timeout=2)

    def action_open_pr(self) -> None:
        """Open PR in browser for selected worktree."""
        from textual.coordinate import Coordinate

        worktree_table = self.query_one(WorktreeTable)
        if worktree_table.row_count == 0:
            self.notify("No worktree selected", timeout=2)
            return

        # Get folder from cursor row key
        try:
            cell_key = worktree_table.coordinate_to_cell_key(
                Coordinate(worktree_table.cursor_row, 0)
            )
            folder = str(cell_key.row_key.value)
        except Exception:
            self.notify("No worktree selected", timeout=2)
            return

        if folder == "empty":
            self.notify("No worktree selected", timeout=2)
            return

        # Check if PR exists
        health = self.state.get_health_for_folder(folder)
        if not health or not health.pr_state:
            self.notify("No PR for this worktree", timeout=2)
            return

        # Get branch for this folder
        local_entry = None
        for e in self.state.read_local():
            if e.folder == folder:
                local_entry = e
                break

        if not local_entry:
            self.notify("Worktree not found", timeout=2)
            return

        # Use gh to get PR URL
        worktree_path = self.workspace / folder
        try:
            result = subprocess.run(
                ["gh", "pr", "view", "--json", "url", "-q", ".url"],
                cwd=worktree_path,
                capture_output=True,
                text=True,
                timeout=10,
            )
            if result.returncode == 0 and result.stdout.strip():
                url = result.stdout.strip()
                webbrowser.open(url)
                self.notify(f"Opened PR", timeout=2)
            else:
                self.notify("Could not get PR URL", timeout=2)
        except FileNotFoundError:
            self.notify("gh command not found", timeout=2)
        except subprocess.TimeoutExpired:
            self.notify("PR lookup timed out", timeout=2)
        except Exception as e:
            self.notify(f"Error: {e}", timeout=2)


def _create_mock_workspace():
    """Create a temporary workspace with mock data for screenshots."""
    import tempfile
    import json

    tmpdir = Path(tempfile.mkdtemp(prefix="bearing-screenshot-"))

    # Mock local.jsonl - impressive scale with many projects and worktrees
    # Note: Projects are sorted alphabetically, so "acme-web" comes first
    local_data = [
        # acme-web - 8 worktrees (main showcase project, first alphabetically)
        {"folder": "acme-web", "repo": "acme-web", "branch": "main", "base": True},
        {"folder": "acme-web-feature-auth", "repo": "acme-web", "branch": "feature-auth", "base": False},
        {"folder": "acme-web-fix-checkout", "repo": "acme-web", "branch": "fix-checkout", "base": False},
        {"folder": "acme-web-perf-images", "repo": "acme-web", "branch": "perf-images", "base": False},
        {"folder": "acme-web-redesign-v2", "repo": "acme-web", "branch": "redesign-v2", "base": False},
        {"folder": "acme-web-dark-mode", "repo": "acme-web", "branch": "dark-mode", "base": False},
        {"folder": "acme-web-i18n", "repo": "acme-web", "branch": "i18n", "base": False},
        {"folder": "acme-web-analytics", "repo": "acme-web", "branch": "analytics", "base": False},
        # acme-api - 5 worktrees
        {"folder": "acme-api", "repo": "acme-api", "branch": "main", "base": True},
        {"folder": "acme-api-graphql", "repo": "acme-api", "branch": "graphql", "base": False},
        {"folder": "acme-api-rate-limit", "repo": "acme-api", "branch": "rate-limit", "base": False},
        {"folder": "acme-api-caching", "repo": "acme-api", "branch": "caching", "base": False},
        {"folder": "acme-api-webhooks", "repo": "acme-api", "branch": "webhooks", "base": False},
        # acme-mobile - 4 worktrees
        {"folder": "acme-mobile", "repo": "acme-mobile", "branch": "main", "base": True},
        {"folder": "acme-mobile-push", "repo": "acme-mobile", "branch": "push-notif", "base": False},
        {"folder": "acme-mobile-offline", "repo": "acme-mobile", "branch": "offline", "base": False},
        {"folder": "acme-mobile-biometric", "repo": "acme-mobile", "branch": "biometric", "base": False},
        # infra - 3 worktrees
        {"folder": "infra", "repo": "infra", "branch": "main", "base": True},
        {"folder": "infra-k8s", "repo": "infra", "branch": "k8s-upgrade", "base": False},
        {"folder": "infra-monitoring", "repo": "infra", "branch": "monitoring", "base": False},
        # shared-libs
        {"folder": "shared-libs", "repo": "shared-libs", "branch": "main", "base": True},
        {"folder": "shared-libs-types", "repo": "shared-libs", "branch": "types", "base": False},
    ]
    with open(tmpdir / "local.jsonl", "w") as f:
        for entry in local_data:
            f.write(json.dumps(entry) + "\n")

    # Mock workflow.jsonl with purposes
    workflow_data = [
        {"repo": "acme-web", "branch": "feature-auth", "basedOn": "main", "purpose": "Add OAuth2 login", "status": "in_progress", "created": "2026-01-15T10:00:00Z"},
        {"repo": "acme-web", "branch": "fix-checkout", "basedOn": "main", "purpose": "Fix cart bug #42", "status": "in_progress", "created": "2026-01-18T14:30:00Z"},
        {"repo": "acme-web", "branch": "perf-images", "basedOn": "main", "purpose": "Lazy load images", "status": "in_progress", "created": "2026-01-17T09:00:00Z"},
        {"repo": "acme-web", "branch": "redesign-v2", "basedOn": "main", "purpose": "New design system", "status": "in_progress", "created": "2026-01-10T11:00:00Z"},
        {"repo": "acme-web", "branch": "dark-mode", "basedOn": "main", "purpose": "Dark theme", "status": "in_progress", "created": "2026-01-14T16:00:00Z"},
        {"repo": "acme-web", "branch": "i18n", "basedOn": "main", "purpose": "i18n support", "status": "in_progress", "created": "2026-01-16T10:00:00Z"},
        {"repo": "acme-web", "branch": "analytics", "basedOn": "main", "purpose": "Add analytics", "status": "in_progress", "created": "2026-01-12T08:00:00Z"},
        {"repo": "acme-api", "branch": "graphql", "basedOn": "main", "purpose": "GraphQL layer", "status": "in_progress", "created": "2026-01-08T09:00:00Z"},
        {"repo": "acme-api", "branch": "rate-limit", "basedOn": "main", "purpose": "Rate limiting", "status": "in_progress", "created": "2026-01-11T13:00:00Z"},
        {"repo": "acme-mobile", "branch": "push-notif", "basedOn": "main", "purpose": "Push notifications", "status": "in_progress", "created": "2026-01-09T10:00:00Z"},
    ]
    with open(tmpdir / "workflow.jsonl", "w") as f:
        for entry in workflow_data:
            f.write(json.dumps(entry) + "\n")

    # Mock health.jsonl with various states (folder names must match local.jsonl)
    health_data = [
        {"folder": "acme-web-feature-auth", "dirty": True, "unpushed": 3, "prState": "OPEN", "lastCheck": "2026-01-19T22:00:00Z"},
        {"folder": "acme-web-fix-checkout", "dirty": False, "unpushed": 0, "prState": "MERGED", "lastCheck": "2026-01-19T22:00:00Z"},
        {"folder": "acme-web-perf-images", "dirty": True, "unpushed": 1, "prState": "OPEN", "lastCheck": "2026-01-19T22:00:00Z"},
        {"folder": "acme-web-redesign-v2", "dirty": False, "unpushed": 12, "prState": "DRAFT", "lastCheck": "2026-01-19T22:00:00Z"},
        {"folder": "acme-web-dark-mode", "dirty": True, "unpushed": 2, "prState": None, "lastCheck": "2026-01-19T22:00:00Z"},
        {"folder": "acme-web-i18n", "dirty": False, "unpushed": 5, "prState": "OPEN", "lastCheck": "2026-01-19T22:00:00Z"},
        {"folder": "acme-web-analytics", "dirty": False, "unpushed": 0, "prState": "OPEN", "lastCheck": "2026-01-19T22:00:00Z"},
        {"folder": "acme-api-graphql", "dirty": True, "unpushed": 8, "prState": "OPEN", "lastCheck": "2026-01-19T22:00:00Z"},
        {"folder": "acme-api-rate-limit", "dirty": False, "unpushed": 0, "prState": "MERGED", "lastCheck": "2026-01-19T22:00:00Z"},
        {"folder": "acme-mobile-push", "dirty": True, "unpushed": 4, "prState": "DRAFT", "lastCheck": "2026-01-19T22:00:00Z"},
    ]
    with open(tmpdir / "health.jsonl", "w") as f:
        for entry in health_data:
            f.write(json.dumps(entry) + "\n")

    return tmpdir


def main():
    """Run the Bearing TUI."""
    import sys

    # Check for --screenshot flag
    if "--screenshot" in sys.argv:
        idx = sys.argv.index("--screenshot")
        output_path = sys.argv[idx + 1] if idx + 1 < len(sys.argv) else "screenshot.svg"

        async def take_screenshot():
            # Use mock data for screenshots
            mock_workspace = _create_mock_workspace()
            app = BearingApp(workspace=mock_workspace)
            async with app.run_test(size=(120, 30)) as pilot:
                # Wait for data to load
                await pilot.pause()
                await pilot.pause()
                # Navigate to first project and select it
                await pilot.press("j")  # Move to first item
                await pilot.pause()
                await pilot.press("enter")  # Select project
                await pilot.pause()
                await pilot.pause()
                # Move to worktree panel and select a row
                await pilot.press("1")  # Focus worktree panel
                await pilot.pause()
                await pilot.press("j")  # Move down to first worktree
                await pilot.press("j")  # Move to second for better visual
                await pilot.pause()
                # Save screenshot
                app.save_screenshot(output_path)
                print(f"Screenshot saved to {output_path}")
            # Cleanup
            import shutil
            shutil.rmtree(mock_workspace, ignore_errors=True)

        import asyncio
        asyncio.run(take_screenshot())
    else:
        app = BearingApp()
        app.run()


if __name__ == "__main__":
    main()
