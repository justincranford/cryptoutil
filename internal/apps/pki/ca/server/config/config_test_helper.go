// Copyright (c) 2025 Justin Cranford
//

// Package config provides configuration management for pki-ca service.
package config

import (
cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// NewTestConfig creates a PKICAServerSettings instance for testing without calling Parse().
// This bypasses pflag's global FlagSet to allow multiple config creations in tests.
//
// Use this in tests instead of Parse() to avoid "flag redefined" panics
// when creating multiple server instances.
//
// Parameters:
//   - bindAddr: public bind address (typically cryptoutilSharedMagic.IPv4Loopback).
//   - bindPort: public bind port (use 0 for dynamic allocation).
//   - devMode: enable development mode (in-memory SQLite, relaxed security).
//
// Returns directly populated PKICAServerSettings matching Parse() behavior.
func NewTestConfig(bindAddr string, bindPort uint16, devMode bool) *PKICAServerSettings {
// Get base template config.
baseConfig := cryptoutilAppsTemplateServiceConfig.NewTestConfig(bindAddr, bindPort, devMode)

// Override template defaults with pki-ca specific values.
baseConfig.BindPublicPort = bindPort
baseConfig.OTLPService = cryptoutilSharedMagic.OTLPServicePKICA

return &PKICAServerSettings{
ServiceTemplateServerSettings: baseConfig,
}
}

// DefaultTestConfig creates a default test configuration suitable for most unit tests.
// Uses loopback address, dynamic port allocation, and dev mode.
func DefaultTestConfig() *PKICAServerSettings {
return NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
}
