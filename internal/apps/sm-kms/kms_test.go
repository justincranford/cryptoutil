// Copyright (c) 2025-2026 Justin Cranford.
package kms

import (
	"testing"

	cryptoutilTestCli "cryptoutil/internal/apps-framework/service/test_help_cli"
)

func TestKMS(t *testing.T) {
	t.Parallel()

	cryptoutilTestCli.RunCLITests(t, Kms)
}
