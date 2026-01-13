// Copyright (c) 2025 Justin Cranford
//
//

package integration

import (
	"context"
	"net/http"
	"os"
	"testing"

	"cryptoutil/internal/apps/cipher/im/server"
	"cryptoutil/internal/apps/cipher/im/server/config"
	cipherTesting "cryptoutil/internal/apps/cipher/im/testing"
	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTLS "cryptoutil/internal/shared/crypto/tls"
)

// Shared test resources (initialized once per package).
var (
	sharedHTTPClient     *http.Client
	cipherImServer       *server.CipherIMServer
	testCipherIMServer   *config.CipherImServerSettings
	publicBaseURL        string
	adminBaseURL         string
	sharedServiceBaseURL string // Deprecated: use publicBaseURL.
)

// TestMain initializes cipher-im server with SQLite in-memory for fast integration tests.
// Integration tests start the full application but use SQLite instead of PostgreSQL,
// and exclude telemetry containers (otel-collector, grafana-lgtm).
func TestMain(m *testing.M) {
	settings := cryptoutilConfig.RequireNewForTest("cipher-im-integration-test")
	settings.DatabaseURL = "file::memory:?cache=shared" // SQLite in-memory for fast integration tests.

	testCipherIMServer = &config.CipherImServerSettings{
		ServiceTemplateServerSettings: settings,
	}

	cipherImServer = cipherTesting.StartCipherIMService(testCipherIMServer)

	defer func() {
		_ = cipherImServer.Shutdown(context.Background())
	}()

	publicBaseURL = cipherImServer.PublicBaseURL()
	adminBaseURL = cipherImServer.AdminBaseURL()
	sharedServiceBaseURL = publicBaseURL // Backward compatibility.
	sharedHTTPClient = cryptoutilTLS.NewClientForTest()

	exitCode := m.Run()

	os.Exit(exitCode)
}
