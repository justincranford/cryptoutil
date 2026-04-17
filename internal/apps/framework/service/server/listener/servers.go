// Copyright (c) 2025 Justin Cranford

package listener

import (
	"context"
	"fmt"
	"net"
	"os"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilAppsFrameworkServiceConfigTlsGenerator "cryptoutil/internal/apps/framework/service/config/tls_generator"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// HTTPSServers is a convenience wrapper for public/admin HTTPS servers and their settings.
type HTTPSServers struct {
	Settings     *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings
	PublicTLS    *cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings
	AdminTLS     *cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings
	PublicServer *PublicHTTPServer
	AdminServer  *AdminServer
}

// NewHTTPServers creates public and admin HTTPS servers using provided ServiceFrameworkServerSettings.
// It will generate TLS material based on TLSPublicMode/TLSPrivateMode and return
// an HTTPSServers wrapper containing servers and TLS configs.
func NewHTTPServers(ctx context.Context, settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*HTTPSServers, error) {
	return newHTTPServersInternal(ctx, settings, cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateTLSMaterial)
}

func newHTTPServersInternal(
	ctx context.Context,
	settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings,
	generateTLSMaterialFn func(cfg *cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings) (*cryptoutilAppsFrameworkServiceConfig.TLSMaterial, error),
) (*HTTPSServers, error) {
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
	publicServer, err := newPublicHTTPServerInternal(ctx, settings, publicTLSGeneratedSettings, generateTLSMaterialFn,
		func(ctx context.Context, network, address string) (net.Listener, error) {
			return (&net.ListenConfig{}).Listen(ctx, network, address)
		},
		func(app *fiber.App, ln net.Listener) error {
			return app.Listener(ln) //nolint:wrapcheck // Pass-through to Fiber framework.
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create public server: %w", err)
	}

	adminServer, err := newAdminHTTPServerInternal(ctx, settings, adminTLSGeneratedSettings, generateTLSMaterialFn,
		func(ctx context.Context, network, address string) (net.Listener, error) {
			return (&net.ListenConfig{}).Listen(ctx, network, address)
		},
		func(app *fiber.App, ln net.Listener) error {
			//nolint:wrapcheck // Pass-through to Fiber framework.
			return app.Listener(ln)
		},
		os.ReadFile,
	)
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

func createPublicTLSGeneratedSettings(settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings, error) {
	var publicTLSGeneratedSettings *cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings

	switch settings.TLSPublicMode {
	case cryptoutilAppsFrameworkServiceConfig.TLSModeStatic:
		publicTLSGeneratedSettings = &cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings{
			StaticCertPEM: settings.TLSStaticCertPEM,
			StaticKeyPEM:  settings.TLSStaticKeyPEM,
		}
	case cryptoutilAppsFrameworkServiceConfig.TLSModeMixed:
		var err error

		publicTLSGeneratedSettings, err = cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateServerCertFromCA(settings.TLSMixedCACertPEM, settings.TLSMixedCAKeyPEM, settings.TLSPublicDNSNames, settings.TLSPublicIPAddresses, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
		if err != nil {
			return nil, fmt.Errorf("failed to generate public server cert from CA: %w", err)
		}
	case cryptoutilAppsFrameworkServiceConfig.TLSModeAuto:
		var err error

		publicTLSGeneratedSettings, err = cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings(settings.TLSPublicDNSNames, settings.TLSPublicIPAddresses, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
		if err != nil {
			return nil, fmt.Errorf("failed to auto-generate public server certs: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown public TLS mode: %s", settings.TLSPublicMode)
	}

	return publicTLSGeneratedSettings, nil
}

func createAdminTLSGeneratedSettings(settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings, error) {
	var adminTLSGeneratedSettings *cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings

	switch settings.TLSPrivateMode {
	case cryptoutilAppsFrameworkServiceConfig.TLSModeStatic:
		adminTLSGeneratedSettings = &cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings{
			StaticCertPEM: settings.TLSStaticCertPEM,
			StaticKeyPEM:  settings.TLSStaticKeyPEM,
		}
	case cryptoutilAppsFrameworkServiceConfig.TLSModeMixed:
		var err error

		adminTLSGeneratedSettings, err = cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateServerCertFromCA(settings.TLSMixedCACertPEM, settings.TLSMixedCAKeyPEM, settings.TLSPrivateDNSNames, settings.TLSPrivateIPAddresses, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
		if err != nil {
			return nil, fmt.Errorf("failed to generate admin server cert from CA: %w", err)
		}
	case cryptoutilAppsFrameworkServiceConfig.TLSModeAuto:
		var err error

		adminTLSGeneratedSettings, err = cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings(settings.TLSPrivateDNSNames, settings.TLSPrivateIPAddresses, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
		if err != nil {
			return nil, fmt.Errorf("failed to auto-generate admin server certs: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown admin TLS mode: %s", settings.TLSPrivateMode)
	}

	return adminTLSGeneratedSettings, nil
}
