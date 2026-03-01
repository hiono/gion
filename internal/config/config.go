package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the gion configuration
type Config struct {
	// Provider specifies the default provider to use (github, gitlab, bitbucket)
	// Empty string means auto-detect
	Provider string `yaml:"provider"`
}

// Valid providers
var validProviders = map[string]bool{
	"github":    true,
	"gitlab":    true,
	"bitbucket": true,
}

// DefaultConfigDir is the default config directory
const DefaultConfigDir = ".gion"

// DefaultConfigFile is the default config file name
const DefaultConfigFile = "config.yaml"

// Load loads configuration from the default config file
func Load() (*Config, error) {
	configPath, err := DefaultConfigPath()
	if err != nil {
		return nil, err
	}
	return LoadFromFile(configPath)
}

// LoadFromFile loads configuration from a specific file
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate provider if specified
	if cfg.Provider != "" && !validProviders[strings.ToLower(cfg.Provider)] {
		return nil, fmt.Errorf("invalid provider: %s (valid: github, gitlab, bitbucket)", cfg.Provider)
	}

	return &cfg, nil
}

// Save saves configuration to the default config file
func Save(cfg *Config) error {
	configPath, err := DefaultConfigPath()
	if err != nil {
		return err
	}
	return SaveToFile(cfg, configPath)
}

// SaveToFile saves configuration to a specific file
func SaveToFile(cfg *Config, path string) error {
	// Validate provider if specified
	if cfg.Provider != "" && !validProviders[strings.ToLower(cfg.Provider)] {
		return fmt.Errorf("invalid provider: %s (valid: github, gitlab, bitbucket)", cfg.Provider)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// DefaultConfigPath returns the default config file path
func DefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, DefaultConfigDir, DefaultConfigFile), nil
}

// Provider returns the configured provider (empty means auto-detect)
func (c *Config) ProviderName() string {
	return strings.ToLower(c.Provider)
}
