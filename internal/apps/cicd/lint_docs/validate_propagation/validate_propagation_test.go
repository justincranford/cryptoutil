// Copyright (c) 2025 Justin Cranford

package validate_propagation

import (
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

func TestCheck_Integration(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-validate-propagation")
	err := Check(logger)

	require.NoError(t, err, "validate-propagation should pass on real project files")
}
