// Copyright (c) 2025-2026 Justin Cranford.
package ca

import (
	"testing"

	cryptoutilTestCli "cryptoutil/internal/apps-framework/service/testing/testcli"
)

func TestCA(t *testing.T) {
	t.Parallel()

	cryptoutilTestCli.RunCLITests(t, Ca)
}
