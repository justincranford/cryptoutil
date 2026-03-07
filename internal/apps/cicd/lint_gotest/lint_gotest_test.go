// Copyright (c) 2025 Justin Cranford

package lint_gotest_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilCmdCicdLintGotest "cryptoutil/internal/apps/cicd/lint_gotest"
)

func TestLint_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := cryptoutilCmdCicdLintGotest.Lint(logger, map[string][]string{})

	require.NoError(t, err, "Lint should succeed with no files")
}

func TestLint_NoTestFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	// Files without _test.go suffix.
	filesByExtension := map[string][]string{
		"go":   {"main.go", "util.go"},
		"json": {"config.json"},
	}

	err := cryptoutilCmdCicdLintGotest.Lint(logger, filesByExtension)
	require.NoError(t, err, "Lint should succeed with no test files")
}

func TestLint_ValidTestFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a valid test file.
	testFile := filepath.Join(tmpDir, "example_test.go")
	content := `package example

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {
	t.Parallel()
	require.True(t, true)
}
`
	err := os.WriteFile(testFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {testFile},
	}

	err = cryptoutilCmdCicdLintGotest.Lint(logger, filesByExtension)

	require.NoError(t, err, "Lint should succeed with valid test file")
}
