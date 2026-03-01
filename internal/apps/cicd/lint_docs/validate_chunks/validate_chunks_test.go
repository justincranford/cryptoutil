// Copyright (c) 2025 Justin Cranford

package validate_chunks

import (
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

func TestCheck_Integration(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-validate-chunks")
	err := Check(logger)

	require.NoError(t, err, "validate-chunks should pass on real project files")
}
