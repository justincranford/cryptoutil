// Copyright (c) 2025 Justin Cranford

package lint_fitness

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

func TestRegisteredLintersNotEmpty(t *testing.T) {
	t.Parallel()

	require.NotEmpty(t, registeredLinters, "At least one fitness linter must be registered")
}

// findProjectRoot finds the project root by looking for go.mod.
func findProjectRoot() (string, error) {
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
			return "", os.ErrNotExist
		}

		dir = parent
	}
}

// Sequential: uses os.Chdir (global process state).
func TestLint_Integration(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.
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

	logger := cryptoutilCmdCicdCommon.NewLogger("test-lint-fitness")

	// The actual project should pass all fitness checks.
	err = Lint(logger)
	require.NoError(t, err, "Project should pass all architecture fitness checks")
}
