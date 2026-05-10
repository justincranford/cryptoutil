// Copyright (c) 2025-2026 Justin Cranford.

// Package test_orch_integration provides orchestration for starting and managing
// individual PS-ID servers in integration tests. It handles server startup, database setup,
// health checks, and graceful shutdown with support for both successful and error-path testing.
//
// The primary API is StartIntegrationServer() which returns an IntegrationServer handle
// with public/admin URLs, database access, and registered cleanup callbacks.
//
// Usage patterns:
//
//  1. In TestMain (with manual cleanup via defer):
//     func TestMain(m *testing.M) {
//     ctx := context.Background()
//     srv, err := NewMyServer(ctx, config)
//     db := testdb.NewInMemorySQLiteDB(&testing.T{}) // simplified
//     is, err := test_orch_integration.StartIntegrationServer(ctx, (*testing.T)(nil), srv, db)
//     defer is.Shutdown(context.Background())
//     os.Exit(m.Run())
//     }
//
//  2. In individual test functions (with tb.Cleanup):
//     func TestSomething(t *testing.T) {
//     ctx := context.Background()
//     srv, _ := NewMyServer(ctx, config)
//     db := testdb.NewInMemorySQLiteDB(t)
//     is, _ := test_orch_integration.StartIntegrationServer(ctx, t, srv, db)
//     t.Cleanup(func() { is.Shutdown(context.Background()) })
//     // Use is.PublicBaseURL(), is.AdminBaseURL(), is.DB()
//     }
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
// Intended for use in individual test functions where testing.TB.Cleanup() is available.
// For TestMain usage, use StartIntegrationServerForTestMain instead.
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

	// Store cleanup function for manual or automatic cleanup
	// Callers can choose: defer is.Shutdown() in TestMain, or tb.Cleanup(is.Shutdown) in individual tests
	is.cleanupFn = nil

	return is, nil
}

// StartIntegrationServerForTestMain starts a new integration test server specifically for use in TestMain.
// Unlike StartIntegrationServer, this does not require testing.TB and does not use tb.Cleanup().
// Callers must manually call is.Shutdown() in a defer statement before os.Exit().
func StartIntegrationServerForTestMain(ctx context.Context, srv cryptoutilAppsFrameworkServiceServer.ServiceServer, db *gorm.DB) (*IntegrationServer, error) {
	if srv == nil {
		return nil, fmt.Errorf("server cannot be nil")
	}

	is := &IntegrationServer{
		tb:  nil, // No testing.TB for TestMain usage
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

	// Store cleanup function for manual or automatic cleanup
	is.cleanupFn = nil

	return is, nil
}

// Shutdown gracefully shuts down the integration server and cleans up resources.
// Can be called manually in TestMain via defer, or registered with tb.Cleanup() in individual tests.
func (is *IntegrationServer) Shutdown(ctx context.Context) error {
	if is.srv == nil {
		return nil
	}

	if shutdownErr := is.srv.Shutdown(ctx); shutdownErr != nil {
		if is.tb != nil {
			is.tb.Logf("integration server: failed to shut down server: %v", shutdownErr)
		}

		return fmt.Errorf("integration server: shutdown failed: %w", shutdownErr)
	}

	if is.cleanupFn != nil {
		if err := is.cleanupFn(); err != nil {
			if is.tb != nil {
				is.tb.Logf("integration server: cleanup error: %v", err)
			}

			return fmt.Errorf("integration server: cleanup failed: %w", err)
		}
	}

	return nil
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
