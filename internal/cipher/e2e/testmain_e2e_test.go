// Copyright (c) 2025 Justin Cranford
//
//

package e2e_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/google/uuid"

	cryptoutilCipherServer "cryptoutil/internal/cipher/server"
	"cryptoutil/internal/cipher/server/config"
	cipherTesting "cryptoutil/internal/cipher/testing"
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// Shared test resources (initialized once per package).
var (
	sharedHTTPClient   *http.Client
	testCipherIMServer *cryptoutilCipherServer.CipherIMServer
	sharedAppConfig    *config.AppConfig
	baseURL            string
	adminURL           string
)

// TestMain initializes cipher-im server with SQLite in-memory for fast E2E tests.
// Service-template handles database, telemetry, and all infrastructure automatically.
func TestMain(m *testing.M) {
	settings := cryptoutilConfig.RequireNewForTest("cipher-im-e2e-test")
	settings.DatabaseURL = "file::memory:?cache=shared" // SQLite in-memory for fast E2E tests.

	sharedAppConfig = &config.AppConfig{
		ServerSettings: *settings,
		JWTSecret:      uuid.Must(uuid.NewUUID()).String(),
	}

	testCipherIMServer = cipherTesting.StartCipherIMServer(sharedAppConfig)

	baseURL = fmt.Sprintf("%s://%s:%d", sharedAppConfig.BindPublicProtocol, sharedAppConfig.BindPublicAddress, testCipherIMServer.PublicPort())

	adminPortValue, _ := testCipherIMServer.AdminPort()
	adminURL = fmt.Sprintf("%s://%s:%d", sharedAppConfig.BindPrivateProtocol, sharedAppConfig.BindPrivateAddress, adminPortValue)

	// Create HTTP client with test TLS config.
	sharedHTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment only.
			},
		},
		Timeout: cryptoutilMagic.CipherDefaultTimeout,
	}

	exitCode := m.Run()

	_ = testCipherIMServer.Shutdown(context.Background())

	os.Exit(exitCode)
}
