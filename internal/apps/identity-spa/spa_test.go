// Copyright (c) 2025-2026 Justin Cranford.
package spa

import (
	"testing"

	cryptoutilTestCli "cryptoutil/internal/apps-framework/service/test_help_cli"
)

func TestSPA(t *testing.T) {
	t.Parallel()

	cryptoutilTestCli.RunCLITests(t, Spa)
}
