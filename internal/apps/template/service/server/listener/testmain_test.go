// Copyright (c) 2025 Justin Cranford

package listener_test

import (
	"os"
	"testing"

	cryptoutilAppsTemplateServiceServerTestutil "cryptoutil/internal/apps/template/service/server/testutil"
)

// TestMain initializes shared test fixtures to avoid Windows firewall prompts from multiple server starts.
func TestMain(m *testing.M) {
	// Initialize shared test fixtures in testutil package (TLS configs, server settings).
	if err := cryptoutilAppsTemplateServiceServerTestutil.Initialize(); err != nil {
		panic("failed to initialize test fixtures: " + err.Error())
	}

	os.Exit(m.Run())
}
