// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ProfileConfig represents a profile configuration file.
type ProfileConfig struct {
	Services ServiceConfigs `yaml:"services"`
}

// ServiceConfigs holds configurations for all identity services.
type ServiceConfigs struct {
	AuthZ ServiceConfig `yaml:"authz"`
	IDP   ServiceConfig `yaml:"idp"`
	RS    ServiceConfig `yaml:"rs"`
}

// ServiceConfig represents configuration for a single service.
type ServiceConfig struct {
	Enabled     bool   `yaml:"enabled"`
	BindAddress string `yaml:"bind_address"`
	DatabaseURL string `yaml:"database_url"`
	LogLevel    string `yaml:"log_level"`
}

// LoadProfile loads a profile configuration by name from configs/identity/profiles/.
func LoadProfile(profileName string) (*ProfileConfig, error) {
	// Get project root directory - go up 3 levels from internal/identity/config
	projectRoot, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	// Check if we're in a subdirectory (tests run from package directory)
	if filepath.Base(projectRoot) == "config" {
		projectRoot = filepath.Join(projectRoot, "..", "..", "..")
	}

	profilePath := filepath.Join(projectRoot, "configs", "identity", "profiles", profileName+".yml")

	return LoadProfileFromFile(profilePath)
}

// LoadProfileFromFile loads a profile configuration from a custom file path.
func LoadProfileFromFile(filePath string) (*ProfileConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile file %s: %w", filePath, err)
	}

	var cfg ProfileConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse profile YAML: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid profile configuration: %w", err)
	}

	return &cfg, nil
}

// Validate checks if the profile configuration is valid.
func (c *ProfileConfig) Validate() error {
	if !c.Services.AuthZ.Enabled && !c.Services.IDP.Enabled && !c.Services.RS.Enabled {
		return fmt.Errorf("at least one service must be enabled")
	}

	if c.Services.AuthZ.Enabled {
		if err := c.Services.AuthZ.validate("authz"); err != nil {
			return err
		}
	}

	if c.Services.IDP.Enabled {
		if err := c.Services.IDP.validate("idp"); err != nil {
			return err
		}
	}

	if c.Services.RS.Enabled {
		if err := c.Services.RS.validate("rs"); err != nil {
			return err
		}
	}

	return nil
}

// validate checks if a single service configuration is valid.
func (s *ServiceConfig) validate(serviceName string) error {
	if s.BindAddress == "" {
		return fmt.Errorf("%s: bind_address is required", serviceName)
	}

	if s.DatabaseURL == "" && (serviceName == "authz" || serviceName == "idp") {
		return fmt.Errorf("%s: database_url is required", serviceName)
	}

	if s.LogLevel == "" {
		return fmt.Errorf("%s: log_level is required", serviceName)
	}

	// Validate log level is recognized.
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLogLevels[s.LogLevel] {
		return fmt.Errorf("%s: invalid log_level '%s' (must be debug, info, warn, or error)", serviceName, s.LogLevel)
	}

	return nil
}
