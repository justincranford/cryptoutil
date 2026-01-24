// Copyright (c) 2025 Justin Cranford

package listener

import (
	"context"
	"fmt"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceConfigTlsGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// HTTPSServers is a convenience wrapper for public/admin HTTPS servers and their settings.
type HTTPSServers struct {
	Settings     *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
	PublicTLS    *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings
	AdminTLS     *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings
	PublicServer *PublicHTTPServer
	AdminServer  *AdminServer
}

// NewHTTPServers creates public and admin HTTPS servers using provided ServiceTemplateServerSettings.
// It will generate TLS material based on TLSPublicMode/TLSPrivateMode and return
// an HTTPSServers wrapper containing servers and TLS configs.
func NewHTTPServers(ctx context.Context, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) (*HTTPSServers, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	}

	publicTLSGeneratedSettings, err := createPublicTLSGeneratedSettings(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create public TLS settings: %w", err)
	}

	adminTLSGeneratedSettings, err := createAdminTLSGeneratedSettings(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin TLS settings: %w", err)
	}

	// Create servers
	publicServer, err := NewPublicHTTPServer(ctx, settings, publicTLSGeneratedSettings)
	if err != nil {
		return nil, fmt.Errorf("failed to create public server: %w", err)
	}

	adminServer, err := NewAdminHTTPServer(ctx, settings, adminTLSGeneratedSettings)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin server: %w", err)
	}

	return &HTTPSServers{
		Settings:     settings,
		PublicTLS:    publicTLSGeneratedSettings,
		AdminTLS:     adminTLSGeneratedSettings,
		PublicServer: publicServer,
		AdminServer:  adminServer,
	}, nil
}

func createPublicTLSGeneratedSettings(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) (*cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings, error) {
	var publicTLSGeneratedSettings *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings

	switch settings.TLSPublicMode {
	case cryptoutilAppsTemplateServiceConfig.TLSModeStatic:
		publicTLSGeneratedSettings = &cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings{
			StaticCertPEM: settings.TLSStaticCertPEM,
			StaticKeyPEM:  settings.TLSStaticKeyPEM,
		}
	case cryptoutilAppsTemplateServiceConfig.TLSModeMixed:
		var err error

		publicTLSGeneratedSettings, err = cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateServerCertFromCA(settings.TLSMixedCACertPEM, settings.TLSMixedCAKeyPEM, settings.TLSPublicDNSNames, settings.TLSPublicIPAddresses, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
		if err != nil {
			return nil, fmt.Errorf("failed to generate public server cert from CA: %w", err)
		}
	case cryptoutilAppsTemplateServiceConfig.TLSModeAuto:
		var err error

		publicTLSGeneratedSettings, err = cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings(settings.TLSPublicDNSNames, settings.TLSPublicIPAddresses, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
		if err != nil {
			return nil, fmt.Errorf("failed to auto-generate public server certs: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown public TLS mode: %s", settings.TLSPublicMode)
	}

	return publicTLSGeneratedSettings, nil
}

func createAdminTLSGeneratedSettings(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) (*cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings, error) {
	var adminTLSGeneratedSettings *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings

	switch settings.TLSPrivateMode {
	case cryptoutilAppsTemplateServiceConfig.TLSModeStatic:
		adminTLSGeneratedSettings = &cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings{
			StaticCertPEM: settings.TLSStaticCertPEM,
			StaticKeyPEM:  settings.TLSStaticKeyPEM,
		}
	case cryptoutilAppsTemplateServiceConfig.TLSModeMixed:
		var err error

		adminTLSGeneratedSettings, err = cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateServerCertFromCA(settings.TLSMixedCACertPEM, settings.TLSMixedCAKeyPEM, settings.TLSPrivateDNSNames, settings.TLSPrivateIPAddresses, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
		if err != nil {
			return nil, fmt.Errorf("failed to generate admin server cert from CA: %w", err)
		}
	case cryptoutilAppsTemplateServiceConfig.TLSModeAuto:
		var err error

		adminTLSGeneratedSettings, err = cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings(settings.TLSPrivateDNSNames, settings.TLSPrivateIPAddresses, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
		if err != nil {
			return nil, fmt.Errorf("failed to auto-generate admin server certs: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown admin TLS mode: %s", settings.TLSPrivateMode)
	}

	return adminTLSGeneratedSettings, nil
}
