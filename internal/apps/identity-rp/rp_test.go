// Copyright (c) 2025-2026 Justin Cranford.
package rp

import (
	"testing"

	cryptoutilTestCli "cryptoutil/internal/apps-framework/service/testing/testcli"
)

func TestRP(t *testing.T) {
	t.Parallel()

	cryptoutilTestCli.RunCLITests(t, Rp)
}
