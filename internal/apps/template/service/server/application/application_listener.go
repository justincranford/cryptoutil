// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"context"
	"fmt"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
)

// Listener represents the top-level service application.
// Orchestrates Core (database + infrastructure) and HTTP servers (public + admin).
type Listener struct {
	Core         *Core
	PublicServer IPublicServer
	AdminServer  IAdminServer
	Settings     *cryptoutilConfig.ServiceTemplateServerSettings
}

// IPublicServer interface defines the contract for public HTTPS servers.
type IPublicServer interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
	ActualPort() int
	PublicBaseURL() string
}

// IAdminServer interface defines the contract for admin HTTPS servers.
type IAdminServer interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
	ActualPort() int
	SetReady(ready bool)
	AdminBaseURL() string
}

// ListenerConfig holds configuration for creating an Listener.
type ListenerConfig struct {
	Settings     *cryptoutilConfig.ServiceTemplateServerSettings
	PublicServer IPublicServer
	AdminServer  IAdminServer
}

// StartListener creates and initializes the top-level service application.
// This function:
// 1. Starts Core (telemetry, JWK gen, unseal, database)
// 2. Accepts pre-configured public and admin servers
// 3. Returns Listener ready to start servers.
//
// Caller is responsible for:
// - Creating public server with business logic handlers
// - Creating admin server with health check handlers
// - Calling Start() to begin serving requests
// - Calling Shutdown() to gracefully stop.
func StartListener(ctx context.Context, config *ListenerConfig) (*Listener, error) {
	if ctx == nil {
		return nil, fmt.Errorf("ctx cannot be nil")
	} else if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	} else if config.Settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	} else if config.PublicServer == nil {
		return nil, fmt.Errorf("publicServer cannot be nil")
	} else if config.AdminServer == nil {
		return nil, fmt.Errorf("adminServer cannot be nil")
	}

	// Start core infrastructure (telemetry, JWK gen, unseal, database).
	core, err := StartCore(ctx, config.Settings)
	if err != nil {
		return nil, fmt.Errorf("failed to start application core: %w", err)
	}

	app := &Listener{
		Core:         core,
		PublicServer: config.PublicServer,
		AdminServer:  config.AdminServer,
		Settings:     config.Settings,
	}

	return app, nil
}

// Start starts both public and admin servers concurrently.
// Blocks until one server fails or context is cancelled.
func (a *Listener) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	// Mark admin server as ready for health checks.
	a.AdminServer.SetReady(true)

	// Start servers in background.
	errChan := make(chan error, 2)

	go func() {
		if err := a.PublicServer.Start(ctx); err != nil {
			errChan <- fmt.Errorf("public server failed: %w", err)
		}
	}()

	go func() {
		if err := a.AdminServer.Start(ctx); err != nil {
			errChan <- fmt.Errorf("admin server failed: %w", err)
		}
	}()

	// Wait for startup errors or context cancellation.
	select {
	case err := <-errChan:
		// One server failed, shutdown the other.
		_ = a.Shutdown(context.Background())

		return err
	case <-ctx.Done():
		// Context cancelled, shutdown gracefully.
		_ = a.Shutdown(context.Background())

		return fmt.Errorf("application startup cancelled: %w", ctx.Err())
	}
}

// Shutdown gracefully shuts down all application components (LIFO order).
func (a *Listener) Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	var shutdownErr error

	// Shutdown admin server.
	if a.AdminServer != nil {
		a.AdminServer.SetReady(false)

		if err := a.AdminServer.Shutdown(ctx); err != nil {
			shutdownErr = fmt.Errorf("failed to shutdown admin server: %w", err)
		}
	}

	// Shutdown public server.
	if a.PublicServer != nil {
		if err := a.PublicServer.Shutdown(ctx); err != nil {
			if shutdownErr != nil {
				return fmt.Errorf("multiple shutdown errors: admin=%w, public=%w", shutdownErr, err)
			}

			shutdownErr = fmt.Errorf("failed to shutdown public server: %w", err)
		}
	}

	// Shutdown core infrastructure (database, telemetry, etc.).
	if a.Core != nil {
		a.Core.Shutdown()
	}

	return shutdownErr
}

// PublicPort returns the actual port the public server is listening on.
func (a *Listener) PublicPort() int {
	if a.PublicServer == nil {
		return 0
	}

	return a.PublicServer.ActualPort()
}

// AdminPort returns the actual port the admin server is listening on.
func (a *Listener) AdminPort() int {
	if a.AdminServer == nil {
		return 0
	}

	return a.AdminServer.ActualPort()
}
