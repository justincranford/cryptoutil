// Copyright (c) 2025 Justin Cranford
//
//

package workflow

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestGetAvailableWorkflows(t *testing.T) {
	t.Parallel()
	// Create a temporary directory structure to simulate .github/workflows/
	tempDir := t.TempDir()
	workflowsDir := filepath.Join(tempDir, "test_workflows")

	// Create the workflows directory
	err := os.MkdirAll(workflowsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
	require.NoError(t, err, "Failed to create workflows directory")

	// Create some test workflow files
	testFiles := []string{
		"ci-quality.yml",
		"ci-coverage.yml",
		"ci-dast.yml",
		"ci-load.yml",
		"not-ci-file.yml", // Should be ignored (doesn't start with ci-)
		"ci-test.txt",     // Should be ignored (not .yml extension)
		"ci-race.yml",
	}

	for _, filename := range testFiles {
		filePath := filepath.Join(workflowsDir, filename)

		err := os.WriteFile(filePath, []byte("name: Test Workflow"), cryptoutilSharedMagic.FilePermOwnerReadWriteOnly)
		require.NoError(t, err, "Failed to create test file %s", filename)
	}

	// Create a subdirectory (should be ignored)
	subDir := filepath.Join(workflowsDir, "subdir")

	err = os.MkdirAll(subDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
	require.NoError(t, err, "Failed to create subdirectory")

	// Change to the temp directory to test the function
	originalWd, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")

	defer func() {
		if chdirErr := os.Chdir(originalWd); chdirErr != nil {
			require.NoError(t, chdirErr, "Failed to restore working directory")
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err, "Failed to change to temp directory")

	// Test the function
	workflows, err := getAvailableWorkflows(workflowsDir)
	require.NoError(t, err, "getAvailableWorkflows() returned error")

	// Verify expected workflows are found
	expectedWorkflows := map[string]bool{
		"quality":  true,
		"coverage": true,
		"dast":     true,
		"load":     true,
		"race":     true,
	}

	require.Equal(t, len(expectedWorkflows), len(workflows))

	for workflowName := range expectedWorkflows {
		require.Contains(t, workflows, workflowName)
	}

	// Verify unexpected workflows are not included
	unexpectedWorkflows := []string{"not-ci-file", "test"}
	for _, workflowName := range unexpectedWorkflows {
		require.NotContains(t, workflows, workflowName)
	}
}

func TestGetAvailableWorkflows_NoWorkflowsDir(t *testing.T) {
	t.Parallel()
	// Test the function - should return error
	_, err := getAvailableWorkflows(".github/workflows_nonexistent")
	require.Error(t, err, "Expected error when workflows directory doesn't exist")
}

func TestGetAvailableWorkflows_EmptyDir(t *testing.T) {
	t.Parallel()
	// Test with an empty directory
	tempDir := t.TempDir()
	workflowsDir := filepath.Join(tempDir, "empty_workflows")

	// Create the empty directory
	err := os.MkdirAll(workflowsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
	require.NoError(t, err, "Failed to create empty workflows directory")

	// Test the function
	workflows, err := getAvailableWorkflows(workflowsDir)
	require.Error(t, err, "getAvailableWorkflows() returned error")
	require.Nil(t, workflows, "Expected nil workflows map for empty directory")
}
