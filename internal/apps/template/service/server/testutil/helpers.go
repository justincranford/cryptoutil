// Copyright (c) 2025 Justin Cranford

package testutil

import (
	"fmt"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceConfigTlsGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	serverSettings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
	publicTLS      *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings
	privateTLS     *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings
)

// Initialize is called from TestMain to setup shared test fixtures.
func Initialize() error {
	// Create shared ServiceTemplateServerSettings fixture for tests (port 0 for dynamic allocation).
	serverSettings = &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BindPublicProtocol:          "https",
		BindPublicAddress:           cryptoutilSharedMagic.IPv4Loopback,
		BindPublicPort:              0,
		BindPrivateProtocol:         "https",
		BindPrivateAddress:          cryptoutilSharedMagic.IPv4Loopback,
		BindPrivatePort:             0,
		PublicBrowserAPIContextPath: "/browser",
		PublicServiceAPIContextPath: "/service",
		PrivateAdminAPIContextPath:  "/admin",
		TLSPublicDNSNames:           []string{"localhost"},
		TLSPublicIPAddresses:        []string{cryptoutilSharedMagic.IPv4Loopback},
		TLSPrivateDNSNames:          []string{"localhost"},
		TLSPrivateIPAddresses:       []string{cryptoutilSharedMagic.IPv4Loopback},
		TLSPublicMode:               cryptoutilAppsTemplateServiceConfig.TLSModeAuto,
		TLSPrivateMode:              cryptoutilAppsTemplateServiceConfig.TLSModeAuto,
	}

	// Generate shared TLS fixtures for tests (auto-mode, localhost/IPs).
	var err error

	publicTLS, err = cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings([]string{"localhost"}, []string{cryptoutilSharedMagic.IPv4Loopback}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	if err != nil {
		return fmt.Errorf("failed to generate public TLS settings: %w", err)
	}

	privateTLS, err = cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings([]string{"localhost"}, []string{cryptoutilSharedMagic.IPv4Loopback}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	if err != nil {
		return fmt.Errorf("failed to generate private TLS settings: %w", err)
	}

	return nil
}

// ServiceTemplateServerSettings returns the shared test ServiceTemplateServerSettings fixture.
func ServiceTemplateServerSettings() *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings {
	return serverSettings
}

// PublicTLS returns the shared test public TLS fixture.
func PublicTLS() *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings {
	return publicTLS
}

// PrivateTLS returns the shared test private TLS fixture.
func PrivateTLS() *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings {
	return privateTLS
}
