// Copyright (c) 2025-2026 Justin Cranford.
package rs

import (
	"testing"

	cryptoutilTestCli "cryptoutil/internal/apps-framework/service/test_help_cli"
)

func TestRS(t *testing.T) {
	t.Parallel()

	cryptoutilTestCli.RunCLITests(t, Rs)
}
