// Copyright (c) 2025 Justin Cranford

package listener

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceConfigTlsGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var errForcedTLSMaterialFailure = errors.New("forced TLS material failure")

const testInvalidBindAddress = "999.999.999.999"

func validAutoTLSSettings(t *testing.T) *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings {
	t.Helper()

	tlsCfg, err := cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{"localhost"},
		[]string{cryptoutilSharedMagic.IPv4Loopback},
		cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year,
	)
	require.NoError(t, err)

	return tlsCfg
}

// === TLS Material Generation Failure (admin.go:55, public.go:64) ===

func TestNewAdminHTTPServer_TLSMaterialGenFailure(t *testing.T) {
	orig := generateTLSMaterialFn
	generateTLSMaterialFn = func(_ *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings) (*cryptoutilAppsTemplateServiceConfig.TLSMaterial, error) {
		return nil, errForcedTLSMaterialFailure
	}

	defer func() { generateTLSMaterialFn = orig }()

	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	tlsCfg := &cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings{}

	server, err := NewAdminHTTPServer(context.Background(), settings, tlsCfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate TLS material")
	assert.Nil(t, server)
}

func TestNewPublicHTTPServer_TLSMaterialGenFailure(t *testing.T) {
	orig := generateTLSMaterialFn
	generateTLSMaterialFn = func(_ *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings) (*cryptoutilAppsTemplateServiceConfig.TLSMaterial, error) {
		return nil, errForcedTLSMaterialFailure
	}

	defer func() { generateTLSMaterialFn = orig }()

	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	tlsCfg := &cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings{}

	server, err := NewPublicHTTPServer(context.Background(), settings, tlsCfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate TLS material")
	assert.Nil(t, server)
}

// === Start with Invalid Bind Address (admin.go:193, public.go:168) ===

func TestAdminServer_Start_InvalidBindAddress(t *testing.T) {
	t.Parallel()

	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	settings.BindPrivateAddress = testInvalidBindAddress

	tlsCfg := validAutoTLSSettings(t)

	server, err := NewAdminHTTPServer(context.Background(), settings, tlsCfg)
	require.NoError(t, err)

	err = server.Start(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create admin listener")
}

func TestPublicHTTPServer_Start_InvalidBindAddress(t *testing.T) {
	t.Parallel()

	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	settings.BindPublicAddress = testInvalidBindAddress

	tlsCfg := validAutoTLSSettings(t)

	server, err := NewPublicHTTPServer(context.Background(), settings, tlsCfg)
	require.NoError(t, err)

	err = server.Start(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create listener")
}

// === Context Cancellation (admin.go ctx.Done branch, public.go ctx.Done branch) ===

func TestAdminServer_Start_ContextCancellation(t *testing.T) {
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	tlsCfg := validAutoTLSSettings(t)

	server, err := NewAdminHTTPServer(context.Background(), settings, tlsCfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	wg.Add(1)

	startErr := make(chan error, 1)

	go func() {
		defer wg.Done()

		startErr <- server.Start(ctx)
	}()

	for i := 0; i < 20; i++ {
		time.Sleep(50 * time.Millisecond)

		if server.ActualPort() > 0 {
			break
		}
	}

	require.Greater(t, int(server.ActualPort()), 0)

	cancel()
	wg.Wait()

	err = <-startErr
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin server stopped")
}

func TestPublicHTTPServer_Start_ContextCancellation(t *testing.T) {
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	tlsCfg := validAutoTLSSettings(t)

	server, err := NewPublicHTTPServer(context.Background(), settings, tlsCfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	wg.Add(1)

	startErr := make(chan error, 1)

	go func() {
		defer wg.Done()

		startErr <- server.Start(ctx)
	}()

	for i := 0; i < 20; i++ {
		time.Sleep(50 * time.Millisecond)

		if server.ActualPort() > 0 {
			break
		}
	}

	require.Greater(t, server.ActualPort(), 0)

	cancel()
	wg.Wait()

	err = <-startErr
	require.Error(t, err)
	assert.Contains(t, err.Error(), "public server stopped")
}

// === NewHTTPServers Server Creation Failures (servers.go:45, servers.go:50) ===

func TestNewHTTPServers_PublicServerCreateFailure(t *testing.T) {
	orig := generateTLSMaterialFn
	generateTLSMaterialFn = func(_ *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings) (*cryptoutilAppsTemplateServiceConfig.TLSMaterial, error) {
		return nil, errForcedTLSMaterialFailure
	}

	defer func() { generateTLSMaterialFn = orig }()

	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	h, err := NewHTTPServers(context.Background(), settings)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create public server")
	assert.Nil(t, h)
}

func TestNewHTTPServers_AdminServerCreateFailure(t *testing.T) {
	callCount := 0
	orig := generateTLSMaterialFn
	generateTLSMaterialFn = func(cfg *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings) (*cryptoutilAppsTemplateServiceConfig.TLSMaterial, error) {
		callCount++
		if callCount >= 2 {
			return nil, errForcedTLSMaterialFailure
		}

		return orig(cfg)
	}

	defer func() { generateTLSMaterialFn = orig }()

	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	h, err := NewHTTPServers(context.Background(), settings)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create admin server")
	assert.Nil(t, h)
}

// === Auto-Mode TLS Generation Errors (servers.go:83, servers.go:113) ===

func TestNewHTTPServers_AutoMode_InvalidPublicIPAutoGeneration(t *testing.T) {
	t.Parallel()

	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	settings.TLSPublicMode = cryptoutilAppsTemplateServiceConfig.TLSModeAuto
	settings.TLSPublicIPAddresses = []string{"not-a-valid-ip"}

	h, err := NewHTTPServers(context.Background(), settings)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auto-generate public server certs")
	assert.Nil(t, h)
}

func TestNewHTTPServers_AutoMode_InvalidPrivateIPAutoGeneration(t *testing.T) {
	t.Parallel()

	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	settings.TLSPublicMode = cryptoutilAppsTemplateServiceConfig.TLSModeAuto
	settings.TLSPublicIPAddresses = []string{cryptoutilSharedMagic.IPv4Loopback}
	settings.TLSPrivateMode = cryptoutilAppsTemplateServiceConfig.TLSModeAuto
	settings.TLSPrivateIPAddresses = []string{"not-a-valid-ip"}

	h, err := NewHTTPServers(context.Background(), settings)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auto-generate admin server certs")
	assert.Nil(t, h)
}

// === Admin Start with Non-Zero Port (admin.go else branch for actualPort assignment) ===

func TestAdminServer_Start_NonZeroPort(t *testing.T) {
	// Find an available port by briefly listening on port 0.
	var lc net.ListenConfig

	tempListener, err := lc.Listen(context.Background(), "tcp", cryptoutilSharedMagic.IPv4Loopback+":0")
	require.NoError(t, err)

	tcpAddr, ok := tempListener.Addr().(*net.TCPAddr)
	require.True(t, ok)

	port := uint16(tcpAddr.Port) //nolint:gosec // Port range validated by OS.

	require.NoError(t, tempListener.Close())

	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	settings.BindPrivatePort = port

	tlsCfg := validAutoTLSSettings(t)

	server, err := NewAdminHTTPServer(context.Background(), settings, tlsCfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		_ = server.Start(ctx)
	}()

	// Wait for server to start.
	time.Sleep(200 * time.Millisecond)
	require.Equal(t, int(port), server.ActualPort())

	cancel()
	wg.Wait()
}
