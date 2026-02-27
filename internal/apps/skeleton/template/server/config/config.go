// Copyright (c) 2025 Justin Cranford
//

// Package config provides configuration management for skeleton-template service.
package config

import (
"fmt"
"strings"

cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

"github.com/spf13/pflag"
)

// SkeletonTemplateServerSettings contains skeleton-template specific configuration.
// The skeleton-template service has no domain-specific settings beyond the base template.
type SkeletonTemplateServerSettings struct {
*cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
}

// ParseWithFlagSet parses command line arguments using provided FlagSet and returns skeleton-template settings.
// This enables test isolation by allowing each test to use its own FlagSet.
func ParseWithFlagSet(fs *pflag.FlagSet, args []string, exitIfHelp bool) (*SkeletonTemplateServerSettings, error) {
// Parse base template settings using the provided FlagSet.
// This will register template flags and call fs.Parse() + viper.BindPFlags().
baseSettings, err := cryptoutilAppsTemplateServiceConfig.ParseWithFlagSet(fs, args, exitIfHelp)
if err != nil {
return nil, fmt.Errorf("failed to parse template settings: %w", err)
}

// Create skeleton-template settings from base template settings.
settings := &SkeletonTemplateServerSettings{
ServiceTemplateServerSettings: baseSettings,
}

// Override template defaults with skeleton-template specific values.
// Only override public port if user didn't explicitly specify one via CLI flag.
if !fs.Changed("bind-public-port") {
settings.BindPublicPort = cryptoutilSharedMagic.SkeletonTemplateServicePort
}

settings.OTLPService = cryptoutilSharedMagic.OTLPServiceSkeletonTemplate

// Validate skeleton-template specific settings.
if err := validateSettings(settings); err != nil {
return nil, fmt.Errorf("skeleton-template settings validation failed: %w", err)
}

return settings, nil
}

// Parse parses command line arguments and returns skeleton-template settings.
// Uses global pflag.CommandLine for backward compatibility.
func Parse(args []string, exitIfHelp bool) (*SkeletonTemplateServerSettings, error) {
return ParseWithFlagSet(pflag.CommandLine, args, exitIfHelp)
}

// validateSettings validates skeleton-template specific configuration.
func validateSettings(s *SkeletonTemplateServerSettings) error {
var validationErrors []string

if s.ServiceTemplateServerSettings == nil {
validationErrors = append(validationErrors, "base template settings cannot be nil")
}

if len(validationErrors) > 0 {
return fmt.Errorf("validation errors: %s", strings.Join(validationErrors, "; "))
}

return nil
}
