// Copyright (c) 2025 Justin Cranford

package server

import (
	"context"
	"fmt"

	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilTLSGenerator "cryptoutil/internal/shared/config/tls_generator"
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
	}

	if settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	}

	var (
		pubTLS *cryptoutilTLSGenerator.TLSGeneratedSettings
		err    error
	)

	switch settings.TLSPublicMode {
	case cryptoutilConfig.TLSModeStatic:
		pubTLS = &cryptoutilTLSGenerator.TLSGeneratedSettings{
			StaticCertPEM: settings.TLSStaticCertPEM,
			StaticKeyPEM:  settings.TLSStaticKeyPEM,
		}
	case cryptoutilConfig.TLSModeMixed:
		pubTLS, err = cryptoutilTLSGenerator.GenerateServerCertFromCA(settings.TLSMixedCACertPEM, settings.TLSMixedCAKeyPEM, settings.TLSPublicDNSNames, settings.TLSPublicIPAddresses, cryptoutilMagic.TLSTestEndEntityCertValidity1Year)
		if err != nil {
			return nil, fmt.Errorf("failed to generate public server cert from CA: %w", err)
		}
	case cryptoutilConfig.TLSModeAuto:
		pubTLS, err = cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(settings.TLSPublicDNSNames, settings.TLSPublicIPAddresses, cryptoutilMagic.TLSTestEndEntityCertValidity1Year)
		if err != nil {
			return nil, fmt.Errorf("failed to auto-generate public server certs: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown public TLS mode: %s", settings.TLSPublicMode)
	}

	var adminTLS *cryptoutilTLSGenerator.TLSGeneratedSettings

	switch settings.TLSPrivateMode {
	case cryptoutilConfig.TLSModeStatic:
		adminTLS = &cryptoutilTLSGenerator.TLSGeneratedSettings{
			StaticCertPEM: settings.TLSStaticCertPEM,
			StaticKeyPEM:  settings.TLSStaticKeyPEM,
		}
	case cryptoutilConfig.TLSModeMixed:
		adminTLS, err = cryptoutilTLSGenerator.GenerateServerCertFromCA(settings.TLSMixedCACertPEM, settings.TLSMixedCAKeyPEM, settings.TLSPrivateDNSNames, settings.TLSPrivateIPAddresses, cryptoutilMagic.TLSTestEndEntityCertValidity1Year)
		if err != nil {
			return nil, fmt.Errorf("failed to generate admin server cert from CA: %w", err)
		}
	case cryptoutilConfig.TLSModeAuto:
		adminTLS, err = cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(settings.TLSPrivateDNSNames, settings.TLSPrivateIPAddresses, cryptoutilMagic.TLSTestEndEntityCertValidity1Year)
		if err != nil {
			return nil, fmt.Errorf("failed to auto-generate admin server certs: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown admin TLS mode: %s", settings.TLSPrivateMode)
	}

	// Create servers
	publicServer, err := NewPublicHTTPServer(ctx, settings, pubTLS)
	if err != nil {
		return nil, fmt.Errorf("failed to create public server: %w", err)
	}

	adminServer, err := NewAdminHTTPServer(ctx, settings, adminTLS)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin server: %w", err)
	}

	return &HTTPSServers{
		Settings:     settings,
		PublicTLS:    pubTLS,
		AdminTLS:     adminTLS,
		PublicServer: publicServer,
		AdminServer:  adminServer,
	}, nil
}
