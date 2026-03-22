// Copyright (c) 2025 Justin Cranford

package validate_propagation

import (
	"fmt"
	"io"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"

	"github.com/stretchr/testify/require"
)

func TestCheck_Integration(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-validate-propagation")
	err := Check(logger)

	require.NoError(t, err, "validate-propagation should pass on real project files")
}

// Sequential: mutates validatePropagationFn package-level state.
func TestCheck_ErrorWithStderr(t *testing.T) {
	original := validatePropagationFn

	t.Cleanup(func() { validatePropagationFn = original })

	validatePropagationFn = func(stdout, stderr io.Writer) int {
		_, _ = fmt.Fprint(stderr, "broken reference error")

		return 1
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger)

	require.Error(t, err)
	require.Contains(t, err.Error(), "broken reference error")
}

// Sequential: mutates validatePropagationFn package-level state.
func TestCheck_ErrorWithoutStderr(t *testing.T) {
	original := validatePropagationFn

	t.Cleanup(func() { validatePropagationFn = original })

	validatePropagationFn = func(stdout, stderr io.Writer) int {
		_, _ = fmt.Fprint(stdout, "some output")

		return 1
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger)

	require.Error(t, err)
	require.Contains(t, err.Error(), "broken @source references")
}
