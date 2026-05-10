// Copyright (c) 2025-2026 Justin Cranford.
//
// TestMain for SM-IM client tests.

package client

import (
	"context"
	"crypto/tls"
	http "net/http"
	"os"
	"testing"
	"time"

	cryptoutilTestOrcIntegration "cryptoutil/internal/apps-framework/service/test_orch_integration"
	cryptoutilAppsSmImServer "cryptoutil/internal/apps/sm-im/server"
	cryptoutilAppsSmImServerConfig "cryptoutil/internal/apps/sm-im/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Shared test resources (initialized once per package).
var (
	sharedHTTPClient     *http.Client
	smIMServer           *cryptoutilAppsSmImServer.SmIMServer
	testIntegrationServer *cryptoutilTestOrcIntegration.IntegrationServer
	publicBaseURL        string
	adminBaseURL         string
	sharedServiceBaseURL string // Deprecated: use publicBaseURL.
)

// TestMain initializes sm-im server with SQLite in-memory for fast tests.
// Tests start the full application but use SQLite instead of PostgreSQL,
// and exclude telemetry containers (otel-collector, grafana-lgtm).
func TestMain(m *testing.M) {
	ctx := context.Background()

	cfg := cryptoutilAppsSmImServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	var err error

	smIMServer, err = cryptoutilAppsSmImServer.NewIMServerFromConfig(ctx, cfg)
	if err != nil {
		panic("TestMain: failed to create server: " + err.Error())
	}

	testIntegrationServer, err = cryptoutilTestOrcIntegration.StartIntegrationServerForTestMain(ctx, smIMServer, nil)
	if err != nil {
		panic("TestMain: failed to start server: " + err.Error())
	}

	publicBaseURL = smIMServer.PublicBaseURL()
	adminBaseURL = smIMServer.AdminBaseURL()
	sharedServiceBaseURL = publicBaseURL // Backward compatibility.
	// Create shared HTTP client using proper TLS certificate validation.
	sharedHTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				RootCAs:    smIMServer.TLSRootCAPool(),
			},
			DisableKeepAlives: true,
		},
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
	}

	exitCode := m.Run()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultDataServerShutdownTimeout)
	defer cancel()

	_ = testIntegrationServer.Shutdown(shutdownCtx)

	os.Exit(exitCode)
}
