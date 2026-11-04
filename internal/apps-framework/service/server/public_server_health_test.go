// Copyright (c) 2025-2026 Justin Cranford.
package server_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceConfigTlsGenerator "cryptoutil/internal/apps-framework/service/config/tls_generator"
	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps-framework/service/server"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const testPublicServerDNSName = cryptoutilSharedMagic.DefaultOTLPHostnameDefault

func TestPublicServerCheckResponsiveUsesVerifiedLoopback(t *testing.T) {
	t.Parallel()

	tlsSettings, err := cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{testPublicServerDNSName},
		[]string{cryptoutilSharedMagic.IPv4Loopback},
		cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year,
	)
	require.NoError(t, err)
	tlsMaterial, err := cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateTLSMaterial(tlsSettings)
	require.NoError(t, err)

	server, err := cryptoutilAppsFrameworkServiceServer.NewPublicServerBase(&cryptoutilAppsFrameworkServiceServer.PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        0,
		TLSMaterial: tlsMaterial,
	})
	require.NoError(t, err)

	select {
	case <-server.Started():
		require.FailNow(t, "startup notification must remain open before bind")
	default:
	}

	ctx, cancel := context.WithTimeout(t.Context(), cryptoutilSharedMagic.TestTimeout)
	defer cancel()

	startResult := make(chan error, 1)
	go func() {
		startResult <- server.Start(ctx)
	}()

	select {
	case <-server.Started():
		require.Positive(t, server.ActualPort(), "notification must follow successful bind")
	case startErr := <-startResult:
		require.NoError(t, startErr, "server must bind before returning")
		require.FailNow(t, "server returned without a startup notification")
	case <-ctx.Done():
		require.NoError(t, ctx.Err(), "server startup must finish within the test deadline")
	}

	require.NoError(t, server.CheckResponsive(ctx))

	cancel()
	require.Error(t, <-startResult)
}

func TestPublicServerCheckResponsiveRejectsUnboundListener(t *testing.T) {
	t.Parallel()

	settings, err := cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{testPublicServerDNSName},
		[]string{cryptoutilSharedMagic.IPv4Loopback},
		cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year,
	)
	require.NoError(t, err)
	material, err := cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateTLSMaterial(settings)
	require.NoError(t, err)

	server, err := cryptoutilAppsFrameworkServiceServer.NewPublicServerBase(&cryptoutilAppsFrameworkServiceServer.PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		TLSMaterial: material,
	})
	require.NoError(t, err)

	require.ErrorContains(t, server.CheckResponsive(t.Context()), "not bound")
}
