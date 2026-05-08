// Copyright (c) 2025-2026 Justin Cranford.
package rs

import (
	"testing"

	cryptoutilTestCli "cryptoutil/internal/apps-framework/service/testing/testcli"
)

func TestRS(t *testing.T) {
	t.Parallel()

	cryptoutilTestCli.RunCLITests(t, Rs)
}
