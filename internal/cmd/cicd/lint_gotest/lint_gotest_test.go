// Copyright (c) 2025 Justin Cranford

package lint_gotest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
	cryptoutilCmdCicdLintGotest "cryptoutil/internal/cmd/cicd/lint_gotest"
)

func TestLint_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := cryptoutilCmdCicdLintGotest.Lint(logger, []string{})

	require.NoError(t, err, "Lint should succeed with no files")
}

func TestLint_NoTestFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	files := []string{"main.go", "util.go", "config.json"}

	err := cryptoutilCmdCicdLintGotest.Lint(logger, files)
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
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = cryptoutilCmdCicdLintGotest.Lint(logger, []string{testFile})

	require.NoError(t, err, "Lint should succeed with valid test file")
}

func TestLint_UUIDNew(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "uuid_test.go")

	// File with uuid.New() which should be flagged.
	content := `package example

import (
	"testing"

	"github.com/google/uuid"
)

func TestUUID(t *testing.T) {
	id := uuid.New()
	_ = id
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = cryptoutilCmdCicdLintGotest.Lint(logger, []string{testFile})

	require.Error(t, err, "Should find uuid.New() issue")
}

func TestLint_TestifyWithoutImport(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "testify_test.go")

	// File using require. without testify import.
	content := `package example

import "testing"

func TestExample(t *testing.T) {
	require.True(t, true)
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = cryptoutilCmdCicdLintGotest.Lint(logger, []string{testFile})

	require.Error(t, err, "Should find testify import issue")
}
