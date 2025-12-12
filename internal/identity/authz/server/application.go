// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"context"
	"fmt"
	"sync"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
)

// Application represents the unified AuthZ server application (public + admin).
type Application struct {
	config      *cryptoutilIdentityConfig.Config
	adminServer *AdminServer
	mu          sync.RWMutex
	shutdown    bool
}

// NewApplication creates a new AuthZ application with public and admin servers.
func NewApplication(
	ctx context.Context,
	config *cryptoutilIdentityConfig.Config,
) (*Application, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	app := &Application{
		config:   config,
		shutdown: false,
	}

	// Create admin server.
	adminServer, err := NewAdminServer(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin server: %w", err)
	}

	app.adminServer = adminServer

	return app, nil
}

// Start starts both public and admin servers concurrently.
func (a *Application) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	// Start admin server in background.
	errChan := make(chan error, 1)

	go func() {
		if err := a.adminServer.Start(ctx); err != nil {
			errChan <- fmt.Errorf("admin server failed: %w", err)
		}
	}()

	// Wait for startup errors (admin server blocks on Listen).
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("application startup cancelled: %w", ctx.Err())
	}
}

// Shutdown gracefully shuts down all servers.
func (a *Application) Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	a.mu.Lock()
	a.shutdown = true
	a.mu.Unlock()

	// Shutdown admin server.
	if a.adminServer != nil {
		if err := a.adminServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown admin server: %w", err)
		}
	}

	return nil
}

// AdminPort returns the actual port the admin server is listening on.
func (a *Application) AdminPort() (int, error) {
	if a.adminServer == nil {
		return 0, fmt.Errorf("admin server not initialized")
	}

	return a.adminServer.ActualPort()
}
