// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"os"
	"testing"

	cryptoutilTestOrcIntegration "cryptoutil/internal/apps-framework/service/test_orch_integration"
	cryptoutilAppsIdentityAuthzServerConfig "cryptoutil/internal/apps/identity-authz/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	testServer            *AuthzServer
	testBaseURL           string
	testErr               error
	testIntegrationServer *cryptoutilTestOrcIntegration.IntegrationServer
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	cfg := cryptoutilAppsIdentityAuthzServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	testServer, testErr = NewFromConfig(ctx, cfg)
	if testErr != nil {
		panic("TestMain: failed to create test server: " + testErr.Error())
	}

	var err error

	testIntegrationServer, err = cryptoutilTestOrcIntegration.StartIntegrationServerForTestMain(ctx, testServer, nil)
	if err != nil {
		panic("TestMain: failed to start test server: " + err.Error())
	}

	testBaseURL = testServer.PublicBaseURL()

	exitCode := m.Run()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultDataServerShutdownTimeout)
	defer cancel()

	_ = testIntegrationServer.Shutdown(shutdownCtx)

	os.Exit(exitCode)
}

func requireTestSetup(t *testing.T) {
	t.Helper()

	if testErr != nil {
		t.Fatalf("Test setup failed: %v", testErr)
	}

	if testServer == nil {
		t.Fatal("Test server not initialized")
	}
}
