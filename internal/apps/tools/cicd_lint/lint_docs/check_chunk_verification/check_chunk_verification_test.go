// Copyright (c) 2025 Justin Cranford

package check_chunk_verification

import (
	"fmt"
	"io"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"

	"github.com/stretchr/testify/require"
)

func TestCheck_Integration(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-check-chunk-verification")
	err := Check(logger)

	require.NoError(t, err, "check-chunk-verification should pass on real project files")
}

func TestCheck_ErrorWithStderr(t *testing.T) {
	t.Parallel()

	stubFn := func(stdout, stderr io.Writer) int {
		_, _ = fmt.Fprint(stderr, "some stderr error")

		return 1
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := check(logger, stubFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "some stderr error")
}

func TestCheck_ErrorWithoutStderr(t *testing.T) {
	t.Parallel()

	stubFn := func(stdout, stderr io.Writer) int {
		_, _ = fmt.Fprint(stdout, "some output")

		return 1
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := check(logger, stubFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "missing chunk references")
}
