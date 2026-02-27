// Copyright (c) 2025 Justin Cranford
//

// Unit tests for skeleton-template server NewFromConfig validation.
package server

import (
"context"
"testing"

cryptoutilAppsSkeletonTemplateServerConfig "cryptoutil/internal/apps/skeleton/template/server/config"
cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

"github.com/stretchr/testify/require"
)

func TestNewFromConfig_NilContext(t *testing.T) {
t.Parallel()

cfg := &cryptoutilAppsSkeletonTemplateServerConfig.SkeletonTemplateServerSettings{}

//nolint:staticcheck // SA1012: intentionally passing nil context to test error path.
_, err := NewFromConfig(nil, cfg)
require.Error(t, err)
require.Contains(t, err.Error(), "context cannot be nil")
}

func TestNewFromConfig_NilConfig(t *testing.T) {
t.Parallel()

_, err := NewFromConfig(context.Background(), nil)
require.Error(t, err)
require.Contains(t, err.Error(), "config cannot be nil")
}

func TestNewFromConfig_InvalidDatabaseURL(t *testing.T) {
t.Parallel()

cfg := &cryptoutilAppsSkeletonTemplateServerConfig.SkeletonTemplateServerSettings{}
cfg.ServiceTemplateServerSettings = &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{}
cfg.DatabaseURL = "invalid://not-a-real-dsn"

_, err := NewFromConfig(context.Background(), cfg)
require.Error(t, err)
require.Contains(t, err.Error(), "failed to build skeleton-template service")
}

func TestStart_NilContext(t *testing.T) {
t.Parallel()

// Create a valid server first, then call Start with nil context.
cfg := cryptoutilAppsSkeletonTemplateServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

server, err := NewFromConfig(context.Background(), cfg)
require.NoError(t, err)

//nolint:staticcheck // SA1012: intentionally passing nil context to test error path.
startErr := server.Start(nil)
require.Error(t, startErr)
require.Contains(t, startErr.Error(), "failed to start application")
}
