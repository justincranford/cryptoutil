// Copyright (c) 2025 Justin Cranford

package listener

import (
	"context"
	"testing"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTLSGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testUnknownTLSMode = "unknown"

func TestNewHTTPServers_AutoMode_HappyPath(t *testing.T) {
	settings := cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true)

	ctx := context.Background()
	h, err := NewHTTPServers(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, h)
	require.NotNil(t, h.PublicServer)
	require.NotNil(t, h.AdminServer)
}

func TestNewHTTPServers_NilContext(t *testing.T) {
	settings := cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true)

	h, err := NewHTTPServers(nil, settings) //nolint:staticcheck // Testing nil context handling.
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context cannot be nil")
	assert.Nil(t, h)
}

func TestNewHTTPServers_NilSettings(t *testing.T) {
	ctx := context.Background()

	h, err := NewHTTPServers(ctx, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "settings cannot be nil")
	assert.Nil(t, h)
}

func TestNewHTTPServers_StaticMode_HappyPath(t *testing.T) {
	// Generate static certs first using auto mode.
	staticTLS, err := cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{"localhost"},
		[]string{cryptoutilMagic.IPv4Loopback},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	require.NoError(t, err)

	// Create settings with static TLS mode.
	settings := cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true)
	settings.TLSPublicMode = cryptoutilConfig.TLSModeStatic
	settings.TLSPrivateMode = cryptoutilConfig.TLSModeStatic
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
	settings := cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true)
	settings.TLSPublicMode = testUnknownTLSMode

	ctx := context.Background()
	h, err := NewHTTPServers(ctx, settings)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown public TLS mode")
	assert.Nil(t, h)
}

func TestNewHTTPServers_UnknownPrivateTLSMode(t *testing.T) {
	settings := cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true)
	settings.TLSPrivateMode = testUnknownTLSMode

	ctx := context.Background()
	h, err := NewHTTPServers(ctx, settings)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown admin TLS mode")
	assert.Nil(t, h)
}

