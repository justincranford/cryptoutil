// Copyright (c) 2025 Justin Cranford

package lint_skeleton

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

func TestLint_AllLintersPass(t *testing.T) {
	// NOT parallel: modifies the package-level registeredLinters seam variable.
	// Override registeredLinters with a no-op linter that always succeeds.
	origLinters := registeredLinters

	defer func() { registeredLinters = origLinters }()

	registeredLinters = []struct {
		name   string
		linter LinterFunc
	}{
		{"mock-pass-linter", func(_ *cryptoutilCmdCicdCommon.Logger) error { return nil }},
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	require.NoError(t, Lint(logger))
}

func TestLint_OneFailure(t *testing.T) {
	// NOT parallel: modifies the package-level registeredLinters seam variable.
	// Override registeredLinters with one failing linter.
	origLinters := registeredLinters

	defer func() { registeredLinters = origLinters }()

	registeredLinters = []struct {
		name   string
		linter LinterFunc
	}{
		{"mock-fail-linter", func(_ *cryptoutilCmdCicdCommon.Logger) error {
			return errors.New("mock linter failure")
		}},
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Lint(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "lint-skeleton failed")
	require.Contains(t, err.Error(), "mock-fail-linter")
}

func TestLint_MultipleFailuresCombined(t *testing.T) {
	// NOT parallel: modifies the package-level registeredLinters seam variable.
	// Override registeredLinters with two failing linters to verify error aggregation.
	origLinters := registeredLinters

	defer func() { registeredLinters = origLinters }()

	registeredLinters = []struct {
		name   string
		linter LinterFunc
	}{
		{"linter-a", func(_ *cryptoutilCmdCicdCommon.Logger) error { return errors.New("error-a") }},
		{"linter-b", func(_ *cryptoutilCmdCicdCommon.Logger) error { return errors.New("error-b") }},
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Lint(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "lint-skeleton failed")
}

func TestLint_EmptyLinterList(t *testing.T) {
	// NOT parallel: modifies the package-level registeredLinters seam variable.
	// An empty linter list should succeed with no errors.
	origLinters := registeredLinters

	defer func() { registeredLinters = origLinters }()

	registeredLinters = []struct {
		name   string
		linter LinterFunc
	}{}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	require.NoError(t, Lint(logger))
}
