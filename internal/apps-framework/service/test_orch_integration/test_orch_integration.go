// Copyright (c) 2025-2026 Justin Cranford.

// Package test_orch_integration provides orchestration for starting and managing
// individual PS-ID servers in integration tests. It handles server startup, database setup,
// health checks, and graceful shutdown with support for both successful and error-path testing.
//
// The primary API is StartIntegrationServer() which returns an IntegrationServer handle
// with public/admin URLs, database access, and registered cleanup callbacks.
//
// Consumed by:
//   - All 28 internal/apps TestMain files (for server integration tests)
//   - Framework integration test suites
//   - Repository and API test packages
package test_orch_integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gorm.io/gorm"

	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps-framework/service/server"
	cryptoutilSharedUtilPoll "cryptoutil/internal/shared/util/poll"
)

const (
	defaultStartupTimeout  = 30 * time.Second
	defaultStartupInterval = 100 * time.Millisecond
	defaultShutdownTimeout = 5 * time.Second
)

// IntegrationServer represents a running integration test environment with a single PS-ID server,
// database, and supporting infrastructure. It wraps the ServiceServer pattern and adds
// database and fixture management.
type IntegrationServer struct {
	tb          testing.TB
	srv         cryptoutilAppsFrameworkServiceServer.ServiceServer
	db          *gorm.DB
	cleanupFn   func() error
	brokeDBErr  error // if non-nil, DB was intentionally broken for error-path testing
	brokeAPIErr error // if non-nil, API was intentionally broken for error-path testing
}

// StartIntegrationServer starts a new integration test server with the given ServiceServer.
// Registers cleanup callback via tb.Cleanup() for automatic shutdown.
// Returns an IntegrationServer handle with URLs and database access.
func StartIntegrationServer(ctx context.Context, tb testing.TB, srv cryptoutilAppsFrameworkServiceServer.ServiceServer, db *gorm.DB) (*IntegrationServer, error) {
	if srv == nil {
		return nil, fmt.Errorf("server cannot be nil")
	}

	is := &IntegrationServer{
		tb:  tb,
		srv: srv,
		db:  db,
	}

	// Start server in background
	errChan := make(chan error, 1)

	go func() {
		if startErr := srv.Start(ctx); startErr != nil {
			errChan <- startErr
		}
	}()

	// Wait for both public and admin ports to be allocated
	pollErr := cryptoutilSharedUtilPoll.Until(ctx, defaultStartupTimeout, defaultStartupInterval, func(_ context.Context) (bool, error) {
		select {
		case startErr := <-errChan:
			return false, fmt.Errorf("server failed to start: %w", startErr)
		default:
		}

		return srv.PublicPort() > 0 && srv.AdminPort() > 0, nil
	})
	if pollErr != nil {
		return nil, fmt.Errorf("integration server: timed out waiting for server ports: %w", pollErr)
	}

	// Mark server ready for health checks
	srv.SetReady(true)

	// Register cleanup callback
	tb.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()

		if shutdownErr := srv.Shutdown(shutdownCtx); shutdownErr != nil {
			tb.Logf("integration server: failed to shut down server: %v", shutdownErr)
		}

		if is.cleanupFn != nil {
			if err := is.cleanupFn(); err != nil {
				tb.Logf("integration server: cleanup error: %v", err)
			}
		}
	})

	return is, nil
}

// PublicBaseURL returns the public HTTPS URL for the running server.
func (is *IntegrationServer) PublicBaseURL() string {
	if is.srv == nil {
		return ""
	}

	return fmt.Sprintf("https://127.0.0.1:%d", is.srv.PublicPort())
}

// AdminBaseURL returns the admin HTTPS URL for the running server.
func (is *IntegrationServer) AdminBaseURL() string {
	if is.srv == nil {
		return ""
	}

	return fmt.Sprintf("https://127.0.0.1:%d", is.srv.AdminPort())
}

// Server returns the underlying ServiceServer.
func (is *IntegrationServer) Server() cryptoutilAppsFrameworkServiceServer.ServiceServer {
	return is.srv
}

// DB returns the GORM database handle for the test suite.
func (is *IntegrationServer) DB() *gorm.DB {
	return is.db
}

// BrokenDB returns true if the database was intentionally broken for error-path testing.
func (is *IntegrationServer) BrokenDB() bool {
	return is.brokeDBErr != nil
}

// BrokenDBError returns the error that broke the database, or nil if DB is not broken.
func (is *IntegrationServer) BrokenDBError() error {
	return is.brokeDBErr
}

// BrokenAPI returns true if the API was intentionally broken for error-path testing.
func (is *IntegrationServer) BrokenAPI() bool {
	return is.brokeAPIErr != nil
}

// BrokenAPIError returns the error that broke the API, or nil if API is not broken.
func (is *IntegrationServer) BrokenAPIError() error {
	return is.brokeAPIErr
}

// BuildBrokenDBFixture returns an IntegrationServer with a deliberately broken database.
// Used for error-path testing. The server is NOT started.
func BuildBrokenDBFixture(tb testing.TB, reason string, srv cryptoutilAppsFrameworkServiceServer.ServiceServer) (*IntegrationServer, error) {
	is := &IntegrationServer{
		tb:         tb,
		srv:        srv,
		brokeDBErr: fmt.Errorf("intentionally broken DB: %s", reason),
	}

	tb.Cleanup(func() {
		if is.cleanupFn != nil {
			if err := is.cleanupFn(); err != nil {
				tb.Logf("integration server: cleanup error: %v", err)
			}
		}
	})

	return is, nil
}

// BuildBrokenAPIFixture returns an IntegrationServer with a deliberately broken API.
// Used for error-path testing. The server is NOT started.
func BuildBrokenAPIFixture(tb testing.TB, reason string, srv cryptoutilAppsFrameworkServiceServer.ServiceServer, db *gorm.DB) (*IntegrationServer, error) {
	is := &IntegrationServer{
		tb:          tb,
		srv:         srv,
		db:          db,
		brokeAPIErr: fmt.Errorf("intentionally broken API: %s", reason),
	}

	tb.Cleanup(func() {
		if is.cleanupFn != nil {
			if err := is.cleanupFn(); err != nil {
				tb.Logf("integration server: cleanup error: %v", err)
			}
		}
	})

	return is, nil
}
