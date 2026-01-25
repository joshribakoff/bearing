"""Bearing TUI widgets."""
from .projects import ProjectList
from .worktrees import WorktreeTable, WorktreeEntry, HealthEntry
from .details import DetailsPanel, LocalEntry, WorkflowEntry
from .details import HealthEntry as DetailsHealthEntry
from .plans import PlansList, load_plans
from .plans import PlanEntry as LegacyPlanEntry
from .plans_table import PlansTable, PlanEntry
from .prs import PRsTable, PRDisplayEntry

__all__ = [
    "ProjectList",
    "WorktreeTable",
    "WorktreeEntry",
    "HealthEntry",
    "DetailsPanel",
    "LocalEntry",
    "WorkflowEntry",
    "PlansList",
    "PlansTable",
    "PlanEntry",
    "LegacyPlanEntry",
    "load_plans",
    "PRsTable",
    "PRDisplayEntry",
]
