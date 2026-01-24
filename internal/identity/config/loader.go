// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"fmt"
	"os"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gopkg.in/yaml.v3"
)

// LoadFromFile reads a configuration file and returns a Config instance.
// Environment variables in the config file (e.g., ${VAR_NAME}) are expanded before parsing.
// Fields not specified in the config file retain their default values.
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables in the config file content
	expandedData := os.ExpandEnv(string(data))

	// Start from default config to ensure all fields have sensible values.
	cfg := DefaultConfig()
	if err := yaml.Unmarshal([]byte(expandedData), cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate after loading.
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// SaveToFile writes a Config instance to a YAML file.
func SaveToFile(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, cryptoutilSharedMagic.FilePermOwnerReadWriteOnly); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
