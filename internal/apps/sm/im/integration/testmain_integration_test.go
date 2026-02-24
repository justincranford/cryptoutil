// Copyright (c) 2025 Justin Cranford
//
//

package integration

import (
	"context"
	http "net/http"
	"os"
	"testing"

	cryptoutilAppsSmImServer "cryptoutil/internal/apps/sm/im/server"
	cryptoutilAppsSmImServerConfig "cryptoutil/internal/apps/sm/im/server/config"
	cryptoutilAppsSmImTesting "cryptoutil/internal/apps/sm/im/testing"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
)

// Shared test resources (initialized once per package).
var (
	sharedHTTPClient     *http.Client
	smIMServer       *cryptoutilAppsSmImServer.SmIMServer
	testSmIMServer   *cryptoutilAppsSmImServerConfig.SmIMServerSettings
	publicBaseURL        string
	adminBaseURL         string
	sharedServiceBaseURL string // Deprecated: use publicBaseURL.
)

// TestMain initializes sm-im server with SQLite in-memory for fast integration tests.
// Integration tests start the full application but use SQLite instead of PostgreSQL,
// and exclude telemetry containers (otel-collector, grafana-lgtm).
func TestMain(m *testing.M) {
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("sm-im-integration-test")
	settings.DatabaseURL = "file::memory:?cache=shared" // SQLite in-memory for fast integration tests.

	testSmIMServer = &cryptoutilAppsSmImServerConfig.SmIMServerSettings{
		ServiceTemplateServerSettings: settings,
	}

	smIMServer = cryptoutilAppsSmImTesting.StartSmIMService(testSmIMServer)

	defer func() {
		_ = smIMServer.Shutdown(context.Background())
	}()

	publicBaseURL = smIMServer.PublicBaseURL()
	adminBaseURL = smIMServer.AdminBaseURL()
	sharedServiceBaseURL = publicBaseURL // Backward compatibility.
	sharedHTTPClient = cryptoutilSharedCryptoTls.NewClientForTest()

	exitCode := m.Run()

	os.Exit(exitCode)
}
