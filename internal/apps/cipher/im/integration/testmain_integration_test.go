// Copyright (c) 2025 Justin Cranford
//
//

package integration

import (
	"context"
	http "net/http"
	"os"
	"testing"

	cryptoutilAppsCipherImServer "cryptoutil/internal/apps/cipher/im/server"
	cryptoutilAppsCipherImServerConfig "cryptoutil/internal/apps/cipher/im/server/config"
	cryptoutilAppsCipherImTesting "cryptoutil/internal/apps/cipher/im/testing"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
)

// Shared test resources (initialized once per package).
var (
	sharedHTTPClient     *http.Client
	cipherImServer       *cryptoutilAppsCipherImServer.CipherIMServer
	testCipherIMServer   *cryptoutilAppsCipherImServerConfig.CipherImServerSettings
	publicBaseURL        string
	adminBaseURL         string
	sharedServiceBaseURL string // Deprecated: use publicBaseURL.
)

// TestMain initializes cipher-im server with SQLite in-memory for fast integration tests.
// Integration tests start the full application but use SQLite instead of PostgreSQL,
// and exclude telemetry containers (otel-collector, grafana-lgtm).
func TestMain(m *testing.M) {
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("cipher-im-integration-test")
	settings.DatabaseURL = "file::memory:?cache=shared" // SQLite in-memory for fast integration tests.

	testCipherIMServer = &cryptoutilAppsCipherImServerConfig.CipherImServerSettings{
		ServiceTemplateServerSettings: settings,
	}

	cipherImServer = cryptoutilAppsCipherImTesting.StartCipherIMService(testCipherIMServer)

	defer func() {
		_ = cipherImServer.Shutdown(context.Background())
	}()

	publicBaseURL = cipherImServer.PublicBaseURL()
	adminBaseURL = cipherImServer.AdminBaseURL()
	sharedServiceBaseURL = publicBaseURL // Backward compatibility.
	sharedHTTPClient = cryptoutilSharedCryptoTls.NewClientForTest()

	exitCode := m.Run()

	os.Exit(exitCode)
}
