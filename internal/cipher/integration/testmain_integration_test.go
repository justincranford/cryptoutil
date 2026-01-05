// Copyright (c) 2025 Justin Cranford
//
//

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"

	"cryptoutil/internal/cipher/server"
	"cryptoutil/internal/cipher/server/config"
	cipherTesting "cryptoutil/internal/cipher/testing"
	cryptoutilConfig "cryptoutil/internal/shared/config"
)

// Shared test resources (initialized once per package).
var (
	cipherImServer       *server.CipherIMServer
	sharedAppConfig      *config.AppConfig
	sharedServiceBaseURL string
)

// TestMain initializes cipher-im server with automatic PostgreSQL testcontainer provisioning.
// Service-template handles container lifecycle, database connection, and cleanup automatically.
func TestMain(m *testing.M) {
	settings := cryptoutilConfig.RequireNewForTest("cipher-im-integration-test")
	settings.DatabaseURL = ""               // Empty = use testcontainer.
	settings.DatabaseContainer = "required" // Require PostgreSQL testcontainer.

	sharedAppConfig = &config.AppConfig{
		ServerSettings: *settings,
		JWTSecret:      uuid.Must(uuid.NewUUID()).String(),
	}

	cipherImServer = cipherTesting.StartCipherIMServer(sharedAppConfig)

	sharedServiceBaseURL = fmt.Sprintf("%s://%s:%d", sharedAppConfig.BindPublicProtocol, sharedAppConfig.BindPublicAddress, cipherImServer.PublicPort())

	exitCode := m.Run()

	_ = cipherImServer.Shutdown(context.Background())

	os.Exit(exitCode)
}
