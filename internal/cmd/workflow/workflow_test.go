package workflow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAvailableWorkflows(t *testing.T) {
	// Create a temporary directory structure to simulate .github/workflows/
	tempDir := t.TempDir()
	workflowsDir := filepath.Join(tempDir, ".github", "workflows")

	// Create the workflows directory
	err := os.MkdirAll(workflowsDir, 0o755)
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

		err := os.WriteFile(filePath, []byte("name: Test Workflow"), 0o600)
		require.NoError(t, err, "Failed to create test file %s", filename)
	}

	// Create a subdirectory (should be ignored)
	subDir := filepath.Join(workflowsDir, "subdir")

	err = os.MkdirAll(subDir, 0o755)
	require.NoError(t, err, "Failed to create subdirectory")

	// Change to the temp directory to test the function
	originalWd, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")

	defer func() {
		if chdirErr := os.Chdir(originalWd); chdirErr != nil {
			assert.NoError(t, chdirErr, "Failed to restore working directory")
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err, "Failed to change to temp directory")

	// Test the function
	workflows, err := getAvailableWorkflows()
	require.NoError(t, err, "getAvailableWorkflows() returned error")

	// Verify expected workflows are found
	expectedWorkflows := map[string]bool{
		"quality":  true,
		"coverage": true,
		"dast":     true,
		"load":     true,
		"race":     true,
	}

	assert.Equal(t, len(expectedWorkflows), len(workflows))

	for workflowName := range expectedWorkflows {
		assert.Contains(t, workflows, workflowName)
	}

	// Verify unexpected workflows are not included
	unexpectedWorkflows := []string{"not-ci-file", "test"}
	for _, workflowName := range unexpectedWorkflows {
		assert.NotContains(t, workflows, workflowName)
	}
}

func TestGetAvailableWorkflows_NoWorkflowsDir(t *testing.T) {
	// Test with a directory that doesn't have .github/workflows/
	tempDir := t.TempDir()

	// Change to the temp directory
	originalWd, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")

	defer func() {
		if chdirErr := os.Chdir(originalWd); chdirErr != nil {
			assert.NoError(t, chdirErr, "Failed to restore working directory")
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err, "Failed to change to temp directory")

	// Test the function - should return error
	_, err = getAvailableWorkflows()
	assert.Error(t, err, "Expected error when workflows directory doesn't exist")
}

func TestGetAvailableWorkflows_EmptyDir(t *testing.T) {
	// Test with an empty .github/workflows/ directory
	tempDir := t.TempDir()
	workflowsDir := filepath.Join(tempDir, ".github", "workflows")

	err := os.MkdirAll(workflowsDir, 0o755)
	require.NoError(t, err, "Failed to create workflows directory")

	// Change to the temp directory
	originalWd, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")

	defer func() {
		if chdirErr := os.Chdir(originalWd); chdirErr != nil {
			assert.NoError(t, chdirErr, "Failed to restore working directory")
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err, "Failed to change to temp directory")

	// Test the function
	workflows, err := getAvailableWorkflows()
	require.NoError(t, err, "getAvailableWorkflows() returned error")

	// Should return empty map
	assert.Empty(t, workflows, "Expected empty workflows map")
}

func TestWorkflowsVariable(t *testing.T) {
	// Test that the workflows variable is properly initialized
	// This tests the package-level variable initialization
	assert.NotNil(t, workflows, "workflows variable should not be nil")

	// Should have at least some workflows (from the actual .github/workflows/ directory)
	assert.NotEmpty(t, workflows, "workflows variable should not be empty - no workflows found in .github/workflows/")

	// Verify all workflows have empty config structs
	for name, config := range workflows {
		// The config should be an empty struct
		assert.Equal(t, WorkflowConfig{}, config, "Workflow '%s' should have empty config", name)
	}
}
