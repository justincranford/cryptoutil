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
		BindPublicProtocol:          cryptoutilSharedMagic.ProtocolHTTPS,
		BindPublicAddress:           cryptoutilSharedMagic.IPv4Loopback,
		BindPublicPort:              0,
		BindPrivateProtocol:         cryptoutilSharedMagic.ProtocolHTTPS,
		BindPrivateAddress:          cryptoutilSharedMagic.IPv4Loopback,
		BindPrivatePort:             0,
		PublicBrowserAPIContextPath: cryptoutilSharedMagic.PathPrefixBrowser,
		PublicServiceAPIContextPath: cryptoutilSharedMagic.PathPrefixService,
		PrivateAdminAPIContextPath:  "/admin",
		TLSPublicDNSNames:           []string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault},
		TLSPublicIPAddresses:        []string{cryptoutilSharedMagic.IPv4Loopback},
		TLSPrivateDNSNames:          []string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault},
		TLSPrivateIPAddresses:       []string{cryptoutilSharedMagic.IPv4Loopback},
		TLSPublicMode:               cryptoutilAppsTemplateServiceConfig.TLSModeAuto,
		TLSPrivateMode:              cryptoutilAppsTemplateServiceConfig.TLSModeAuto,
	}

	// Generate shared TLS fixtures for tests (auto-mode, localhost/IPs).
	var err error

	publicTLS, err = cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings([]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, []string{cryptoutilSharedMagic.IPv4Loopback}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	if err != nil {
		return fmt.Errorf("failed to generate public TLS settings: %w", err)
	}

	privateTLS, err = cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings([]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, []string{cryptoutilSharedMagic.IPv4Loopback}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	if err != nil {
		return fmt.Errorf("failed to generate private TLS settings: %w", err)
	}

	return nil
}

// ServiceTemplateServerSettings returns a *copy* of the test ServiceTemplateServerSettings fixture.
// Returns a copy to prevent race conditions when parallel tests modify settings.
func ServiceTemplateServerSettings() *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings {
	// Deep copy to prevent race conditions when tests modify settings.
	settingsCopy := *serverSettings

	// Deep copy slices to prevent shared slice modifications.
	settingsCopy.TLSPublicDNSNames = make([]string, len(serverSettings.TLSPublicDNSNames))
	settingsCopy.TLSPublicIPAddresses = make([]string, len(serverSettings.TLSPublicIPAddresses))
	settingsCopy.TLSPrivateDNSNames = make([]string, len(serverSettings.TLSPrivateDNSNames))
	settingsCopy.TLSPrivateIPAddresses = make([]string, len(serverSettings.TLSPrivateIPAddresses))

	copy(settingsCopy.TLSPublicDNSNames, serverSettings.TLSPublicDNSNames)
	copy(settingsCopy.TLSPublicIPAddresses, serverSettings.TLSPublicIPAddresses)
	copy(settingsCopy.TLSPrivateDNSNames, serverSettings.TLSPrivateDNSNames)
	copy(settingsCopy.TLSPrivateIPAddresses, serverSettings.TLSPrivateIPAddresses)

	return &settingsCopy
}

// PublicTLS returns the shared test public TLS fixture.
func PublicTLS() *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings {
	return publicTLS
}

// PrivateTLS returns the shared test private TLS fixture.
func PrivateTLS() *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings {
	return privateTLS
}
