// Copyright (c) 2025-2026 Justin Cranford.
package ja

import (
	"testing"

	cryptoutilTestCli "cryptoutil/internal/apps-framework/service/test_help_cli"
)

func TestJA(t *testing.T) {
	t.Parallel()

	cryptoutilTestCli.RunCLITests(t, Ja)
}
