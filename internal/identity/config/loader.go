package config

import (
	"fmt"
	"os"

	cryptoutilMagic "cryptoutil/internal/common/magic"

	"gopkg.in/yaml.v3"
)

// LoadFromFile reads a configuration file and returns a Config instance.
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate after loading.
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// SaveToFile writes a Config instance to a YAML file.
func SaveToFile(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, cryptoutilMagic.FilePermOwnerReadWriteOnly); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
