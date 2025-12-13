// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"context"
	"testing"
	"time"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"

	"github.com/stretchr/testify/require"
)

func TestNewApplication_NilContext(t *testing.T) {
	t.Parallel()

	config := &cryptoutilIdentityConfig.Config{}

	app, err := NewApplication(nil, config) //nolint:staticcheck // Testing nil context validation requires passing nil.
	require.Error(t, err)
	require.Nil(t, app)
	require.Contains(t, err.Error(), "context cannot be nil")
}

func TestNewApplication_NilConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	app, err := NewApplication(ctx, nil)
	require.Error(t, err)
	require.Nil(t, app)
	require.Contains(t, err.Error(), "config cannot be nil")
}

func TestNewApplication_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := &cryptoutilIdentityConfig.Config{
		AuthZ: &cryptoutilIdentityConfig.ServerConfig{
			AdminPort: 0, // Dynamic port
		},
	}

	app, err := NewApplication(ctx, config)
	require.NoError(t, err)
	require.NotNil(t, app)
	require.NotNil(t, app.adminServer)
	require.Equal(t, config, app.config)
	require.False(t, app.shutdown)
}

func TestApplication_Start_NilContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := &cryptoutilIdentityConfig.Config{
		AuthZ: &cryptoutilIdentityConfig.ServerConfig{
			AdminPort: 0,
		},
	}

	app, err := NewApplication(ctx, config)
	require.NoError(t, err)

	err = app.Start(nil) //nolint:staticcheck // Testing nil context validation requires passing nil.
	require.Error(t, err)
	require.Contains(t, err.Error(), "context cannot be nil")
}

func TestApplication_Start_ContextCancelled(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := &cryptoutilIdentityConfig.Config{
		AuthZ: &cryptoutilIdentityConfig.ServerConfig{
			AdminPort: 0,
		},
	}

	app, err := NewApplication(ctx, config)
	require.NoError(t, err)

	// Create cancelled context.
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()

	err = app.Start(cancelledCtx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "application startup cancelled")
}

func TestApplication_Shutdown_NilContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := &cryptoutilIdentityConfig.Config{
		AuthZ: &cryptoutilIdentityConfig.ServerConfig{
			AdminPort: 0,
		},
	}

	app, err := NewApplication(ctx, config)
	require.NoError(t, err)

	// Shutdown with nil context should use Background().
	err = app.Shutdown(nil) //nolint:staticcheck // Testing nil fallback to Background() requires passing nil.
	require.NoError(t, err)
	require.True(t, app.shutdown)
}

func TestApplication_Shutdown_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := &cryptoutilIdentityConfig.Config{
		AuthZ: &cryptoutilIdentityConfig.ServerConfig{
			AdminPort: 0,
		},
	}

	app, err := NewApplication(ctx, config)
	require.NoError(t, err)

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = app.Shutdown(shutdownCtx)
	require.NoError(t, err)
	require.True(t, app.shutdown)
}

func TestApplication_AdminPort(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := &cryptoutilIdentityConfig.Config{
		AuthZ: &cryptoutilIdentityConfig.ServerConfig{
			AdminPort: 0,
		},
	}

	app, err := NewApplication(ctx, config)
	require.NoError(t, err)

	// Start server so listener is initialized.
	startCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		_ = app.Start(startCtx)
	}()

	// Wait for server to start.
	time.Sleep(100 * time.Millisecond)

	// AdminPort should delegate to adminServer.
	port, err := app.AdminPort()
	require.NoError(t, err)
	require.NotZero(t, port) // Dynamic port should be assigned

	// Cleanup.
	cancel()

	_ = app.Shutdown(ctx)
}

func TestApplication_AdminPort_NilServer(t *testing.T) {
	t.Parallel()

	// Create application with nil admin server.
	app := &Application{}

	port, err := app.AdminPort()
	require.Error(t, err)
	require.Zero(t, port)
	require.Contains(t, err.Error(), "admin server not initialized")
}
