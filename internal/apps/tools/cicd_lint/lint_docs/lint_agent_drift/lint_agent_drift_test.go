// Copyright (c) 2025 Justin Cranford

package lint_agent_drift

import (
	"fmt"
	"io"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"

	"github.com/stretchr/testify/require"
)

func TestCheck_Integration(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-lint-agent-drift")
	err := Check(logger)

	require.NoError(t, err, "lint-agent-drift should pass on real project files")
}

// Sequential: mutates agentDriftFn package-level state.
func TestCheck_ErrorWithStderr(t *testing.T) {
	original := agentDriftFn

	t.Cleanup(func() { agentDriftFn = original })

	agentDriftFn = func(stdout, stderr io.Writer) int {
		_, _ = fmt.Fprint(stderr, "agent drift detail")

		return 1
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger)

	require.Error(t, err)
	require.Contains(t, err.Error(), "agent drift detail")
}

// Sequential: mutates agentDriftFn package-level state.
func TestCheck_ErrorWithoutStderr(t *testing.T) {
	original := agentDriftFn

	t.Cleanup(func() { agentDriftFn = original })

	agentDriftFn = func(stdout, stderr io.Writer) int {
		_, _ = fmt.Fprint(stdout, "some output")

		return 1
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger)

	require.Error(t, err)
	require.Contains(t, err.Error(), "agent drift violations found")
}

// Sequential: mutates agentDriftFn package-level state.
func TestCheck_Success(t *testing.T) {
	original := agentDriftFn

	t.Cleanup(func() { agentDriftFn = original })

	agentDriftFn = func(stdout, stderr io.Writer) int {
		_, _ = fmt.Fprint(stdout, "All agent pairs are in sync\n")

		return 0
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger)

	require.NoError(t, err)
}
