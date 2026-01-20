package daemon

import (
	"time"

	"github.com/joshribakoff/bearing/internal/jsonl"
)

// HealthChecker performs health checks on worktrees
type HealthChecker struct {
	store *jsonl.Store
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(store *jsonl.Store) *HealthChecker {
	return &HealthChecker{store: store}
}

// IsStale returns true if the health data is older than the given duration
func IsStale(entry jsonl.HealthEntry, maxAge time.Duration) bool {
	return time.Since(entry.LastCheck) > maxAge
}

// NeedsAttention returns true if the worktree needs user attention
func NeedsAttention(entry jsonl.HealthEntry, local jsonl.LocalEntry) bool {
	// Base folders shouldn't have uncommitted changes
	if local.Base && entry.Dirty {
		return true
	}

	// Non-base folders with unpushed commits
	if !local.Base && entry.Unpushed > 0 {
		return true
	}

	// PR in mergeable state that hasn't been merged
	if entry.PRState != nil && *entry.PRState == "OPEN" {
		return true
	}

	return false
}
