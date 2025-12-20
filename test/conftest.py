"""Pytest fixtures for Sailkit integration tests."""
import json
import os
import shutil
import subprocess
import tempfile
from dataclasses import dataclass
from pathlib import Path
from typing import Optional

import pytest


@dataclass
class CommandResult:
    """Result of running a Sailkit command."""
    returncode: int
    stdout: str
    stderr: str

    @property
    def success(self) -> bool:
        return self.returncode == 0


class SailkitTestEnv:
    """Isolated test environment for Sailkit commands."""

    def __init__(self, tmp_path: Path, sailkit_src: Path):
        self.root = tmp_path
        self.sailkit_dir = tmp_path / "sailkit-dev"
        self.workflow_file = tmp_path / "workflow.jsonl"
        self.local_file = tmp_path / "local.jsonl"

        # Copy sailkit scripts
        shutil.copytree(sailkit_src, self.sailkit_dir)

        # Initialize empty state files
        self.workflow_file.touch()
        self.local_file.touch()

    def create_repo(self, name: str, initial_branch: str = "main") -> Path:
        """Create a test git repository."""
        repo_path = self.root / name
        subprocess.run(
            ["git", "init", "--initial-branch", initial_branch, str(repo_path)],
            capture_output=True, check=True
        )
        # Create initial commit
        (repo_path / "README.md").write_text(f"# {name}\n")
        subprocess.run(["git", "-C", str(repo_path), "add", "."], check=True)
        subprocess.run(
            ["git", "-C", str(repo_path), "commit", "-m", "initial"],
            capture_output=True, check=True
        )
        return repo_path

    def run(
        self,
        script: str,
        *args: str,
        stdin: Optional[str] = None,
        env: Optional[dict] = None
    ) -> CommandResult:
        """Run a Sailkit script with optional stdin."""
        script_path = self.sailkit_dir / "scripts" / script
        cmd = [str(script_path)] + list(args)

        run_env = os.environ.copy()
        if env:
            run_env.update(env)

        result = subprocess.run(
            cmd,
            cwd=str(self.root),
            input=stdin,
            capture_output=True,
            text=True,
            env=run_env
        )
        return CommandResult(result.returncode, result.stdout, result.stderr)

    def run_installer(self, stdin: str = "1\ny\n") -> CommandResult:
        """Run install.sh with mocked stdin."""
        result = subprocess.run(
            [str(self.sailkit_dir / "install.sh")],
            cwd=str(self.root),
            input=stdin,
            capture_output=True,
            text=True
        )
        return CommandResult(result.returncode, result.stdout, result.stderr)

    def read_workflow(self) -> list[dict]:
        """Parse workflow.jsonl into list of dicts."""
        if not self.workflow_file.exists():
            return []
        lines = self.workflow_file.read_text().strip().split("\n")
        return [json.loads(line) for line in lines if line.strip()]

    def read_local(self) -> list[dict]:
        """Parse local.jsonl into list of dicts."""
        if not self.local_file.exists():
            return []
        lines = self.local_file.read_text().strip().split("\n")
        return [json.loads(line) for line in lines if line.strip()]

    def find_workflow_entry(self, repo: str, branch: str) -> Optional[dict]:
        """Find a workflow entry by repo and branch."""
        for entry in self.read_workflow():
            if entry.get("repo") == repo and entry.get("branch") == branch:
                return entry
        return None

    def find_local_entry(self, folder: str) -> Optional[dict]:
        """Find a local entry by folder name."""
        for entry in self.read_local():
            if entry.get("folder") == folder:
                return entry
        return None

    def worktree_exists(self, name: str) -> bool:
        """Check if a worktree directory exists."""
        return (self.root / name).is_dir()

    def get_branch(self, folder: str) -> str:
        """Get current branch of a repo/worktree."""
        result = subprocess.run(
            ["git", "-C", str(self.root / folder), "rev-parse", "--abbrev-ref", "HEAD"],
            capture_output=True, text=True
        )
        return result.stdout.strip()


@pytest.fixture
def sailkit_src() -> Path:
    """Path to the actual Sailkit source directory."""
    return Path(__file__).parent.parent


@pytest.fixture
def env(tmp_path: Path, sailkit_src: Path) -> SailkitTestEnv:
    """Create an isolated Sailkit test environment."""
    return SailkitTestEnv(tmp_path, sailkit_src)
