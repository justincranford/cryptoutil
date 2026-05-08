// Copyright (c) 2025-2026 Justin Cranford.
package template

import (
	"testing"

	cryptoutilTestCli "cryptoutil/internal/apps-framework/service/testing/testcli"
)

func TestTemplate(t *testing.T) {
	t.Parallel()

	cryptoutilTestCli.RunCLITests(t, Template)
}
