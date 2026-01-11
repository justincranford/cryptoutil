// Copyright (c) 2025 Justin Cranford
//
//

package integration

import (
	"context"
	"os"
	"testing"

	"cryptoutil/internal/apps/cipher/im/server"
	"cryptoutil/internal/apps/cipher/im/server/config"
	cipherTesting "cryptoutil/internal/apps/cipher/im/testing"
	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
)

// Shared test resources (initialized once per package).
var (
	cipherImServer       *server.CipherIMServer
	testCipherIMServer   *config.CipherImServerSettings
	sharedServiceBaseURL string
)

// TestMain initializes cipher-im server with automatic PostgreSQL testcontainer provisioning.
// Service-template handles container lifecycle, database connection, and cleanup automatically.
func TestMain(m *testing.M) {
	settings := cryptoutilConfig.RequireNewForTest("cipher-im-integration-test")
	settings.DatabaseURL = ""               // Empty = use testcontainer.
	settings.DatabaseContainer = "required" // Require PostgreSQL testcontainer.

	testCipherIMServer = &config.CipherImServerSettings{
		ServiceTemplateServerSettings: *settings,
	}

	cipherImServer = cipherTesting.StartCipherIMService(testCipherIMServer)

	defer func() {
		_ = cipherImServer.Shutdown(context.Background())
	}()

	sharedServiceBaseURL = cipherImServer.PublicBaseURL()

	exitCode := m.Run()

	os.Exit(exitCode)
}
