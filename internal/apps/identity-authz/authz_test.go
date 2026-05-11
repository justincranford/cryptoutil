// Copyright (c) 2025-2026 Justin Cranford.
package authz

import (
	"testing"

	cryptoutilTestCli "cryptoutil/internal/apps-framework/service/test_help_cli"
)

func TestAUTHZ(t *testing.T) {
	t.Parallel()

	cryptoutilTestCli.RunCLITests(t, Authz)
}
