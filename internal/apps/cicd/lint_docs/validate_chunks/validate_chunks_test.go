// Copyright (c) 2025 Justin Cranford

package validate_chunks

import (
	"fmt"
	"io"
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

// Sequential: mutates validateChunksFn package-level state.
func TestCheck_ErrorWithStderr(t *testing.T) {
	original := validateChunksFn

	t.Cleanup(func() { validateChunksFn = original })

	validateChunksFn = func(stdout, stderr io.Writer) int {
		_, _ = fmt.Fprint(stderr, "chunk mismatch error")

		return 1
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger)

	require.Error(t, err)
	require.Contains(t, err.Error(), "chunk mismatch error")
}

// Sequential: mutates validateChunksFn package-level state.
func TestCheck_ErrorWithoutStderr(t *testing.T) {
	original := validateChunksFn

	t.Cleanup(func() { validateChunksFn = original })

	validateChunksFn = func(stdout, stderr io.Writer) int {
		_, _ = fmt.Fprint(stdout, "some output")

		return 1
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger)

	require.Error(t, err)
	require.Contains(t, err.Error(), "out of sync")
}
