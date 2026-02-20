// Copyright (c) 2025 Justin Cranford

package lint_go

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

func TestLint(t *testing.T) {
	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Lint(logger)

	require.Error(t, err, "Lint fails when go.mod not in current directory")
	require.Contains(t, err.Error(), "lint-go failed")
}

// findProjectRoot finds the project root by looking for go.mod.
func findProjectRoot() (string, error) {
	// Start from current directory and walk up.
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root.
			return "", os.ErrNotExist
		}

		dir = parent
	}
}

func TestLint_Integration(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.

	// Find and change to project root.
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping integration test - cannot find project root (no go.mod)")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(projectRoot))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-lint")

	// The actual project should pass all lint checks.
	err = Lint(logger)
	require.NoError(t, err, "Project should pass all lint checks")
}
