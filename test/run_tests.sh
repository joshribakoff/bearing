#!/bin/bash
set -e

# Integration test that exercises real git operations
# BEARING_AI_ENABLED=0 to skip AI features

export BEARING_AI_ENABLED=0
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

echo "Test directory: $TEST_DIR"

# 1. Create test workspace
echo "1. Creating test repo..."
mkdir -p "$TEST_DIR/test-repo"
git -C "$TEST_DIR/test-repo" init --initial-branch=main
git -C "$TEST_DIR/test-repo" config user.email "test@test.com"
git -C "$TEST_DIR/test-repo" config user.name "Test"
git -C "$TEST_DIR/test-repo" commit --allow-empty -m "initial"

# 2. Initialize state files
echo "2. Initializing state files..."
touch "$TEST_DIR/workflow.jsonl" "$TEST_DIR/local.jsonl"

# 3. Test worktree-new
echo "3. Testing worktree new..."
bearing -w "$TEST_DIR" worktree new test-repo feature-x --purpose "Test"
[ -d "$TEST_DIR/test-repo-feature-x" ] || { echo "FAIL: worktree not created"; exit 1; }
echo "   OK: worktree created"

# 4. Test worktree-list
echo "4. Testing worktree list..."
bearing -w "$TEST_DIR" worktree list --json | grep -q "feature-x" || { echo "FAIL: worktree not in list"; exit 1; }
echo "   OK: worktree in list"

# 5. Test worktree-status
echo "5. Testing worktree status..."
bearing -w "$TEST_DIR" worktree status --json | grep -q "test-repo-feature-x" || { echo "FAIL: status failed"; exit 1; }
echo "   OK: status works"

# 6. Test worktree-check (Claude Code hook format)
echo "6. Testing worktree check..."
bearing -w "$TEST_DIR" worktree check --json | grep -q '"continue":true' || { echo "FAIL: check failed"; exit 1; }
echo "   OK: check works"

# 7. Test daemon status
echo "7. Testing daemon status..."
bearing -w "$TEST_DIR" daemon status | grep -q "not running" || { echo "FAIL: daemon status failed"; exit 1; }
echo "   OK: daemon status works"

# 8. Test worktree-cleanup
echo "8. Testing worktree cleanup..."
bearing -w "$TEST_DIR" worktree cleanup test-repo feature-x
[ ! -d "$TEST_DIR/test-repo-feature-x" ] || { echo "FAIL: worktree not removed"; exit 1; }
echo "   OK: worktree removed"

# 9. Test worktree-sync
echo "9. Testing worktree sync..."
bearing -w "$TEST_DIR" worktree sync
cat "$TEST_DIR/local.jsonl" | grep -q "test-repo" || { echo "FAIL: sync failed"; exit 1; }
echo "   OK: sync works"

# 10. Test init (idempotent hook setup)
echo "10. Testing init..."
bearing -w "$TEST_DIR" init
bearing -w "$TEST_DIR" init | grep -q "already configured" || { echo "FAIL: init not idempotent"; exit 1; }
[ -f "$TEST_DIR/.claude/settings.json" ] || { echo "FAIL: settings.json not created"; exit 1; }
echo "   OK: init works and is idempotent"

echo ""
echo "All tests passed!"
