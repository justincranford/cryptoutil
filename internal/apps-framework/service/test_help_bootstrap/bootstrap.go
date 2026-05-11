// Copyright (c) 2025-2026 Justin Cranford.

// Package test_help_bootstrap provides configuration, environment, and startup wiring helpers
// for integration and E2E test suites. It handles config loading, environment variable setup,
// and bootstrap orchestration needed before starting test servers or compose stacks.
//
// Consumed by:
//   - test_orch_e2e: compose environment and config setup
//   - test_orch_integration: server startup config wiring
//   - Integration/E2E test suites: config loading and env setup
package test_help_bootstrap

import (
	"testing"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func cloneStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	return append([]string(nil), values...)
}

// NewTestServerSettings returns an isolated server settings instance suitable for
// test bootstrap code and parallel execution.
//
// The returned settings always use loopback + dynamic ports and auto TLS provisioning.
func NewTestServerSettings(t *testing.T) *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings {
	t.Helper()

	settings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	// Enforce explicit dynamic port binding on both listeners in tests.
	settings.BindPublicPort = 0
	settings.BindPrivatePort = 0

	// Ensure test helpers always run in auto TLS mode.
	settings.TLSPublicProvisionMode = cryptoutilAppsFrameworkServiceConfig.TLSProvisionModeAuto
	settings.TLSPrivateProvisionMode = cryptoutilAppsFrameworkServiceConfig.TLSProvisionModeAuto

	// Copy slices so callers can mutate settings without cross-test sharing.
	settings.TLSPublicDNSNames = cloneStringSlice(settings.TLSPublicDNSNames)
	settings.TLSPublicIPAddresses = cloneStringSlice(settings.TLSPublicIPAddresses)
	settings.TLSPrivateDNSNames = cloneStringSlice(settings.TLSPrivateDNSNames)
	settings.TLSPrivateIPAddresses = cloneStringSlice(settings.TLSPrivateIPAddresses)
	settings.CORSAllowedOrigins = cloneStringSlice(settings.CORSAllowedOrigins)
	settings.CORSAllowedMethods = cloneStringSlice(settings.CORSAllowedMethods)
	settings.CORSAllowedHeaders = cloneStringSlice(settings.CORSAllowedHeaders)

	return settings
}
