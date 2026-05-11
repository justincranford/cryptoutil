// Copyright (c) 2025-2026 Justin Cranford.
package idp

import (
	"testing"

	cryptoutilTestCli "cryptoutil/internal/apps-framework/service/test_help_cli"
)

func TestIDP(t *testing.T) {
	t.Parallel()

	cryptoutilTestCli.RunCLITests(t, Idp)
}
