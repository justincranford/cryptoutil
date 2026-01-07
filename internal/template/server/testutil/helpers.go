// Copyright (c) 2025 Justin Cranford

package testutil

import (
	"fmt"

	cryptoutilConfig "cryptoutil/internal/template/config"
	cryptoutilTLSGenerator "cryptoutil/internal/template/config/tls_generator"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

var (
	serverSettings *cryptoutilConfig.ServerSettings
	publicTLS      *cryptoutilTLSGenerator.TLSGeneratedSettings
	privateTLS     *cryptoutilTLSGenerator.TLSGeneratedSettings
)

// Initialize is called from TestMain to setup shared test fixtures.
func Initialize() error {
	// Create shared ServerSettings fixture for tests (port 0 for dynamic allocation).
	serverSettings = &cryptoutilConfig.ServerSettings{
		BindPublicProtocol:          "https",
		BindPublicAddress:           cryptoutilMagic.IPv4Loopback,
		BindPublicPort:              0,
		BindPrivateProtocol:         "https",
		BindPrivateAddress:          cryptoutilMagic.IPv4Loopback,
		BindPrivatePort:             0,
		PublicBrowserAPIContextPath: "/browser",
		PublicServiceAPIContextPath: "/service",
		PrivateAdminAPIContextPath:  "/admin",
		TLSPublicDNSNames:           []string{"localhost"},
		TLSPublicIPAddresses:        []string{cryptoutilMagic.IPv4Loopback},
		TLSPrivateDNSNames:          []string{"localhost"},
		TLSPrivateIPAddresses:       []string{cryptoutilMagic.IPv4Loopback},
		TLSPublicMode:               cryptoutilConfig.TLSModeAuto,
		TLSPrivateMode:              cryptoutilConfig.TLSModeAuto,
	}

	// Generate shared TLS fixtures for tests (auto-mode, localhost/IPs).
	var err error

	publicTLS, err = cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings([]string{"localhost"}, []string{cryptoutilMagic.IPv4Loopback}, cryptoutilMagic.TLSTestEndEntityCertValidity1Year)
	if err != nil {
		return fmt.Errorf("failed to generate public TLS settings: %w", err)
	}

	privateTLS, err = cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings([]string{"localhost"}, []string{cryptoutilMagic.IPv4Loopback}, cryptoutilMagic.TLSTestEndEntityCertValidity1Year)
	if err != nil {
		return fmt.Errorf("failed to generate private TLS settings: %w", err)
	}

	return nil
}

// ServerSettings returns the shared test ServerSettings fixture.
func ServerSettings() *cryptoutilConfig.ServerSettings {
	return serverSettings
}

// PublicTLS returns the shared test public TLS fixture.
func PublicTLS() *cryptoutilTLSGenerator.TLSGeneratedSettings {
	return publicTLS
}

// PrivateTLS returns the shared test private TLS fixture.
func PrivateTLS() *cryptoutilTLSGenerator.TLSGeneratedSettings {
	return privateTLS
}
