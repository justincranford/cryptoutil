// Copyright (c) 2025-2026 Justin Cranford.
package im

import (
	"testing"

	cryptoutilTestCli "cryptoutil/internal/apps-framework/service/testing/testcli"
)

func TestIM(t *testing.T) {
	t.Parallel()

	cryptoutilTestCli.RunCLITests(t, Im)
}
