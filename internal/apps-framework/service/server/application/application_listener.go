// Copyright (c) 2025-2026 Justin Cranford.
//
//

package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilSharedContainer "cryptoutil/internal/shared/container"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// Listener represents the top-level service application.
// Orchestrates Core (database + infrastructure) and HTTP servers (public + admin).
type Listener struct {
	Core         *Core
	PublicServer IPublicServer
	AdminServer  IAdminServer
	Settings     *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings
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
	Settings     *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings
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
	return startListenerInternal(ctx, config,
		cryptoutilSharedCryptoJose.NewJWKGenService,
		cryptoutilSharedContainer.StartPostgres,
		sql.Open,
		func(dialector gorm.Dialector, cfg *gorm.Config) (*gorm.DB, error) {
			return gorm.Open(dialector, cfg)
		},
	)
}

func startListenerInternal(
	ctx context.Context,
	config *ListenerConfig,
	newJWKGenServiceFn func(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, devMode bool) (*cryptoutilSharedCryptoJose.JWKGenService, error),
	startPostgresFn func(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, dbName, username, password string) (string, func(), error),
	sqlOpenFn func(driverName, dataSourceName string) (*sql.DB, error),
	gormOpenSQLiteFn func(dialector gorm.Dialector, config *gorm.Config) (*gorm.DB, error),
) (*Listener, error) {
	if ctx == nil {
		return nil, errors.New(valueCtxCannotBeNil)
	} else if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	} else if config.Settings == nil {
		return nil, errors.New(valueSettingsCannotBeNil)
	} else if config.PublicServer == nil {
		return nil, fmt.Errorf("publicServer cannot be nil")
	} else if config.AdminServer == nil {
		return nil, fmt.Errorf("adminServer cannot be nil")
	}

	// Start core infrastructure (telemetry, JWK gen, unseal, database).
	core, err := startCoreInternal(ctx, config.Settings, newJWKGenServiceFn, startPostgresFn, sqlOpenFn, gormOpenSQLiteFn)
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

	errChan := make(chan error, 2)
	startedChan := make(chan struct{}, 2)

	startServer := func(name string, server interface{ Start(context.Context) error }) {
		go func() {
			if startNotifier, ok := server.(interface{ Started() <-chan struct{} }); ok {
				started := startNotifier.Started()
				if started == nil {
					startedChan <- struct{}{}
				} else {
					go func() {
						select {
						case <-started:
							startedChan <- struct{}{}
						case <-ctx.Done():
						}
					}()
				}
			} else {
				startedChan <- struct{}{}
			}

			if err := server.Start(ctx); err != nil {
				errChan <- fmt.Errorf("%s server failed: %w", name, err)
			}
		}()
	}

	startServer(cryptoutilSharedMagic.PublicLabel, a.PublicServer)
	startServer("admin", a.AdminServer)

	startedServers := 0
	for startedServers < 2 {
		select {
		case <-startedChan:
			startedServers++
		case err := <-errChan:
			_ = a.Shutdown(context.Background())

			return err
		case <-ctx.Done():
			_ = a.Shutdown(context.Background())

			return fmt.Errorf("application startup cancelled: %w", ctx.Err())
		}
	}

	if startupAdmin, ok := a.AdminServer.(interface{ MarkStartupComplete() }); ok {
		startupAdmin.MarkStartupComplete()
	}

	a.AdminServer.SetReady(true)

	select {
	case err := <-errChan:
		_ = a.Shutdown(context.Background())

		return err
	case <-ctx.Done():
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

	// Mark the process unready before draining public requests.
	if a.AdminServer != nil {
		if drainingAdmin, ok := a.AdminServer.(interface{ BeginDrain() }); ok {
			drainingAdmin.BeginDrain()
		} else {
			a.AdminServer.SetReady(false)
		}
	}

	// Shutdown public server first so admin probes remain available during request drain.
	if a.PublicServer != nil {
		if err := a.PublicServer.Shutdown(ctx); err != nil {
			shutdownErr = fmt.Errorf("failed to shutdown public server: %w", err)
		}
	}

	// Shutdown admin server after public requests have drained.
	if a.AdminServer != nil {
		if err := a.AdminServer.Shutdown(ctx); err != nil {
			if shutdownErr != nil {
				return fmt.Errorf("multiple shutdown errors: public=%w, admin=%w", shutdownErr, err)
			}

			shutdownErr = fmt.Errorf("failed to shutdown admin server: %w", err)
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
