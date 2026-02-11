// Copyright (c) 2025 Justin Cranford
//
// Unit tests for JOSE-JA server NewFromConfig validation.
package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsJoseJaServerConfig "cryptoutil/internal/apps/jose/ja/server/config"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
)

func TestNewFromConfig_NilContext(t *testing.T) {
	t.Parallel()

	// Create a valid config.
	cfg := &cryptoutilAppsJoseJaServerConfig.JoseJAServerSettings{}

	// Call with nil context - should fail validation.
	//nolint:staticcheck // SA1012: Intentionally passing nil context to test error handling
	_, err := NewFromConfig(nil, cfg)

	require.Error(t, err)
	require.Contains(t, err.Error(), "context cannot be nil")
}

func TestNewFromConfig_NilConfig(t *testing.T) {
	t.Parallel()

	// Call with nil config - should fail validation.
	_, err := NewFromConfig(context.Background(), nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "config cannot be nil")
}

func TestNewFromConfig_InvalidDatabaseURL(t *testing.T) {
	t.Parallel()

	// Create config with invalid database URL to trigger builder failure.
	cfg := &cryptoutilAppsJoseJaServerConfig.JoseJAServerSettings{}
	cfg.ServiceTemplateServerSettings = &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{}
	cfg.DatabaseURL = "invalid://not-a-real-dsn"

	// Call with invalid config - should fail during builder.Build().
	_, err := NewFromConfig(context.Background(), cfg)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to build jose-ja service")
}
