// Copyright (c) 2025 Justin Cranford

package check_chunk_verification

import (
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

func TestCheck_Integration(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-check-chunk-verification")
	err := Check(logger)

	require.NoError(t, err, "check-chunk-verification should pass on real project files")
}
