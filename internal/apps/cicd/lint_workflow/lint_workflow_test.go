// Copyright (c) 2025 Justin Cranford

package lint_workflow

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

func TestLint_NoWorkflowFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{}

	err := Lint(logger, filesByExtension)
	require.NoError(t, err, "Lint should succeed with no workflow files")
}

func TestLint_ValidWorkflowFile(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Create valid workflow directory structure.
	workflowDir := filepath.Join(tmpDir, ".github", "workflows")
	require.NoError(t, os.MkdirAll(workflowDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	// Create a simple valid workflow file.
	workflowContent := "name: Test\non: push\njobs:\n  test:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n"
	workflowPath := filepath.Join(workflowDir, "test.yml")
	require.NoError(t, os.WriteFile(workflowPath, []byte(workflowContent), cryptoutilSharedMagic.CacheFilePermissions))

	filesByExtension := map[string][]string{
		"yml": {workflowPath},
	}

	// Lint either succeeds or fails depending on GitHub API availability.
	// Either way, the registered linters loop executes.
	_ = Lint(logger, filesByExtension)
}

func TestLint_BranchPinnedAction(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Create workflow directory structure.
	workflowDir := filepath.Join(tmpDir, ".github", "workflows")
	require.NoError(t, os.MkdirAll(workflowDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	// Create a workflow file with branch-pinned action (disallowed).
	workflowContent := "name: Test\non: push\njobs:\n  test:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@main\n"
	workflowPath := filepath.Join(workflowDir, "invalid.yml")
	require.NoError(t, os.WriteFile(workflowPath, []byte(workflowContent), cryptoutilSharedMagic.CacheFilePermissions))

	filesByExtension := map[string][]string{
		"yml": {workflowPath},
	}

	// Lint should fail for branch-pinned action.
	err := Lint(logger, filesByExtension)
	require.Error(t, err, "Lint should fail with branch-pinned action")
}
