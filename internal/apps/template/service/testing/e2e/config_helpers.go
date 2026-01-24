// Copyright (c) 2025 Justin Cranford

// Package e2e provides reusable end-to-end testing helpers for all cryptoutil services.
// Extracted from cipher-im implementation to support 9-service migration.
package e2e

import (
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// NewTestServerSettings creates ServiceTemplateServerSettings with test-friendly defaults.
// Reusable for all services requiring ServiceTemplateServerSettings in tests.
//
// All bind addresses use 127.0.0.1 (loopback only) to prevent Windows Firewall prompts.
// All ports use 0 (dynamic allocation) to prevent port conflicts in parallel tests.
// DevMode is enabled to use random unseal key (avoids sysinfo collection that can timeout).
func NewTestServerSettings() *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings {
	return &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:                     cryptoutilSharedMagic.TestDefaultDevMode, // Use random unseal key, avoids sysinfo timeout
		PublicBrowserAPIContextPath: cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath,
		PublicServiceAPIContextPath: cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath,
		BindPublicProtocol:          cryptoutilSharedMagic.ProtocolHTTPS,
		BindPublicAddress:           cryptoutilSharedMagic.IPv4Loopback,
		BindPublicPort:              0, // Dynamic allocation
		BindPrivateProtocol:         cryptoutilSharedMagic.ProtocolHTTPS,
		BindPrivateAddress:          cryptoutilSharedMagic.IPv4Loopback,
		BindPrivatePort:             0, // Dynamic allocation
		TLSPublicDNSNames:           []string{cryptoutilSharedMagic.HostnameLocalhost},
		TLSPublicIPAddresses:        []string{cryptoutilSharedMagic.IPv4Loopback},
		TLSPrivateDNSNames:          []string{cryptoutilSharedMagic.HostnameLocalhost},
		TLSPrivateIPAddresses:       []string{cryptoutilSharedMagic.IPv4Loopback},
		CORSAllowedOrigins:          []string{},
		OTLPService:                 "test-service",
		OTLPEndpoint:                "grpc://localhost:4317",
		OTLPEnabled:                 false, // Disable actual OTLP export in tests
		LogLevel:                    "error",
		UnsealMode:                  cryptoutilSharedMagic.DefaultUnsealModeSysInfo, // Ignored when DevMode=true
		// Session Manager settings - use OPAQUE for simplicity in tests (no JWK generation needed).
		BrowserSessionAlgorithm:    cryptoutilSharedMagic.DefaultBrowserSessionAlgorithm,
		BrowserSessionJWSAlgorithm: cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		BrowserSessionJWEAlgorithm: cryptoutilSharedMagic.DefaultBrowserSessionJWEAlgorithm,
		BrowserSessionExpiration:   cryptoutilSharedMagic.DefaultBrowserSessionExpiration,
		ServiceSessionAlgorithm:    cryptoutilSharedMagic.DefaultServiceSessionAlgorithm,
		ServiceSessionJWSAlgorithm: cryptoutilSharedMagic.DefaultServiceSessionJWSAlgorithm,
		ServiceSessionJWEAlgorithm: cryptoutilSharedMagic.DefaultServiceSessionJWEAlgorithm,
		ServiceSessionExpiration:   cryptoutilSharedMagic.DefaultServiceSessionExpiration,
		SessionIdleTimeout:         cryptoutilSharedMagic.DefaultSessionIdleTimeout,
		SessionCleanupInterval:     cryptoutilSharedMagic.DefaultSessionCleanupInterval,
	}
}

// NewTestServerSettingsWithService creates ServiceTemplateServerSettings with custom service name.
// Useful when multiple service instances need distinct telemetry names.
func NewTestServerSettingsWithService(serviceName string) *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings {
	settings := NewTestServerSettings()
	settings.OTLPService = serviceName

	return settings
}
