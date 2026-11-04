// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package server provides reusable server infrastructure for cryptoutil services.
//
// This package extracts the dual HTTPS server pattern from the JOSE and Identity services,
// providing a clean separation between public (business) and admin (health check) servers.
//
// Key Features:
// - Dual HTTPS servers (public + admin) with independent lifecycles
// - Dynamic port allocation for testing (port 0)
// - Configured ports for production deployments
// - Health check endpoints (/admin/api/v1/livez, /admin/api/v1/readyz)
// - Graceful shutdown with context-based timeout
// - Self-signed TLS certificate generation
// - Mutex-protected state management
package server

import (
	"context"
	"crypto/x509"
	"fmt"
	"sync"

	cryptoutilAppsFrameworkServiceServerReadiness "cryptoutil/internal/apps-framework/service/server/readiness"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Application represents a unified service application managing both public and admin servers.
//
// The Application follows the dual HTTPS server pattern where:
// - Public server handles business logic (APIs, UI, external clients)
// - Admin server handles health checks and graceful shutdown (Kubernetes probes, monitoring)
//
// Both servers run concurrently and have independent lifecycles managed by the Application.
type Application struct {
	publicServer     IPublicServer
	adminServer      IAdminServer
	readinessMonitor *cryptoutilAppsFrameworkServiceServerReadiness.Monitor
	mu               sync.RWMutex
	shutdown         bool
}

// IPublicServer interface defines the contract for public HTTPS servers.
// Implementations must provide:
// - Start: Begin listening for HTTPS requests (blocks until shutdown or error)
// - Shutdown: Gracefully shutdown the server with context timeout
// - ActualPort: Return the actual port after dynamic allocation.
// - PublicBaseURL: Return the base URL for public API access.
type IPublicServer interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
	ActualPort() int
	PublicBaseURL() string
}

// IAdminServer interface defines the contract for admin HTTPS servers.
// Implementations must provide:
// - Start: Begin listening on 127.0.0.1:9090 for admin API requests (blocks until shutdown or error)
// - Shutdown: Gracefully shutdown the admin server with context timeout
// - ActualPort: Return the actual port (should always be 9090)
// - SetReady: Mark server as ready to handle readyz health checks (thread-safe).
// - AdminBaseURL: Return the base URL for admin API access.
// - AdminTLSRootCAPool: Return the root CA certificate pool for the admin server's TLS chain.
type IAdminServer interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
	ActualPort() int
	SetReady(ready bool)
	AdminBaseURL() string
	AdminTLSRootCAPool() *x509.CertPool
}

// NewApplication creates a new service application with public and admin servers.
//
// Parameters:
// - ctx: Context for initialization (must not be nil)
// - publicServer: Public server instance implementing IPublicServer interface
// - adminServer: Admin server instance implementing IAdminServer interface
//
// Returns:
// - *Application: Configured application ready to start
// - error: Non-nil if parameters invalid or initialization fails.
func NewApplication(
	ctx context.Context,
	publicServer IPublicServer,
	adminServer IAdminServer,
) (*Application, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if publicServer == nil {
		return nil, fmt.Errorf("publicServer cannot be nil")
	} else if adminServer == nil {
		return nil, fmt.Errorf("adminServer cannot be nil")
	}

	return &Application{
		publicServer: publicServer,
		adminServer:  adminServer,
		shutdown:     false,
	}, nil
}

// Start starts both public and admin servers concurrently.
//
// Servers are started in separate goroutines to allow parallel initialization.
// This method blocks until:
// - One server fails to start (returns error, shuts down other server)
// - Context is cancelled (gracefully shuts down both servers)
//
// Error Handling:
// - If public server fails: Admin server is shutdown, error returned
// - If admin server fails: Public server is shutdown, error returned
// - If context cancelled: Both servers shutdown gracefully
//
// Parameters:
// - ctx: Context for controlling server lifecycle (cancellation triggers shutdown)
//
// Returns:
// - error: Non-nil if either server fails to start or context cancelled.
func (a *Application) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	errChan := make(chan error, 2)
	startedChan := make(chan struct{}, 2)

	startServer := func(name string, server interface{ Start(context.Context) error }) {
		go func() {
			if startNotifier, ok := server.(interface{ Started() <-chan struct{} }); ok {
				started := startNotifier.Started()
				if started != nil {
					go func() {
						select {
						case <-started:
							startedChan <- struct{}{}
						case <-ctx.Done():
						}
					}()
				} else {
					startedChan <- struct{}{}
				}
			} else {
				startedChan <- struct{}{}
			}

			if err := server.Start(ctx); err != nil {
				errChan <- fmt.Errorf("%s server failed: %w", name, err)
			}
		}()
	}

	startServer(cryptoutilSharedMagic.PublicLabel, a.publicServer)
	startServer("admin", a.adminServer)

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

	if startupAdmin, ok := a.adminServer.(interface{ MarkStartupComplete() }); ok {
		if err := a.startReadinessMonitor(ctx); err != nil {
			_ = a.Shutdown(context.Background())

			return err
		}

		startupAdmin.MarkStartupComplete()
	} else if err := a.startReadinessMonitor(ctx); err != nil {
		_ = a.Shutdown(context.Background())

		return err
	}

	a.SetReady(true)

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

// Shutdown gracefully shuts down both public and admin servers.
//
// Shutdown respects the provided context timeout and attempts to:
// 1. Mark application as shutdown (prevents new requests)
// 2. Shutdown public server (drain existing connections)
// 3. Shutdown admin server (stop health checks)
//
// If both servers fail to shutdown, both errors are returned (public error takes precedence).
//
// Parameters:
// - ctx: Context with timeout for graceful shutdown (if nil, uses context.Background())
//
// Returns:
// - error: Non-nil if either server fails to shutdown cleanly.
func (a *Application) Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	a.mu.Lock()
	a.shutdown = true
	readinessMonitor := a.readinessMonitor
	a.mu.Unlock()

	var shutdownErr error

	if drainingAdmin, ok := a.adminServer.(interface{ BeginDrain() }); ok {
		drainingAdmin.BeginDrain()
	} else if a.adminServer != nil {
		a.adminServer.SetReady(false)
	}

	// Shutdown public server.
	if a.publicServer != nil {
		if err := a.publicServer.Shutdown(ctx); err != nil {
			shutdownErr = fmt.Errorf("failed to shutdown public server: %w", err)
		}
	}

	if readinessMonitor != nil {
		if err := readinessMonitor.Stop(ctx); err != nil {
			if shutdownErr != nil {
				shutdownErr = fmt.Errorf("multiple shutdown errors: public=%w, monitor=%w", shutdownErr, err)
			} else {
				shutdownErr = fmt.Errorf("failed to stop readiness monitor: %w", err)
			}
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

// SetReadinessMonitor installs dependency observations before Start. The monitor is
// also exposed to the admin server so readiness reads only its in-memory snapshot.
func (a *Application) SetReadinessMonitor(monitor *cryptoutilAppsFrameworkServiceServerReadiness.Monitor) {
	a.mu.Lock()
	a.readinessMonitor = monitor
	a.mu.Unlock()

	if adminMonitorSetter, ok := a.adminServer.(interface {
		SetReadinessMonitor(*cryptoutilAppsFrameworkServiceServerReadiness.Monitor)
	}); ok {
		adminMonitorSetter.SetReadinessMonitor(monitor)
	}
}

func (a *Application) startReadinessMonitor(ctx context.Context) error {
	a.mu.RLock()
	monitor := a.readinessMonitor
	a.mu.RUnlock()

	if monitor == nil {
		return nil
	}

	if err := monitor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start readiness monitor: %w", err)
	}

	startupCtx, cancel := context.WithTimeout(ctx, cryptoutilSharedMagic.ReadinessStartupTimeout)
	defer cancel()

	if err := monitor.WaitReady(startupCtx); err != nil {
		return fmt.Errorf("required dependency observations failed during startup: %w", err)
	}

	return nil
}

// PublicPort returns the actual port the public server is listening on.
//
// Useful for tests using dynamic port allocation (port 0) where the OS assigns
// an available port. This method returns the actual assigned port after server starts.
//
// Returns:
// - int: Actual port number, or 0 if public server not initialized.
func (a *Application) PublicPort() int {
	if a.publicServer == nil {
		return 0
	}

	return a.publicServer.ActualPort()
}

// AdminPort returns the actual port the admin server is listening on.
//
// The admin server always binds to 127.0.0.1:9090 per security requirements,
// so this method should always return 9090 (or 0 if not initialized).
//
// Returns:
// - int: Actual port number (should be 9090, or 0 if not initialized).
func (a *Application) AdminPort() int {
	if a.adminServer == nil {
		return 0
	}

	return a.adminServer.ActualPort()
}

// PublicBaseURL returns the base URL for the public server.
//
// Returns the complete base URL (protocol + address + port) for making requests
// to the public API endpoints. Useful for constructing full URLs in tests and clients.
//
// Returns:
// - string: Base URL (e.g., "https://127.0.0.1:8080"), or empty string if not initialized.
func (a *Application) PublicBaseURL() string {
	if a.publicServer == nil {
		return ""
	}

	return a.publicServer.PublicBaseURL()
}

// AdminBaseURL returns the base URL for the admin server.
//
// Returns the complete base URL (protocol + address + port) for making requests
// to the admin API endpoints (health checks, shutdown, etc.). Useful for tests and monitoring.
//
// Returns:
// - string: Base URL (e.g., "https://127.0.0.1:9090"), or empty string if not initialized.
func (a *Application) AdminBaseURL() string {
	if a.adminServer == nil {
		return ""
	}

	return a.adminServer.AdminBaseURL()
}

// IsShutdown returns whether the application is shutting down or shutdown.
//
// Thread-safe check for shutdown state, useful for health checks and request handling.
//
// Returns:
// - bool: true if shutdown initiated, false otherwise.
func (a *Application) IsShutdown() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.shutdown
}

// SetReady marks the admin server as ready to accept traffic.
//
// Applications should call SetReady(true) after initializing all dependencies
// (database connections, caches, external services, etc.) but before the server
// starts accepting requests. This enables the /admin/api/v1/readyz endpoint to return
// HTTP 200 OK instead of 503 Service Unavailable.
//
// Parameters:
// - ready: true to mark ready, false to mark not ready.
func (a *Application) SetReady(ready bool) {
	if a.adminServer != nil {
		a.adminServer.SetReady(ready)
	}
}

// PublicServerBase returns the underlying PublicServerBase if the public server is of that type.
// This is used for testing to access the base infrastructure.
// Returns nil if the public server is not a *PublicServerBase (e.g., a mock).
func (a *Application) PublicServerBase() *PublicServerBase {
	if base, ok := a.publicServer.(*PublicServerBase); ok {
		// Type assertion succeeded - return the PublicServerBase.
		return base
	}

	// Type assertion failed - public server is not a PublicServerBase (e.g., a mock).
	return nil
}

// TLSRootCAPool returns the root CA certificate pool from the public server's TLS chain.
// Used by test infrastructure to configure secure HTTP clients without InsecureSkipVerify.
// Returns nil if the public server is not a PublicServerBase or has no TLS material.
func (a *Application) TLSRootCAPool() *x509.CertPool {
	base := a.PublicServerBase()
	if base == nil {
		return nil
	}

	return base.TLSRootCAPool()
}

// AdminTLSRootCAPool returns the root CA certificate pool from the admin server's TLS chain.
// Used by test infrastructure to configure secure HTTP clients for the admin endpoint without InsecureSkipVerify.
// Returns nil if the admin server is nil.
func (a *Application) AdminTLSRootCAPool() *x509.CertPool {
	if a.adminServer == nil {
		return nil
	}

	return a.adminServer.AdminTLSRootCAPool()
}
