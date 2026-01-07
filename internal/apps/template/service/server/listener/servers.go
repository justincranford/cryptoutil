// Copyright (c) 2025 Justin Cranford

package listener

import (
	"context"
	"fmt"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTLSGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// HTTPSServers is a convenience wrapper for public/admin HTTPS servers and their settings.
type HTTPSServers struct {
	Settings     *cryptoutilConfig.ServerSettings
	PublicTLS    *cryptoutilTLSGenerator.TLSGeneratedSettings
	AdminTLS     *cryptoutilTLSGenerator.TLSGeneratedSettings
	PublicServer *PublicHTTPServer
	AdminServer  *AdminServer
}

// NewHTTPServers creates public and admin HTTPS servers using provided ServerSettings.
// It will generate TLS material based on TLSPublicMode/TLSPrivateMode and return
// an HTTPSServers wrapper containing servers and TLS configs.
func NewHTTPServers(ctx context.Context, settings *cryptoutilConfig.ServerSettings) (*HTTPSServers, error) {
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

func createPublicTLSGeneratedSettings(settings *cryptoutilConfig.ServerSettings) (*cryptoutilTLSGenerator.TLSGeneratedSettings, error) {
	var publicTLSGeneratedSettings *cryptoutilTLSGenerator.TLSGeneratedSettings

	switch settings.TLSPublicMode {
	case cryptoutilConfig.TLSModeStatic:
		publicTLSGeneratedSettings = &cryptoutilTLSGenerator.TLSGeneratedSettings{
			StaticCertPEM: settings.TLSStaticCertPEM,
			StaticKeyPEM:  settings.TLSStaticKeyPEM,
		}
	case cryptoutilConfig.TLSModeMixed:
		var err error

		publicTLSGeneratedSettings, err = cryptoutilTLSGenerator.GenerateServerCertFromCA(settings.TLSMixedCACertPEM, settings.TLSMixedCAKeyPEM, settings.TLSPublicDNSNames, settings.TLSPublicIPAddresses, cryptoutilMagic.TLSTestEndEntityCertValidity1Year)
		if err != nil {
			return nil, fmt.Errorf("failed to generate public server cert from CA: %w", err)
		}
	case cryptoutilConfig.TLSModeAuto:
		var err error

		publicTLSGeneratedSettings, err = cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(settings.TLSPublicDNSNames, settings.TLSPublicIPAddresses, cryptoutilMagic.TLSTestEndEntityCertValidity1Year)
		if err != nil {
			return nil, fmt.Errorf("failed to auto-generate public server certs: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown public TLS mode: %s", settings.TLSPublicMode)
	}

	return publicTLSGeneratedSettings, nil
}

func createAdminTLSGeneratedSettings(settings *cryptoutilConfig.ServerSettings) (*cryptoutilTLSGenerator.TLSGeneratedSettings, error) {
	var adminTLSGeneratedSettings *cryptoutilTLSGenerator.TLSGeneratedSettings

	switch settings.TLSPrivateMode {
	case cryptoutilConfig.TLSModeStatic:
		adminTLSGeneratedSettings = &cryptoutilTLSGenerator.TLSGeneratedSettings{
			StaticCertPEM: settings.TLSStaticCertPEM,
			StaticKeyPEM:  settings.TLSStaticKeyPEM,
		}
	case cryptoutilConfig.TLSModeMixed:
		var err error

		adminTLSGeneratedSettings, err = cryptoutilTLSGenerator.GenerateServerCertFromCA(settings.TLSMixedCACertPEM, settings.TLSMixedCAKeyPEM, settings.TLSPrivateDNSNames, settings.TLSPrivateIPAddresses, cryptoutilMagic.TLSTestEndEntityCertValidity1Year)
		if err != nil {
			return nil, fmt.Errorf("failed to generate admin server cert from CA: %w", err)
		}
	case cryptoutilConfig.TLSModeAuto:
		var err error

		adminTLSGeneratedSettings, err = cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(settings.TLSPrivateDNSNames, settings.TLSPrivateIPAddresses, cryptoutilMagic.TLSTestEndEntityCertValidity1Year)
		if err != nil {
			return nil, fmt.Errorf("failed to auto-generate admin server certs: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown admin TLS mode: %s", settings.TLSPrivateMode)
	}

	return adminTLSGeneratedSettings, nil
}
