// Copyright (c) 2025 Justin Cranford
//

// Package testserver provides shared server startup helpers for cryptoutil service tests.
// Centralizes the TestMain server startup boilerplate to eliminate duplication across services.
package testserver

import (
	"context"
	"fmt"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilSharedUtilPoll "cryptoutil/internal/shared/util/poll"
)

const (
	defaultStartupTimeout  = 5 * time.Second
	defaultStartupInterval = 100 * time.Millisecond
	defaultShutdownTimeout = 5 * time.Second
)

// StartAndWait starts a ServiceServer in the background, waits for both public and admin ports
// to be allocated, marks the server ready, and registers cleanup for graceful shutdown.
//
// Replaces the 40-line TestMain server startup boilerplate pattern.
// The returned server is the same instance passed in (for chaining).
//
// Usage in TestMain:
//
//	func TestMain(m *testing.M) {
//	    cfg := config.NewTestConfig(magic.IPv4Loopback, 0, true)
//	    srv, _ := server.NewFromConfig(ctx, cfg)
//	    testserver.StartAndWait(ctx, m, srv)
//	    os.Exit(m.Run())
//	}
func StartAndWait(ctx context.Context, t testing.TB, srv cryptoutilAppsTemplateServiceServer.ServiceServer) cryptoutilAppsTemplateServiceServer.ServiceServer {
	t.Helper()

	errChan := make(chan error, 1)

	go func() {
		if startErr := srv.Start(ctx); startErr != nil {
			errChan <- startErr
		}
	}()

	pollErr := cryptoutilSharedUtilPoll.Until(ctx, defaultStartupTimeout, defaultStartupInterval, func(_ context.Context) (bool, error) {
		select {
		case startErr := <-errChan:
			return false, fmt.Errorf("server failed to start: %w", startErr)
		default:
		}

		return srv.PublicPort() > 0 && srv.AdminPort() > 0, nil
	})
	if pollErr != nil {
		t.Fatalf("testserver: timed out waiting for server ports: %v", pollErr)
	}

	srv.SetReady(true)

	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()

		if shutdownErr := srv.Shutdown(shutdownCtx); shutdownErr != nil {
			t.Logf("testserver: failed to shut down server: %v", shutdownErr)
		}
	})

	return srv
}
