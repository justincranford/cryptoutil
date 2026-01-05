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

// Application represents the unified RS server application (public + admin).
type Application struct {
	config       *cryptoutilIdentityConfig.Config
	publicServer *PublicServer
	adminServer  *AdminServer
	mu           sync.RWMutex
	shutdown     bool
}

// NewApplication creates a new RS application with public and admin servers.
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

	// Create public server.
	publicServer, err := NewPublicServer(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create public server: %w", err)
	}

	app.publicServer = publicServer

	// Create admin server.
	adminServer, err := NewAdminHTTPServer(ctx, config)
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

	// Start servers in background.
	errChan := make(chan error, 2)

	go func() {
		if err := a.publicServer.Start(ctx); err != nil {
			errChan <- fmt.Errorf("public server failed: %w", err)
		}
	}()

	go func() {
		if err := a.adminServer.Start(ctx); err != nil {
			errChan <- fmt.Errorf("admin server failed: %w", err)
		}
	}()

	// Wait for startup errors (both servers block on Listen).
	select {
	case err := <-errChan:
		_ = a.Shutdown(context.Background())

		return err
	case <-ctx.Done():
		_ = a.Shutdown(context.Background())

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

	var shutdownErr error

	// Shutdown public server.
	if a.publicServer != nil {
		if err := a.publicServer.Shutdown(); err != nil {
			shutdownErr = fmt.Errorf("failed to shutdown public server: %w", err)
		}
	}

	// Shutdown admin server.
	if a.adminServer != nil {
		if err := a.adminServer.Shutdown(ctx); err != nil {
			if shutdownErr != nil {
				return fmt.Errorf("multiple shutdown errors: public=%w, admin=%w", shutdownErr, err)
			}

			return fmt.Errorf("failed to shutdown admin server: %w", err)
		}
	}

	return shutdownErr
}

// PublicPort returns the actual port the public server is listening on.
func (a *Application) PublicPort() int {
	if a.publicServer == nil {
		return 0
	}

	return a.publicServer.ActualPort()
}

// AdminPort returns the actual port the admin server is listening on.
func (a *Application) AdminPort() int {
	if a.adminServer == nil {
		return 0
	}

	return a.adminServer.ActualPort()
}
