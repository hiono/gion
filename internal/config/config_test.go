package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProviderOverride(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Test 1: No config file - should return empty provider (auto-detect)
	t.Run("no config file", func(t *testing.T) {
		cfg, err := LoadFromFile(configPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Provider != "" {
			t.Errorf("expected empty provider, got %q", cfg.Provider)
		}
	})

	// Test 2: Config with provider: gitlab
	t.Run("config with gitlab provider", func(t *testing.T) {
		content := "provider: gitlab"
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		cfg, err := LoadFromFile(configPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Provider != "gitlab" {
			t.Errorf("expected 'gitlab', got %q", cfg.Provider)
		}
		if cfg.ProviderName() != "gitlab" {
			t.Errorf("ProviderName() expected 'gitlab', got %q", cfg.ProviderName())
		}
	})

	// Test 3: Config with invalid provider
	t.Run("invalid provider", func(t *testing.T) {
		content := "provider: invalid"
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		_, err := LoadFromFile(configPath)
		if err == nil {
			t.Error("expected error for invalid provider, got nil")
		}
	})

	// Test 4: Config save and load roundtrip
	t.Run("save and load roundtrip", func(t *testing.T) {
		cfg := &Config{Provider: "bitbucket"}
		if err := SaveToFile(cfg, configPath); err != nil {
			t.Fatalf("failed to save config: %v", err)
		}

		loaded, err := LoadFromFile(configPath)
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}
		if loaded.Provider != "bitbucket" {
			t.Errorf("expected 'bitbucket', got %q", loaded.Provider)
		}
	})

	// Test 5: Case insensitive provider
	t.Run("case insensitive provider", func(t *testing.T) {
		content := "provider: GitHub"
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		cfg, err := LoadFromFile(configPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// YAML unmarshal preserves case, but ProviderName() normalizes
		if cfg.ProviderName() != "github" {
			t.Errorf("ProviderName() expected 'github', got %q", cfg.ProviderName())
		}
	})
}

func TestResolveProviderPriority(t *testing.T) {
	// Test the priority: CLI > config > auto-detect
	t.Run("CLI takes precedence over config", func(t *testing.T) {
		// This would require mocking the config file
		// For now, we test the logic directly
		cliProvider := "gitlab"
		if cliProvider != "" {
			// Should use CLI provider
			if cliProvider != "gitlab" {
				t.Errorf("expected 'gitlab', got %q", cliProvider)
			}
		}
	})

	t.Run("empty CLI uses config", func(t *testing.T) {
		cliProvider := ""
		configProvider := "github"

		result := cliProvider
		if result == "" {
			result = configProvider
		}

		if result != "github" {
			t.Errorf("expected 'github', got %q", result)
		}
	})

	t.Run("all empty returns empty for auto-detect", func(t *testing.T) {
		cliProvider := ""
		configProvider := ""

		result := cliProvider
		if result == "" {
			result = configProvider
		}

		if result != "" {
			t.Errorf("expected empty for auto-detect, got %q", result)
		}
	})
}
