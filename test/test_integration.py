"""Integration tests for Sailkit commands.

These tests run actual scripts in isolated temp directories,
mocking stdin/stdout to test the full user experience.
"""
import pytest

from conftest import SailkitTestEnv


class TestWorktreeNew:
    """Tests for worktree-new command."""

    def test_creates_worktree_directory(self, env: SailkitTestEnv):
        env.create_repo("test-repo")
        result = env.run("worktree-new", "test-repo", "feature-branch")

        assert result.success
        assert env.worktree_exists("test-repo-feature-branch")

    def test_worktree_on_correct_branch(self, env: SailkitTestEnv):
        env.create_repo("test-repo")
        env.run("worktree-new", "test-repo", "feature-branch")

        assert env.get_branch("test-repo-feature-branch") == "feature-branch"

    def test_updates_workflow_jsonl(self, env: SailkitTestEnv):
        env.create_repo("test-repo")
        env.run("worktree-new", "test-repo", "feature-branch")

        entry = env.find_workflow_entry("test-repo", "feature-branch")
        assert entry is not None
        assert entry["status"] == "in_progress"

    def test_updates_local_jsonl(self, env: SailkitTestEnv):
        env.create_repo("test-repo")
        env.run("worktree-new", "test-repo", "feature-branch")

        entry = env.find_local_entry("test-repo-feature-branch")
        assert entry is not None
        assert entry["base"] is False

    def test_with_purpose_flag(self, env: SailkitTestEnv):
        env.create_repo("test-repo")
        env.run("worktree-new", "test-repo", "feature-branch", "--purpose", "Add login")

        entry = env.find_workflow_entry("test-repo", "feature-branch")
        assert entry["purpose"] == "Add login"

    def test_with_based_on_flag(self, env: SailkitTestEnv):
        env.create_repo("test-repo")
        # Create develop branch first so --based-on can reference it
        import subprocess
        subprocess.run(
            ["git", "-C", str(env.root / "test-repo"), "branch", "develop"],
            check=True
        )
        env.run("worktree-new", "test-repo", "feature-branch", "--based-on", "develop")

        entry = env.find_workflow_entry("test-repo", "feature-branch")
        assert entry["basedOn"] == "develop"


class TestWorktreeCleanup:
    """Tests for worktree-cleanup command."""

    def test_removes_worktree_directory(self, env: SailkitTestEnv):
        env.create_repo("test-repo")
        env.run("worktree-new", "test-repo", "feature-branch")
        env.run("worktree-cleanup", "test-repo", "feature-branch")

        assert not env.worktree_exists("test-repo-feature-branch")

    def test_removes_local_jsonl_entry(self, env: SailkitTestEnv):
        env.create_repo("test-repo")
        env.run("worktree-new", "test-repo", "feature-branch")
        env.run("worktree-cleanup", "test-repo", "feature-branch")

        assert env.find_local_entry("test-repo-feature-branch") is None

    def test_updates_workflow_status(self, env: SailkitTestEnv):
        env.create_repo("test-repo")
        env.run("worktree-new", "test-repo", "feature-branch")
        env.run("worktree-cleanup", "test-repo", "feature-branch")

        entry = env.find_workflow_entry("test-repo", "feature-branch")
        assert entry["status"] == "completed"


class TestWorktreeRegister:
    """Tests for worktree-register command."""

    def test_registers_base_repo(self, env: SailkitTestEnv):
        env.create_repo("test-repo")
        env.run("worktree-register", "test-repo")

        entry = env.find_local_entry("test-repo")
        assert entry is not None

    def test_marks_as_base(self, env: SailkitTestEnv):
        env.create_repo("test-repo")
        env.run("worktree-register", "test-repo")

        entry = env.find_local_entry("test-repo")
        assert entry["base"] is True


class TestWorktreeSync:
    """Tests for worktree-sync command."""

    def test_discovers_manually_created_worktree(self, env: SailkitTestEnv):
        env.create_repo("test-repo")
        # Manually create worktree (bypassing sailkit)
        import subprocess
        subprocess.run(
            ["git", "-C", str(env.root / "test-repo"), "worktree", "add",
             str(env.root / "test-repo-manual-branch"), "-b", "manual-branch"],
            capture_output=True, check=True
        )

        env.run("worktree-sync")

        assert env.find_local_entry("test-repo-manual-branch") is not None

    def test_discovers_base_repo(self, env: SailkitTestEnv):
        env.create_repo("test-repo")
        import subprocess
        subprocess.run(
            ["git", "-C", str(env.root / "test-repo"), "worktree", "add",
             str(env.root / "test-repo-feature"), "-b", "feature"],
            capture_output=True, check=True
        )

        env.run("worktree-sync")

        entry = env.find_local_entry("test-repo")
        assert entry is not None
        assert entry["base"] is True


class TestWorktreeList:
    """Tests for worktree-list command."""

    def test_shows_header(self, env: SailkitTestEnv):
        env.create_repo("test-repo")
        env.run("worktree-register", "test-repo")
        result = env.run("worktree-list")

        assert "FOLDER" in result.stdout

    def test_shows_entries(self, env: SailkitTestEnv):
        env.create_repo("test-repo")
        env.run("worktree-register", "test-repo")
        result = env.run("worktree-list")

        assert "test-repo" in result.stdout


class TestInstaller:
    """Tests for install.sh."""

    def test_creates_symlink(self, env: SailkitTestEnv):
        env.run_installer(stdin="1\ny\n")

        symlink = env.root / ".claude" / "skills" / "worktree"
        assert symlink.is_symlink()

    def test_skill_accessible(self, env: SailkitTestEnv):
        env.run_installer(stdin="1\ny\n")

        skill_file = env.root / ".claude" / "skills" / "worktree" / "SKILL.md"
        assert skill_file.exists()


class TestWorktreeCheck:
    """Tests for worktree-check command."""

    def test_passes_when_base_on_main(self, env: SailkitTestEnv):
        env.create_repo("test-repo")
        env.run("worktree-register", "test-repo")
        result = env.run("worktree-check")

        assert result.success

    def test_quiet_mode_no_output_on_success(self, env: SailkitTestEnv):
        env.create_repo("test-repo")
        env.run("worktree-register", "test-repo")
        result = env.run("worktree-check", "--quiet")

        assert result.success
        assert result.stdout.strip() == ""

    def test_always_exits_zero(self, env: SailkitTestEnv):
        """Script always exits 0 for hook compatibility."""
        env.create_repo("test-repo")
        env.run("worktree-register", "test-repo")
        # Switch base to feature branch (violation)
        import subprocess
        subprocess.run(
            ["git", "-C", str(env.root / "test-repo"), "checkout", "-b", "feature"],
            capture_output=True, check=True
        )
        env.run("worktree-sync")  # Update manifest
        result = env.run("worktree-check")

        assert result.success  # Still exits 0

    def test_json_output_no_violations(self, env: SailkitTestEnv):
        """JSON mode outputs valid JSON with continue:true when no violations."""
        import json
        env.create_repo("test-repo")
        env.run("worktree-register", "test-repo")
        result = env.run("worktree-check", "--json")

        assert result.success
        data = json.loads(result.stdout)
        assert data["continue"] is True
        assert "systemMessage" not in data

    def test_json_output_with_violations(self, env: SailkitTestEnv):
        """JSON mode includes systemMessage when violations exist."""
        import json
        import subprocess
        env.create_repo("test-repo")
        env.run("worktree-register", "test-repo")
        subprocess.run(
            ["git", "-C", str(env.root / "test-repo"), "checkout", "-b", "feature"],
            capture_output=True, check=True
        )
        env.run("worktree-sync")
        result = env.run("worktree-check", "--json")

        assert result.success
        data = json.loads(result.stdout)
        assert data["continue"] is True
        assert "systemMessage" in data
        assert "test-repo" in data["systemMessage"]
        assert "feature" in data["systemMessage"]


class TestEdgeCases:
    """Edge cases and error handling."""

    def test_worktree_new_nonexistent_repo(self, env: SailkitTestEnv):
        result = env.run("worktree-new", "nonexistent", "feature")
        assert not result.success

    def test_worktree_cleanup_nonexistent_worktree(self, env: SailkitTestEnv):
        env.create_repo("test-repo")
        result = env.run("worktree-cleanup", "test-repo", "nonexistent")
        # Should not fail catastrophically
        assert result.returncode in [0, 1]

    def test_multiple_worktrees_same_repo(self, env: SailkitTestEnv):
        env.create_repo("test-repo")
        env.run("worktree-new", "test-repo", "feature-a")
        env.run("worktree-new", "test-repo", "feature-b")

        assert env.worktree_exists("test-repo-feature-a")
        assert env.worktree_exists("test-repo-feature-b")
        assert len([e for e in env.read_local() if e.get("repo") == "test-repo"]) == 2
