//go:build e2e

// Copyright (c) 2025-2026 Justin Cranford.
package e2e

import (
	"os"
	"testing"

	cryptoutilTestOrchE2e "cryptoutil/internal/apps-framework/service/test_orch_e2e"
)

func TestMain(m *testing.M) {
	os.Exit(cryptoutilTestOrchE2e.SetupE2ETestMain(m, cryptoutilTestOrchE2e.E2ETestConfig{}, nil))
}
