#!/bin/bash
# Smoke tests for Sailkit
# Run from repo root: ./test/smoke-test.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SAILKIT_DIR="$(dirname "$SCRIPT_DIR")"
TEST_DIR=$(mktemp -d)
PASSED=0
FAILED=0

cleanup() {
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

log() {
    echo "[TEST] $1"
}

pass() {
    echo "  ✓ $1"
    PASSED=$((PASSED + 1))
}

fail() {
    echo "  ✗ $1"
    FAILED=$((FAILED + 1))
}

# Setup test environment
setup() {
    log "Setting up test environment in $TEST_DIR"

    # Copy sailkit
    cp -r "$SAILKIT_DIR" "$TEST_DIR/sailkit-dev"

    # Create a test repo
    git init --initial-branch=main "$TEST_DIR/test-repo" >/dev/null 2>&1
    echo "test" > "$TEST_DIR/test-repo/file.txt"
    git -C "$TEST_DIR/test-repo" add .
    git -C "$TEST_DIR/test-repo" commit -m "initial" >/dev/null 2>&1
}

# Test: worktree-new creates worktree
test_worktree_new() {
    log "Testing worktree-new"

    cd "$TEST_DIR"
    ./sailkit-dev/scripts/worktree-new test-repo feature-branch >/dev/null 2>&1

    if [ -d "$TEST_DIR/test-repo-feature-branch" ]; then
        pass "Worktree directory created"
    else
        fail "Worktree directory not created"
    fi

    if git -C "$TEST_DIR/test-repo-feature-branch" rev-parse --abbrev-ref HEAD | grep -q "feature-branch"; then
        pass "Worktree on correct branch"
    else
        fail "Worktree not on correct branch"
    fi

    if grep -q "test-repo-feature-branch" "$TEST_DIR/sailkit-dev/WORKTREES.md"; then
        pass "Manifest updated"
    else
        fail "Manifest not updated"
    fi
}

# Test: worktree-cleanup removes worktree
test_worktree_cleanup() {
    log "Testing worktree-cleanup"

    cd "$TEST_DIR"
    ./sailkit-dev/scripts/worktree-cleanup test-repo feature-branch >/dev/null 2>&1

    if [ ! -d "$TEST_DIR/test-repo-feature-branch" ]; then
        pass "Worktree directory removed"
    else
        fail "Worktree directory not removed"
    fi

    if ! grep -q "test-repo-feature-branch" "$TEST_DIR/sailkit-dev/WORKTREES.md"; then
        pass "Manifest entry removed"
    else
        fail "Manifest entry not removed"
    fi
}

# Test: worktree-sync updates manifest
test_worktree_sync() {
    log "Testing worktree-sync"

    cd "$TEST_DIR"

    # Manually create a worktree
    git -C "$TEST_DIR/test-repo" worktree add "$TEST_DIR/test-repo-manual-branch" -b manual-branch >/dev/null 2>&1

    # Sync should find it
    ./sailkit-dev/scripts/worktree-sync >/dev/null 2>&1

    if grep -q "test-repo-manual-branch" "$TEST_DIR/sailkit-dev/WORKTREES.md"; then
        pass "Sync found manually created worktree"
    else
        fail "Sync did not find manually created worktree"
    fi
}

# Test: install.sh creates symlinks
test_install() {
    log "Testing install.sh"

    cd "$TEST_DIR"
    printf "1\ny\n" | ./sailkit-dev/install.sh >/dev/null 2>&1

    if [ -L "$TEST_DIR/.claude/skills/worktree" ]; then
        pass "Symlink created"
    else
        fail "Symlink not created"
    fi

    if [ -f "$TEST_DIR/.claude/skills/worktree/SKILL.md" ]; then
        pass "Skill file accessible via symlink"
    else
        fail "Skill file not accessible via symlink"
    fi
}

# Run tests
setup
test_worktree_new
test_worktree_cleanup
test_worktree_sync
test_install

# Summary
echo ""
echo "================================"
echo "Results: $PASSED passed, $FAILED failed"
echo "================================"

if [ $FAILED -gt 0 ]; then
    exit 1
fi
