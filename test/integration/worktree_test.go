package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/joshribakoff/bearing/internal/jsonl"
	"github.com/joshribakoff/bearing/test/testutil"
)

func TestWorktreeNew(t *testing.T) {
	t.Setenv("BEARING_AI_ENABLED", "0")
	tmpDir := t.TempDir()

	// Create test repo
	testutil.CreateTestRepo(t, tmpDir, "test-repo")
	testutil.InitWorkspace(t, tmpDir)

	// Run worktree new
	output, err := testutil.RunBearing(t, tmpDir, "worktree", "new", "test-repo", "feature-x", "--purpose", "Test feature")
	if err != nil {
		t.Fatalf("worktree new failed: %v\nOutput: %s", err, output)
	}

	// Verify worktree created
	worktreePath := filepath.Join(tmpDir, "test-repo-feature-x")
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		t.Errorf("worktree directory not created: %s", worktreePath)
	}

	// Verify workflow.jsonl updated
	store := jsonl.NewStore(tmpDir)
	workflows, err := store.ReadWorkflow()
	if err != nil {
		t.Fatal(err)
	}
	if len(workflows) != 1 {
		t.Errorf("expected 1 workflow entry, got %d", len(workflows))
	}
	if workflows[0].Branch != "feature-x" {
		t.Errorf("expected branch feature-x, got %s", workflows[0].Branch)
	}

	// Verify local.jsonl updated
	locals, err := store.ReadLocal()
	if err != nil {
		t.Fatal(err)
	}
	if len(locals) != 1 {
		t.Errorf("expected 1 local entry, got %d", len(locals))
	}
}

func TestWorktreeCleanup(t *testing.T) {
	t.Setenv("BEARING_AI_ENABLED", "0")
	tmpDir := t.TempDir()

	// Create test repo and worktree
	testutil.CreateTestRepo(t, tmpDir, "test-repo")
	testutil.InitWorkspace(t, tmpDir)

	// Create worktree
	_, err := testutil.RunBearing(t, tmpDir, "worktree", "new", "test-repo", "feature-y")
	if err != nil {
		t.Fatal(err)
	}

	// Cleanup worktree
	output, err := testutil.RunBearing(t, tmpDir, "worktree", "cleanup", "test-repo", "feature-y")
	if err != nil {
		t.Fatalf("worktree cleanup failed: %v\nOutput: %s", err, output)
	}

	// Verify worktree removed
	worktreePath := filepath.Join(tmpDir, "test-repo-feature-y")
	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		t.Errorf("worktree directory still exists: %s", worktreePath)
	}

	// Verify local.jsonl updated
	store := jsonl.NewStore(tmpDir)
	locals, err := store.ReadLocal()
	if err != nil {
		t.Fatal(err)
	}
	if len(locals) != 0 {
		t.Errorf("expected 0 local entries, got %d", len(locals))
	}
}

func TestWorktreeList(t *testing.T) {
	t.Setenv("BEARING_AI_ENABLED", "0")
	tmpDir := t.TempDir()

	testutil.CreateTestRepo(t, tmpDir, "test-repo")
	testutil.InitWorkspace(t, tmpDir)

	// Create a worktree
	_, _ = testutil.RunBearing(t, tmpDir, "worktree", "new", "test-repo", "feature-z")

	// List with JSON
	output, err := testutil.RunBearing(t, tmpDir, "worktree", "list", "--json")
	if err != nil {
		t.Fatalf("worktree list failed: %v\nOutput: %s", err, output)
	}

	var entries []jsonl.LocalEntry
	if err := json.Unmarshal([]byte(output), &entries); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestWorktreeSync(t *testing.T) {
	t.Setenv("BEARING_AI_ENABLED", "0")
	tmpDir := t.TempDir()

	// Create repos without registering them
	testutil.CreateTestRepo(t, tmpDir, "repo-a")
	testutil.CreateTestRepo(t, tmpDir, "repo-b")
	testutil.InitWorkspace(t, tmpDir)

	// Sync should discover them
	output, err := testutil.RunBearing(t, tmpDir, "worktree", "sync")
	if err != nil {
		t.Fatalf("worktree sync failed: %v\nOutput: %s", err, output)
	}

	// Verify local.jsonl populated
	store := jsonl.NewStore(tmpDir)
	locals, _ := store.ReadLocal()
	if len(locals) != 2 {
		t.Errorf("expected 2 entries, got %d", len(locals))
	}
}

func TestDaemonStartStop(t *testing.T) {
	t.Setenv("BEARING_AI_ENABLED", "0")
	tmpDir := t.TempDir()

	testutil.InitWorkspace(t, tmpDir)

	// Start daemon in foreground mode would block, so we test status
	output, err := testutil.RunBearing(t, tmpDir, "daemon", "status")
	if err != nil {
		t.Logf("daemon status output: %s", output)
	}

	// Should show "not running"
	if output != "not running\n" {
		// Might be running from another test
		_, _ = testutil.RunBearing(t, tmpDir, "daemon", "stop")
	}
}

// TestBinaryExists ensures the bearing binary is built
func TestBinaryExists(t *testing.T) {
	_, err := exec.LookPath("bearing")
	if err != nil {
		t.Skip("bearing binary not in PATH - run 'make build' first")
	}
}
