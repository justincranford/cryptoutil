// Copyright (c) 2025 Justin Cranford

package validate_coverage

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"

	"github.com/stretchr/testify/require"
)

func TestCheck_Integration(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-validate-coverage")
	err := Check(logger)

	require.NoError(t, err, "validate-coverage should pass on real project files")
}

func TestCheck_ErrorWithStderr(t *testing.T) {
	t.Parallel()

	stubFn := func(stdout, stderr io.Writer) int {
		_, _ = fmt.Fprint(stderr, "coverage error message")

		return 1
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := check(logger, stubFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "coverage error message")
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
	require.Contains(t, err.Error(), "required @propagate chunks are missing @source blocks")
}

func TestCheck_StdoutLogged(t *testing.T) {
	t.Parallel()

	stubFn := func(stdout, stderr io.Writer) int {
		_, _ = fmt.Fprint(stdout, "All required @propagate chunks are covered")

		return 0
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := check(logger, stubFn)

	require.NoError(t, err)
}

// Sequential: mutates os.Stderr (global process state, cannot run in parallel).
func TestCheck_EmptyStdout_NoLogCall(t *testing.T) {
	stubFn := func(_, _ io.Writer) int {
		return 0
	}

	oldStderr := os.Stderr

	r, w, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)

	os.Stderr = w

	t.Cleanup(func() { os.Stderr = oldStderr })

	logger := cryptoutilCmdCicdCommon.NewLogger("test-empty")
	checkErr := check(logger, stubFn)

	require.NoError(t, w.Close())

	captured, readErr := io.ReadAll(r)
	require.NoError(t, readErr)
	require.NoError(t, checkErr)

	cicdLines := strings.Count(string(captured), "[CICD]")
	require.Equal(t, 1, cicdLines, "only NewLogger start line expected; Log should not be called for empty stdout")
}
