//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford
//
//

package integration

import (
	"context"
	"crypto/tls"
	http "net/http"
	"os"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilAppsSmImServer "cryptoutil/internal/apps/sm-im/server"
	cryptoutilAppsSmImServerConfig "cryptoutil/internal/apps/sm-im/server/config"
	cryptoutilAppsSmImTesting "cryptoutil/internal/apps/sm-im/testing"
	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
)

// Shared test resources (initialized once per package).
var (
	sharedHTTPClient     *http.Client
	smIMServer           *cryptoutilAppsSmImServer.SmIMServer
	testSmIMServer       *cryptoutilAppsSmImServerConfig.SmIMServerSettings
	publicBaseURL        string
	adminBaseURL         string
	sharedServiceBaseURL string // Deprecated: use publicBaseURL.
)

// TestMain initializes sm-im server with SQLite in-memory for fast integration tests.
// Integration tests start the full application but use SQLite instead of PostgreSQL,
// and exclude telemetry containers (otel-collector, grafana-lgtm).
func TestMain(m *testing.M) {
	settings := cryptoutilAppsFrameworkServiceConfig.RequireNewForTest("sm-im-integration-test")
	settings.DatabaseURL = cryptoutilSharedMagic.SQLiteInMemoryDSN // SQLite in-memory for fast integration tests.

	testSmIMServer = &cryptoutilAppsSmImServerConfig.SmIMServerSettings{
		ServiceFrameworkServerSettings: settings,
	}

	smIMServer = cryptoutilAppsSmImTesting.StartSmIMService(testSmIMServer)

	defer func() {
		_ = smIMServer.Shutdown(context.Background())
	}()

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

	os.Exit(exitCode)
}
