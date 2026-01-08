// Copyright (c) 2025 Justin Cranford

// Package e2e provides reusable end-to-end testing helpers for all cryptoutil services.
// Extracted from cipher-im implementation to support 9-service migration.
package e2e

import (
	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// NewTestServerSettings creates ServiceTemplateServerSettings with test-friendly defaults.
// Reusable for all services requiring ServiceTemplateServerSettings in tests.
//
// All bind addresses use 127.0.0.1 (loopback only) to prevent Windows Firewall prompts.
// All ports use 0 (dynamic allocation) to prevent port conflicts in parallel tests.
func NewTestServerSettings() *cryptoutilConfig.ServiceTemplateServerSettings {
	return &cryptoutilConfig.ServiceTemplateServerSettings{
		PublicBrowserAPIContextPath: cryptoutilMagic.DefaultPublicBrowserAPIContextPath,
		PublicServiceAPIContextPath: cryptoutilMagic.DefaultPublicServiceAPIContextPath,
		BindPublicProtocol:          cryptoutilMagic.ProtocolHTTPS,
		BindPublicAddress:           cryptoutilMagic.IPv4Loopback,
		BindPublicPort:              0, // Dynamic allocation
		BindPrivateProtocol:         cryptoutilMagic.ProtocolHTTPS,
		BindPrivateAddress:          cryptoutilMagic.IPv4Loopback,
		BindPrivatePort:             0, // Dynamic allocation
		TLSPublicDNSNames:           []string{cryptoutilMagic.HostnameLocalhost},
		TLSPublicIPAddresses:        []string{cryptoutilMagic.IPv4Loopback},
		TLSPrivateDNSNames:          []string{cryptoutilMagic.HostnameLocalhost},
		TLSPrivateIPAddresses:       []string{cryptoutilMagic.IPv4Loopback},
		CORSAllowedOrigins:          []string{},
		OTLPService:                 "test-service",
		OTLPEndpoint:                "grpc://localhost:4317",
		OTLPEnabled:                 false, // Disable actual OTLP export in tests
		LogLevel:                    "error",
		UnsealMode:                  cryptoutilMagic.DefaultUnsealModeSysInfo,
		// Session Manager settings - use OPAQUE for simplicity in tests (no JWK generation needed).
		BrowserSessionAlgorithm:    cryptoutilMagic.DefaultBrowserSessionAlgorithm,
		BrowserSessionJWSAlgorithm: cryptoutilMagic.DefaultBrowserSessionJWSAlgorithm,
		BrowserSessionJWEAlgorithm: cryptoutilMagic.DefaultBrowserSessionJWEAlgorithm,
		BrowserSessionExpiration:   cryptoutilMagic.DefaultBrowserSessionExpiration,
		ServiceSessionAlgorithm:    cryptoutilMagic.DefaultServiceSessionAlgorithm,
		ServiceSessionJWSAlgorithm: cryptoutilMagic.DefaultServiceSessionJWSAlgorithm,
		ServiceSessionJWEAlgorithm: cryptoutilMagic.DefaultServiceSessionJWEAlgorithm,
		ServiceSessionExpiration:   cryptoutilMagic.DefaultServiceSessionExpiration,
		SessionIdleTimeout:         cryptoutilMagic.DefaultSessionIdleTimeout,
		SessionCleanupInterval:     cryptoutilMagic.DefaultSessionCleanupInterval,
	}
}

// NewTestServerSettingsWithService creates ServiceTemplateServerSettings with custom service name.
// Useful when multiple service instances need distinct telemetry names.
func NewTestServerSettingsWithService(serviceName string) *cryptoutilConfig.ServiceTemplateServerSettings {
	settings := NewTestServerSettings()
	settings.OTLPService = serviceName

	return settings
}
