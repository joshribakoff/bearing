package daemon

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// PlanSyncConfig holds plan sync configuration
type PlanSyncConfig struct {
	Enabled         bool     `json:"enabled"`
	IntervalMinutes int      `json:"interval_minutes"`
	Projects        []string `json:"projects"`
	MaxAPICallsPerCycle int  `json:"max_api_calls_per_cycle"`
}

// BearingConfig holds the full ~/.bearing/config.json structure
type BearingConfig struct {
	PlanSync PlanSyncConfig `json:"plan_sync"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() BearingConfig {
	return BearingConfig{
		PlanSync: PlanSyncConfig{
			Enabled:             true,
			IntervalMinutes:     5,
			Projects:            []string{}, // empty means all
			MaxAPICallsPerCycle: 5,
		},
	}
}

// LoadConfig loads config from ~/.bearing/config.json
func LoadConfig(bearingDir string) (BearingConfig, error) {
	cfg := DefaultConfig()

	configPath := filepath.Join(bearingDir, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Use defaults
		}
		return cfg, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	// Apply defaults for missing fields
	if cfg.PlanSync.MaxAPICallsPerCycle == 0 {
		cfg.PlanSync.MaxAPICallsPerCycle = 5
	}
	if cfg.PlanSync.IntervalMinutes == 0 {
		cfg.PlanSync.IntervalMinutes = 5
	}

	return cfg, nil
}
