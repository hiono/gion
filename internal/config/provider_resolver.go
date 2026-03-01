package config

import (
	"fmt"
	"strings"
)

// ResolveProvider resolves the provider using the lookup priority:
// 1. CLI flag (provided)
// 2. Config file
// 3. Auto-detect (empty)
func ResolveProvider(cliProvider string) (string, error) {
	// Priority 1: CLI flag
	if cliProvider != "" {
		cliProvider = strings.ToLower(cliProvider)
		if !validProviders[cliProvider] {
			return "", fmt.Errorf("invalid provider: %s (valid: github, gitlab, bitbucket)", cliProvider)
		}
		return cliProvider, nil
	}

	// Priority 2: Config file
	cfg, err := Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Provider != "" {
		return strings.ToLower(cfg.Provider), nil
	}

	// Priority 3: Auto-detect (return empty)
	return "", nil
}
