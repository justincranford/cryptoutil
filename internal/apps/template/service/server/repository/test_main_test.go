// Copyright (c) 2025 Justin Cranford

package repository_test

import (
	"os"
	"testing"

	cryptoutilAppsTemplateServiceServerTestutil "cryptoutil/internal/apps/template/service/server/testutil"
)

func TestMain(m *testing.M) {
	// Initialize shared test fixtures in testutil package.
	if err := cryptoutilAppsTemplateServiceServerTestutil.Initialize(); err != nil {
		panic("failed to initialize test fixtures: " + err.Error())
	}

	os.Exit(m.Run())
}
