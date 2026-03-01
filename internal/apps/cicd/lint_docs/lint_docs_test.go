// Copyright (c) 2025 Justin Cranford

package lint_docs

import (
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

func TestLint_Integration(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-lint-docs")
	err := Lint(logger)

	require.NoError(t, err, "lint-docs should pass on real project files")
}
