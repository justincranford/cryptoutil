// Copyright (c) 2025 Justin Cranford

package listener

import (
	"context"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceConfigTlsGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const testUnknownTLSMode = "unknown"

func TestNewHTTPServers_AutoMode_HappyPath(t *testing.T) {
	t.Parallel()

	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	ctx := context.Background()
	h, err := NewHTTPServers(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, h)
	require.NotNil(t, h.PublicServer)
	require.NotNil(t, h.AdminServer)
}

func TestNewHTTPServers_NilContext(t *testing.T) {
	t.Parallel()

	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	h, err := NewHTTPServers(nil, settings) //nolint:staticcheck // Testing nil context handling.
	require.Error(t, err)
	require.Contains(t, err.Error(), "context cannot be nil")
	require.Nil(t, h)
}

func TestNewHTTPServers_NilSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	h, err := NewHTTPServers(ctx, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "settings cannot be nil")
	require.Nil(t, h)
}

func TestNewHTTPServers_StaticMode_HappyPath(t *testing.T) {
	t.Parallel()
	// Generate static certs first using auto mode.
	staticTLS, err := cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault},
		[]string{cryptoutilSharedMagic.IPv4Loopback},
		cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year,
	)
	require.NoError(t, err)

	// Create settings with static TLS mode.
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	settings.TLSPublicMode = cryptoutilAppsTemplateServiceConfig.TLSModeStatic
	settings.TLSPrivateMode = cryptoutilAppsTemplateServiceConfig.TLSModeStatic
	settings.TLSStaticCertPEM = staticTLS.StaticCertPEM
	settings.TLSStaticKeyPEM = staticTLS.StaticKeyPEM

	ctx := context.Background()
	h, err := NewHTTPServers(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, h)
	require.NotNil(t, h.PublicServer)
	require.NotNil(t, h.AdminServer)
}

func TestNewHTTPServers_UnknownPublicTLSMode(t *testing.T) {
	t.Parallel()

	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	settings.TLSPublicMode = testUnknownTLSMode

	ctx := context.Background()
	h, err := NewHTTPServers(ctx, settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown public TLS mode")
	require.Nil(t, h)
}

func TestNewHTTPServers_UnknownPrivateTLSMode(t *testing.T) {
	t.Parallel()

	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	settings.TLSPrivateMode = testUnknownTLSMode

	ctx := context.Background()
	h, err := NewHTTPServers(ctx, settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown admin TLS mode")
	require.Nil(t, h)
}

func TestNewHTTPServers_MixedMode_HappyPath(t *testing.T) {
	t.Parallel()
	// First, generate a CA to use for mixed mode.
	caCertPEM, caKeyPEM, err := cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateTestCA()
	require.NoError(t, err)

	// Create settings with mixed TLS mode using the generated CA.
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	settings.TLSPublicMode = cryptoutilAppsTemplateServiceConfig.TLSModeMixed
	settings.TLSPrivateMode = cryptoutilAppsTemplateServiceConfig.TLSModeMixed
	settings.TLSMixedCACertPEM = caCertPEM
	settings.TLSMixedCAKeyPEM = caKeyPEM
	settings.TLSPublicDNSNames = []string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}
	settings.TLSPublicIPAddresses = []string{cryptoutilSharedMagic.IPv4Loopback}
	settings.TLSPrivateDNSNames = []string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}
	settings.TLSPrivateIPAddresses = []string{cryptoutilSharedMagic.IPv4Loopback}

	ctx := context.Background()
	h, err := NewHTTPServers(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, h)
	require.NotNil(t, h.PublicServer)
	require.NotNil(t, h.AdminServer)
}

func TestNewHTTPServers_MixedMode_InvalidPublicCA(t *testing.T) {
	t.Parallel()
	// Create settings with mixed TLS mode but invalid CA cert.
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	settings.TLSPublicMode = cryptoutilAppsTemplateServiceConfig.TLSModeMixed
	settings.TLSMixedCACertPEM = []byte("invalid-ca-cert")
	settings.TLSMixedCAKeyPEM = []byte("invalid-ca-key")
	settings.TLSPublicDNSNames = []string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}
	settings.TLSPublicIPAddresses = []string{cryptoutilSharedMagic.IPv4Loopback}

	ctx := context.Background()
	h, err := NewHTTPServers(ctx, settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate public server cert from CA")
	require.Nil(t, h)
}

func TestNewHTTPServers_MixedMode_InvalidPrivateCA(t *testing.T) {
	t.Parallel()
	// Create settings with auto for public, mixed for private with invalid CA.
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	settings.TLSPublicMode = cryptoutilAppsTemplateServiceConfig.TLSModeAuto
	settings.TLSPrivateMode = cryptoutilAppsTemplateServiceConfig.TLSModeMixed
	settings.TLSMixedCACertPEM = []byte("invalid-ca-cert")
	settings.TLSMixedCAKeyPEM = []byte("invalid-ca-key")
	settings.TLSPrivateDNSNames = []string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}
	settings.TLSPrivateIPAddresses = []string{cryptoutilSharedMagic.IPv4Loopback}

	ctx := context.Background()
	h, err := NewHTTPServers(ctx, settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate admin server cert from CA")
	require.Nil(t, h)
}

// ========================
// Dual HTTPS Server Integration Tests
// ========================

// TestDualServers_StartBothServers tests that both public and admin servers can start concurrently.
// DELETED: TestDualServers_StartBothServers - violated copilot instructions (real HTTPS listener).
// Coverage provided by NewHTTPServers constructor tests (port allocation, error paths).

// DELETED: TestDualServers_HealthEndpoints - violated copilot instructions (real HTTPS requests).
// Coverage provided by application/application_listener_test.go using app.Test() pattern.

// DELETED: TestDualServers_GracefulShutdown - violated copilot instructions (real network shutdown).
// Coverage provided by graceful shutdown tests using app.Test() pattern.

// DELETED: TestDualServers_BothServersAccessibleSimultaneously - violated copilot instructions (real concurrent HTTPS).
