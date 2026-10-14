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
//  1. In TestMain (with explicit cleanup before os.Exit):
//     func TestMain(m *testing.M) {
//     ctx := context.Background()
//     db, dbCleanup, err := test_help_db.NewInMemorySQLiteDBForTestMain()
//     if err != nil { log.Fatal(err) }
//     srv, err := NewMyServer(ctx, config)
//     if err != nil {
//     if cleanupErr := dbCleanup(); cleanupErr != nil { log.Printf("DB cleanup failed: %v", cleanupErr) }
//     log.Fatal(err)
//     }
//     is, err := test_orch_integration.StartIntegrationServerForTestMain(ctx, srv, db)
//     if err != nil { dbCleanup(); log.Fatal(err) }
//     exitCode := m.Run()
//     if err := is.Shutdown(ctx); err != nil { log.Printf("shutdown failed: %v", err); exitCode = 1 }
//     if cleanupErr := dbCleanup(); cleanupErr != nil { log.Printf("DB cleanup failed: %v", cleanupErr); exitCode = 1 }
//     os.Exit(exitCode)
//     }
//
//  2. In individual test functions (cleanup is registered automatically):
//     func TestSomething(t *testing.T) {
//     ctx := context.Background()
//     srv, _ := NewMyServer(ctx, config)
//     db := test_help_db.NewInMemorySQLiteDB(t)
//     is, err := test_orch_integration.StartIntegrationServer(ctx, t, srv, db)
//     if err != nil { t.Fatal(err) }
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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	http "net/http"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps-framework/service/server"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilPoll "cryptoutil/internal/shared/util/poll"
)

const (
	defaultStartupTimeout  = 30 * time.Second
	defaultStartupInterval = 100 * time.Millisecond
	defaultShutdownTimeout = 5 * time.Second
)

const (
	integrationContextCannotBeNil   = "context cannot be nil"
	integrationServerStartFailure   = "server failed to start"
	integrationServerCannotBeNil    = "server cannot be nil"
	integrationReadinessCheckFailed = "readiness check failed"
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

// readinessCheckFn validates that both HTTPS listeners can serve health requests.
type readinessCheckFn func(context.Context, cryptoutilAppsFrameworkServiceServer.ServiceServer) error

type readinessProbeFn func(context.Context, *http.Client, string) (int, error)

// StartIntegrationServer starts a server and waits for both HTTPS health endpoints.
// Intended for use in individual test functions where testing.TB.Cleanup() is available.
// For TestMain usage, use StartIntegrationServerForTestMain instead.
func StartIntegrationServer(ctx context.Context, tb testing.TB, srv cryptoutilAppsFrameworkServiceServer.ServiceServer, db *gorm.DB) (*IntegrationServer, error) {
	return startIntegrationServer(ctx, tb, srv, db, waitForServerReadiness)
}

func startIntegrationServer(ctx context.Context, tb testing.TB, srv cryptoutilAppsFrameworkServiceServer.ServiceServer, db *gorm.DB, readinessCheck readinessCheckFn) (*IntegrationServer, error) {
	switch {
	case ctx == nil:
		return nil, errors.New(integrationContextCannotBeNil)
	case srv == nil:
		return nil, errors.New(integrationServerCannotBeNil)
	case readinessCheck == nil:
		return nil, errors.New("readiness check cannot be nil")
	}

	is := &IntegrationServer{
		tb:  tb,
		srv: srv,
		db:  db,
	}

	errChan := make(chan error, 1)

	go func() {
		if startErr := srv.Start(ctx); startErr != nil {
			errChan <- startErr
		}
	}()

	pollErr := cryptoutilSharedUtilPoll.Until(ctx, defaultStartupTimeout, defaultStartupInterval, func(_ context.Context) (bool, error) {
		select {
		case startErr := <-errChan:
			return false, fmt.Errorf("%s: %w", integrationServerStartFailure, startErr)
		default:
		}

		return srv.PublicPort() > 0 && srv.AdminPort() > 0, nil
	})
	if pollErr != nil {
		return nil, fmt.Errorf("integration server: timed out waiting for server ports: %w", pollErr)
	}

	srv.SetReady(true)

	if err := readinessCheck(ctx, srv); err != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()

		if shutdownErr := srv.Shutdown(shutdownCtx); shutdownErr != nil {
			return nil, fmt.Errorf("integration server: %s: %w; cleanup failed: %w", integrationReadinessCheckFailed, err, shutdownErr)
		}

		return nil, fmt.Errorf("integration server: %s: %w", integrationReadinessCheckFailed, err)
	}

	is.cleanupFn = nil

	return is, nil
}

// StartIntegrationServerForTestMain starts a server for TestMain without testing.TB cleanup.
// Callers must manually call is.Shutdown() before os.Exit().
func StartIntegrationServerForTestMain(ctx context.Context, srv cryptoutilAppsFrameworkServiceServer.ServiceServer, db *gorm.DB) (*IntegrationServer, error) {
	return startIntegrationServer(ctx, nil, srv, db, waitForServerReadiness)
}

func waitForServerReadiness(ctx context.Context, srv cryptoutilAppsFrameworkServiceServer.ServiceServer) error {
	return waitForServerReadinessWithProbe(ctx, srv, requestHealthEndpoint)
}

func waitForServerReadinessWithProbe(ctx context.Context, srv cryptoutilAppsFrameworkServiceServer.ServiceServer, probe readinessProbeFn) error {
	switch {
	case ctx == nil:
		return errors.New(integrationContextCannotBeNil)
	case srv == nil:
		return errors.New(integrationServerCannotBeNil)
	case probe == nil:
		return errors.New("readiness probe cannot be nil")
	}

	publicRoots := srv.TLSRootCAPool()
	if publicRoots == nil {
		return errors.New("public TLS root CA pool is nil")
	}

	adminRoots := srv.AdminTLSRootCAPool()
	if adminRoots == nil {
		return errors.New("admin TLS root CA pool is nil")
	}

	publicBaseURL := strings.TrimRight(srv.PublicBaseURL(), "/")
	if publicBaseURL == "" {
		return errors.New("public base URL is empty")
	}

	adminBaseURL := strings.TrimRight(srv.AdminBaseURL(), "/")
	if adminBaseURL == "" {
		return errors.New("admin base URL is empty")
	}

	publicClient := newReadinessHTTPClient(publicRoots)
	defer publicClient.CloseIdleConnections()

	adminClient := newReadinessHTTPClient(adminRoots)
	defer adminClient.CloseIdleConnections()

	endpoints := []struct {
		name   string
		url    string
		client *http.Client
	}{
		{name: cryptoutilSharedMagic.PublicLabel, url: publicBaseURL + cryptoutilSharedMagic.DefaultServiceHealthEndpoint, client: publicClient},
		{name: "admin", url: adminBaseURL + cryptoutilSharedMagic.DefaultAdminLivezEndpoint, client: adminClient},
	}

	readinessCtx, cancel := context.WithTimeout(ctx, defaultStartupTimeout)
	defer cancel()

	var lastProbeErr error

	pollErr := cryptoutilSharedUtilPoll.Until(readinessCtx, defaultStartupTimeout, defaultStartupInterval, func(probeCtx context.Context) (bool, error) {
		for _, endpoint := range endpoints {
			statusCode, err := probe(probeCtx, endpoint.client, endpoint.url)
			if err != nil {
				lastProbeErr = fmt.Errorf("%s health endpoint request failed: %w", endpoint.name, err)

				return false, nil
			}

			if statusCode != http.StatusOK {
				lastProbeErr = fmt.Errorf("%s health endpoint returned HTTP %d", endpoint.name, statusCode)

				return false, nil
			}
		}

		lastProbeErr = nil

		return true, nil
	})
	if pollErr != nil {
		if lastProbeErr != nil {
			return fmt.Errorf("integration server: dual HTTPS readiness failed: %w; last probe: %w", pollErr, lastProbeErr)
		}

		return fmt.Errorf("integration server: dual HTTPS readiness failed: %w", pollErr)
	}

	return nil
}

func newReadinessHTTPClient(rootCAs *x509.CertPool) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				RootCAs:    rootCAs,
			},
			DisableKeepAlives: true,
		},
		Timeout: defaultStartupTimeout,
	}
}

func requestHealthEndpoint(ctx context.Context, client *http.Client, endpointURL string) (int, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpointURL, nil)
	if err != nil {
		return 0, fmt.Errorf("create readiness request: %w", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return 0, fmt.Errorf("send readiness request: %w", err)
	}

	if err := response.Body.Close(); err != nil {
		return 0, fmt.Errorf("close readiness response: %w", err)
	}

	return response.StatusCode, nil
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
