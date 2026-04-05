// Copyright (c) 2025 Justin Cranford
//

// Unit tests for pki-ca server NewFromConfig validation.
package server

import (
	"context"
	"testing"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilAppsCaServerConfig "cryptoutil/internal/apps/pki-ca/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestNewFromConfig_InvalidDatabaseURL(t *testing.T) {
	t.Parallel()

	cfg := &cryptoutilAppsCaServerConfig.CAServerSettings{}
	cfg.ServiceFrameworkServerSettings = &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{}
	cfg.DatabaseURL = "invalid://not-a-real-dsn"

	_, err := NewFromConfig(context.Background(), cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to build pki-ca service")
}

func TestStart_NilContext(t *testing.T) {
	t.Parallel()

	// Create a valid server first, then call Start with nil context.
	cfg := cryptoutilAppsCaServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	server, err := NewFromConfig(context.Background(), cfg)
	require.NoError(t, err)

	//nolint:staticcheck // SA1012: intentionally passing nil context to test error path.
	startErr := server.Start(nil)
	require.Error(t, startErr)
	require.Contains(t, startErr.Error(), "context cannot be nil")
}
