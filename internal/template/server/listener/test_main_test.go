// Copyright (c) 2025 Justin Cranford

package listener_test

import (
	"os"
	"testing"

	cryptoutilTemplateServerTestutil "cryptoutil/internal/template/server/testutil"
)

func TestMain(m *testing.M) {
	// Initialize shared test fixtures in testutil package.
	if err := cryptoutilTemplateServerTestutil.Initialize(); err != nil {
		panic("failed to initialize test fixtures: " + err.Error())
	}

	os.Exit(m.Run())
}
