// Copyright (c) 2025 Justin Cranford

// Package config provides configuration management for cipher-im service.
package config

import (
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// NewTestConfig creates a CipherImServerSettings instance for testing without calling Parse().
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
// Returns directly populated CipherImServerSettings matching Parse() behavior.
func NewTestConfig(bindAddr string, bindPort uint16, devMode bool) *CipherImServerSettings {
	// Get base template config.
	baseConfig := cryptoutilAppsTemplateServiceConfig.NewTestConfig(bindAddr, bindPort, devMode)

	// Override template defaults with cipher-im specific values.
	baseConfig.BindPublicPort = bindPort
	baseConfig.OTLPService = cryptoutilSharedMagic.OTLPServiceCipherIM

	return &CipherImServerSettings{
		ServiceTemplateServerSettings: baseConfig,
		MessageJWEAlgorithm:           cryptoutilSharedMagic.CipherJWEAlgorithm,
		MessageMinLength:              cryptoutilSharedMagic.CipherMessageMinLength,
		MessageMaxLength:              cryptoutilSharedMagic.CipherMessageMaxLength,
		RecipientsMinCount:            cryptoutilSharedMagic.CipherRecipientsMinCount,
		RecipientsMaxCount:            cryptoutilSharedMagic.CipherRecipientsMaxCount,
	}
}

// DefaultTestConfig creates a default test configuration suitable for most unit tests.
// Uses loopback address, dynamic port allocation, and dev mode.
func DefaultTestConfig() *CipherImServerSettings {
	return NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
}
