// Copyright (c) 2025 Justin Cranford

package lint_docs

import (
	"fmt"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"

	"github.com/stretchr/testify/require"
)

func TestLint_Integration(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-lint-docs")
	err := Lint(logger)

	require.NoError(t, err, "lint-docs should pass on real project files")
}

// Sequential: mutates registeredLinters package-level state.
func TestLint_SingleLinterError(t *testing.T) {
	original := registeredLinters

	t.Cleanup(func() { registeredLinters = original })

	registeredLinters = []struct {
		name   string
		linter LinterFunc
	}{
		{"test-linter", func(_ *cryptoutilCmdCicdCommon.Logger) error {
			return fmt.Errorf("test linter error")
		}},
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Lint(logger)

	require.Error(t, err)
	require.Contains(t, err.Error(), "lint-docs failed")
	require.Contains(t, err.Error(), "test linter error")
}

// Sequential: mutates registeredLinters package-level state.
func TestLint_MultipleLinterErrors(t *testing.T) {
	original := registeredLinters

	t.Cleanup(func() { registeredLinters = original })

	registeredLinters = []struct {
		name   string
		linter LinterFunc
	}{
		{"linter-one", func(_ *cryptoutilCmdCicdCommon.Logger) error {
			return fmt.Errorf("first error")
		}},
		{"linter-two", func(_ *cryptoutilCmdCicdCommon.Logger) error {
			return fmt.Errorf("second error")
		}},
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Lint(logger)

	require.Error(t, err)
	require.Contains(t, err.Error(), "lint-docs failed")
}
