// Copyright (c) 2025 Justin Cranford

package lint_workflow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

// TestValidateAndParseWorkflowFile_BranchPinned verifies that branch-pinned
// versions (e.g. @main) are flagged as validation errors.
func TestValidateAndParseWorkflowFile_BranchPinned(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workflowFile := filepath.Join(tmpDir, "ci.yml")
	content := []byte("uses: actions/checkout@main\n")

	err := os.WriteFile(workflowFile, content, 0o600)
	require.NoError(t, err)

	actionDetails, validationErrors, err := validateAndParseWorkflowFile(workflowFile)
	require.NoError(t, err, "File read should succeed")
	require.NotEmpty(t, validationErrors, "Should flag branch-pinned version")
	require.Contains(t, validationErrors[0], "main")
	require.Contains(t, validationErrors[0], "branch")
	// Action is still parsed even though it has a validation error.
	require.Len(t, actionDetails, 1)
}

// TestValidateAndParseWorkflowFile_BranchPinned_AllBranchNames verifies all
// disallowed branch names are detected.
func TestValidateAndParseWorkflowFile_BranchPinned_AllBranchNames(t *testing.T) {
	t.Parallel()

	branchNames := []string{"main", "master", "latest", "develop", "dev", "trunk", "MAIN", "MASTER"}

	for _, branch := range branchNames {

		t.Run(branch, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			workflowFile := filepath.Join(tmpDir, "ci.yml")
			content := []byte("uses: actions/checkout@" + branch + "\n")

			err := os.WriteFile(workflowFile, content, 0o600)
			require.NoError(t, err)

			_, validationErrors, err := validateAndParseWorkflowFile(workflowFile)
			require.NoError(t, err)
			require.NotEmpty(t, validationErrors, "Branch %s should be flagged", branch)
		})
	}
}

// TestValidateAndGetWorkflowActionsDetails_BranchPinned verifies that branch-
// pinned actions cause validateAndGetWorkflowActionsDetails to return an error.
func TestValidateAndGetWorkflowActionsDetails_BranchPinned(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()
	workflowFile := filepath.Join(tmpDir, "ci.yml")
	content := []byte("uses: actions/checkout@main\n")

	err := os.WriteFile(workflowFile, content, 0o600)
	require.NoError(t, err)

	details, err := validateAndGetWorkflowActionsDetails(logger, []string{workflowFile})
	require.Error(t, err, "Should fail when branch-pinned actions are found")
	require.Nil(t, details)
	require.Contains(t, err.Error(), "1 workflow validation errors")
}

// TestLintGitHubWorkflows_BranchPinnedAction verifies that lintGitHubWorkflows
// returns an error when a workflow has branch-pinned actions.
// Note: Not parallel - changes working directory.
func TestLintGitHubWorkflows_BranchPinnedAction(t *testing.T) {
	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	defer func() {
		err := os.Chdir(origDir)
		require.NoError(t, err)
	}()

	workflowFile := filepath.Join(tmpDir, "ci.yml")
	content := []byte("uses: actions/checkout@main\n")

	err = os.WriteFile(workflowFile, content, 0o600)
	require.NoError(t, err)

	err = lintGitHubWorkflows(logger, []string{workflowFile})
	require.Error(t, err, "Should fail when branch-pinned actions are found")
	require.Contains(t, err.Error(), "workflow validation failed")
}

// TestLintGitHubWorkflows_ExceptionVersionMismatch verifies that
// lintGitHubWorkflows prints warnings when exception version mismatches.
// Note: Not parallel - changes working directory to set up exceptions file.
func TestLintGitHubWorkflows_ExceptionVersionMismatch(t *testing.T) {
	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Set up the .github directory with the exceptions file.
	githubDir := filepath.Join(tmpDir, ".github")

	err := os.MkdirAll(githubDir, 0o755)
	require.NoError(t, err)

	exceptionsFile := filepath.Join(githubDir, "workflow-action-exceptions.json")
	exceptionsContent := []byte(`{"exceptions":{"actions/checkout":{"version":"v3","reason":"Testing stale exception"}}}`)

	err = os.WriteFile(exceptionsFile, exceptionsContent, 0o600)
	require.NoError(t, err)

	// Change working directory so loadWorkflowActionExceptions finds the file.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	defer func() {
		err := os.Chdir(origDir)
		require.NoError(t, err)
	}()

	// Workflow uses v4 but exception specifies v3 - triggers stale-exception warning.
	workflowFile := filepath.Join(tmpDir, "ci.yml")
	content := []byte("uses: actions/checkout@v4\n")

	err = os.WriteFile(workflowFile, content, 0o600)
	require.NoError(t, err)

	err = lintGitHubWorkflows(logger, []string{workflowFile})
	require.NoError(t, err, "Stale exception warnings do not cause failure")
}
