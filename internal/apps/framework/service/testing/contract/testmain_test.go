// Copyright (c) 2025 Justin Cranford
//
// TestMain for contract package tests.
// Creates a skeleton-template server to run all contract tests against.
package contract

import (
	"context"
	"fmt"
	"os"
	"testing"

	cryptoutilAppsFrameworkServiceTestingE2eHelpers "cryptoutil/internal/apps/framework/service/testing/e2e_helpers"
	cryptoutilAppsSkeletonTemplateServer "cryptoutil/internal/apps/skeleton-template/server"
	cryptoutilAppsSkeletonTemplateServerConfig "cryptoutil/internal/apps/skeleton-template/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var testContractServer *cryptoutilAppsSkeletonTemplateServer.SkeletonTemplateServer

func TestMain(m *testing.M) {
	ctx := context.Background()

	cfg := cryptoutilAppsSkeletonTemplateServerConfig.DefaultTestConfig()

	var err error

	testContractServer, err = cryptoutilAppsSkeletonTemplateServer.NewFromConfig(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("TestMain: failed to create contract test server: %v", err))
	}

	cryptoutilAppsFrameworkServiceTestingE2eHelpers.MustStartAndWaitForDualPorts(testContractServer, func() error {
		return testContractServer.Start(ctx)
	})

	testContractServer.SetReady(true)

	exitCode := m.Run()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultDataServerShutdownTimeout)
	defer cancel()

	_ = testContractServer.Shutdown(shutdownCtx)

	os.Exit(exitCode)
}
