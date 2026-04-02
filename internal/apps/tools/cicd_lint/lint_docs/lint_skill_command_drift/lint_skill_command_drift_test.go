// Copyright (c) 2025 Justin Cranford

package lint_skill_command_drift

import (
	"fmt"
	"io"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"

	"github.com/stretchr/testify/require"
)

func TestCheck_Integration(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-lint-skill-command-drift")
	err := Check(logger)

	require.NoError(t, err, "lint-skill-command-drift should pass on real project files")
}

// Sequential: mutates skillCommandDriftFn package-level state.
func TestCheck_ErrorWithStderr(t *testing.T) {
	original := skillCommandDriftFn

	t.Cleanup(func() { skillCommandDriftFn = original })

	skillCommandDriftFn = func(stdout, stderr io.Writer) int {
		_, _ = fmt.Fprint(stderr, "skill command drift detail")

		return 1
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger)

	require.Error(t, err)
	require.Contains(t, err.Error(), "skill command drift detail")
}

// Sequential: mutates skillCommandDriftFn package-level state.
func TestCheck_ErrorWithoutStderr(t *testing.T) {
	original := skillCommandDriftFn

	t.Cleanup(func() { skillCommandDriftFn = original })

	skillCommandDriftFn = func(stdout, stderr io.Writer) int {
		_, _ = fmt.Fprint(stdout, "some output")

		return 1
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger)

	require.Error(t, err)
	require.Contains(t, err.Error(), "skill/command drift violations found")
}

// Sequential: mutates skillCommandDriftFn package-level state.
func TestCheck_Success(t *testing.T) {
	original := skillCommandDriftFn

	t.Cleanup(func() { skillCommandDriftFn = original })

	skillCommandDriftFn = func(stdout, stderr io.Writer) int {
		_, _ = fmt.Fprint(stdout, "All skill/command pairs are in sync\n")

		return 0
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger)

	require.NoError(t, err)
}
