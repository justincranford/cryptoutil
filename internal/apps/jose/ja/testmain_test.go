// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package ja

import (
	"context"
	"fmt"
	http "net/http"
	"os"
	"testing"
	"time"

	cryptoutilAppsJoseJaServer "cryptoutil/internal/apps/jose/ja/server"
	cryptoutilAppsJoseJaServerConfig "cryptoutil/internal/apps/jose/ja/server/config"
	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilPoll "cryptoutil/internal/shared/util/poll"
)

var (
testJoseJAService *cryptoutilAppsJoseJaServer.JoseJAServer
sharedHTTPClient  *http.Client
publicBaseURL     string
adminBaseURL      string
)

func TestMain(m *testing.M) {
// Create in-memory SQLite configuration for testing.
cfg := cryptoutilAppsJoseJaServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

ctx := context.Background()

// Create server.
var err error

testJoseJAService, err = cryptoutilAppsJoseJaServer.NewFromConfig(ctx, cfg)
if err != nil {
panic(fmt.Sprintf("TestMain: failed to create server: %v", err))
}

	// Start server in background.
	errChan := make(chan error, 1)

	go func() {
		if startErr := testJoseJAService.Start(ctx); startErr != nil {
			errChan <- startErr
		}
	}()

	// Wait for server ports to be assigned.
	const (
		pollTimeout  = 5 * time.Second
		pollInterval = 100 * time.Millisecond
	)

	pollErr := cryptoutilSharedUtilPoll.Until(ctx, pollTimeout, pollInterval, func(_ context.Context) (bool, error) {
		select {
		case startErr := <-errChan:
			return false, fmt.Errorf("server failed to start: %w", startErr)
		default:
		}

		return testJoseJAService.PublicPort() > 0 && testJoseJAService.AdminPort() > 0, nil
	})
	if pollErr != nil {
		panic(fmt.Sprintf("TestMain: %v", pollErr))
	}

// Mark server as ready.
testJoseJAService.SetReady(true)

// Store base URLs for tests.
publicBaseURL = testJoseJAService.PublicBaseURL()
adminBaseURL = testJoseJAService.AdminBaseURL()

// Create shared HTTP client for all tests (accepts self-signed certs).
sharedHTTPClient = cryptoutilSharedCryptoTls.NewClientForTest()

// Run all tests.
exitCode := m.Run()

// Cleanup: Shutdown server.
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

_ = testJoseJAService.Shutdown(shutdownCtx)

os.Exit(exitCode)
}
