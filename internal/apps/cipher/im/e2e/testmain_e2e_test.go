// Copyright (c) 2025 Justin Cranford
//
//

package e2e_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/google/uuid"

	cryptoutilCipherServer "cryptoutil/internal/apps/cipher/im/server"
	"cryptoutil/internal/apps/cipher/im/server/config"
	cipherTesting "cryptoutil/internal/apps/cipher/im/testing"
	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTLS "cryptoutil/internal/shared/crypto/tls"
)

// Shared test resources (initialized once per package).
var (
	sharedHTTPClient   *http.Client
	testCipherIMServer *cryptoutilCipherServer.CipherIMServer
	sharedAppConfig    *config.AppConfig
	publicBaseURL      string
	adminBaseURL       string
)

// TestMain initializes cipher-im server with SQLite in-memory for fast E2E tests.
// Service-template handles database, telemetry, and all infrastructure automatically.
func TestMain(m *testing.M) {
	settings := cryptoutilConfig.RequireNewForTest("cipher-im-e2e-test")
	settings.DatabaseURL = "file::memory:?cache=shared" // SQLite in-memory for fast E2E tests.

	sharedAppConfig = &config.AppConfig{
		ServiceTemplateServerSettings: *settings,
		JWTSecret:      uuid.Must(uuid.NewUUID()).String(),
	}

	testCipherIMServer = cipherTesting.StartCipherIMService(sharedAppConfig)

	defer func() {
		_ = testCipherIMServer.Shutdown(context.Background())
	}()

	publicBaseURL = testCipherIMServer.PublicBaseURL()
	adminBaseURL = testCipherIMServer.AdminBaseURL()
	sharedHTTPClient = cryptoutilTLS.NewClientForTest()

	exitCode := m.Run()

	os.Exit(exitCode)
}
