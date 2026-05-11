// Copyright (c) 2025-2026 Justin Cranford.
package im

import (
	"testing"

	cryptoutilTestCli "cryptoutil/internal/apps-framework/service/test_help_cli"
)

func TestIM(t *testing.T) {
	t.Parallel()

	cryptoutilTestCli.RunCLITests(t, Im)
}
