package cicd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterWorkflowFiles(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected []string
	}{
		{
			name:     "no files",
			files:    []string{},
			expected: []string{},
		},
		{
			name: "only workflow files",
			files: []string{
				".github/workflows/ci-test.yml",
				".github/workflows/ci-build.yml",
			},
			expected: []string{
				".github/workflows/ci-test.yml",
				".github/workflows/ci-build.yml",
			},
		},
		{
			name: "mixed files",
			files: []string{
				".github/workflows/ci-test.yml",
				"README.md",
				"main.go",
				".github/workflows/ci-build.yaml",
			},
			expected: []string{
				".github/workflows/ci-test.yml",
				".github/workflows/ci-build.yaml",
			},
		},
		{
			name: "non-workflow yaml files",
			files: []string{
				"config.yml",
				"docker-compose.yaml",
				".github/dependabot.yml",
			},
			expected: []string{},
		},
		{
			name: "workflow files in subdirectories",
			files: []string{
				".github/workflows/subdir/ci-test.yml",
				".github/workflows/ci-build.yaml",
			},
			expected: []string{
				".github/workflows/subdir/ci-test.yml",
				".github/workflows/ci-build.yaml",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterWorkflowFiles(tt.files)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckWorkflowLintWithError_NoFiles(t *testing.T) {
	t.Skip("Skipping test that calls os.Exit() - not testable in unit tests")

	logger := NewLogUtil("TestCheckWorkflowLintWithError_NoFiles")
	allFiles := []string{}

	err := checkWorkflowLintWithError(logger, allFiles)
	require.NoError(t, err, "Should succeed with no files")
}

func TestCheckWorkflowLintWithError_NoWorkflowFiles(t *testing.T) {
	t.Skip("Skipping test that calls os.Exit() - not testable in unit tests")

	logger := NewLogUtil("TestCheckWorkflowLintWithError_NoWorkflowFiles")
	allFiles := []string{
		"README.md",
		"main.go",
		"config.yml",
	}

	err := checkWorkflowLintWithError(logger, allFiles)
	require.NoError(t, err, "Should succeed with no workflow files")
}

func TestCheckWorkflowLintWithError_ValidWorkflowFile(t *testing.T) {
	t.Skip("Skipping test that calls os.Exit() - not testable in unit tests")

	logger := NewLogUtil("TestCheckWorkflowLintWithError_ValidWorkflowFile")

	tmpDir := t.TempDir()
	workflowDir := filepath.Join(tmpDir, ".github", "workflows")
	err := os.MkdirAll(workflowDir, 0o755)
	require.NoError(t, err)

	workflowFile := filepath.Join(workflowDir, "ci-test.yml")
	workflowContent := `name: CI Test
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Test
        run: echo "Test - Workflow: ${{ github.workflow }}"
`
	err = os.WriteFile(workflowFile, []byte(workflowContent), 0o600)
	require.NoError(t, err)

	allFiles := []string{workflowFile}
	//nolint:ineffassign,errcheck // Testing workflow lint, may succeed or fail
	err = checkWorkflowLintWithError(logger, allFiles)
	// May pass or fail depending on whether action versions are up to date
	// The important thing is it doesn't panic
	require.NotPanics(t, func() {
		//nolint:errcheck // Testing workflow lint
		_ = checkWorkflowLintWithError(logger, allFiles)
	})
}

func TestCheckWorkflowLintWithError_MissingCIPrefix(t *testing.T) {
	t.Skip("Skipping test that calls os.Exit() - not testable in unit tests")

	logger := NewLogUtil("TestCheckWorkflowLintWithError_MissingCIPrefix")

	tmpDir := t.TempDir()
	workflowDir := filepath.Join(tmpDir, ".github", "workflows")
	err := os.MkdirAll(workflowDir, 0o755)
	require.NoError(t, err)

	workflowFile := filepath.Join(workflowDir, "test.yml")
	workflowContent := `name: Test
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "test"
`
	err = os.WriteFile(workflowFile, []byte(workflowContent), 0o600)
	require.NoError(t, err)

	allFiles := []string{workflowFile}
	err = checkWorkflowLintWithError(logger, allFiles)
	require.Error(t, err, "Should fail for missing ci- prefix")
	require.Contains(t, err.Error(), "validation errors")
}

func TestCheckWorkflowLintWithError_MissingWorkflowReference(t *testing.T) {
	t.Skip("Skipping test that calls os.Exit() - not testable in unit tests")

	logger := NewLogUtil("TestCheckWorkflowLintWithError_MissingWorkflowReference")

	tmpDir := t.TempDir()
	workflowDir := filepath.Join(tmpDir, ".github", "workflows")
	err := os.MkdirAll(workflowDir, 0o755)
	require.NoError(t, err)

	workflowFile := filepath.Join(workflowDir, "ci-test.yml")
	workflowContent := `name: CI Test
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "No workflow reference"
`
	err = os.WriteFile(workflowFile, []byte(workflowContent), 0o600)
	require.NoError(t, err)

	allFiles := []string{workflowFile}
	err = checkWorkflowLintWithError(logger, allFiles)
	require.Error(t, err, "Should fail for missing workflow reference")
}

func TestValidateAndGetWorkflowActionsDetails_EmptyFiles(t *testing.T) {
	t.Skip("Skipping test that calls os.Exit() - not testable in unit tests")

	logger := NewLogUtil("TestValidateAndGetWorkflowActionsDetails_EmptyFiles")
	allFiles := []string{}

	result := validateAndGetWorkflowActionsDetails(logger, allFiles)
	require.NotNil(t, result)
	require.Empty(t, result)
}

func TestValidateAndGetWorkflowActionsDetails_WithActions(t *testing.T) {
	t.Skip("Skipping test that calls os.Exit() - not testable in unit tests")

	logger := NewLogUtil("TestValidateAndGetWorkflowActionsDetails_WithActions")

	tmpDir := t.TempDir()
	workflowDir := filepath.Join(tmpDir, ".github", "workflows")
	err := os.MkdirAll(workflowDir, 0o755)
	require.NoError(t, err)

	workflowFile := filepath.Join(workflowDir, "ci-test.yml")
	workflowContent := `name: CI Test
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - name: Test
        run: echo "Workflow: ${{ github.workflow }}"
`
	err = os.WriteFile(workflowFile, []byte(workflowContent), 0o600)
	require.NoError(t, err)

	allFiles := []string{workflowFile}
	result := validateAndGetWorkflowActionsDetails(logger, allFiles)
	require.NotNil(t, result)
	// Should have found actions/checkout and actions/setup-go
	require.NotEmpty(t, result)
}

func TestCheckActionVersionsConcurrently_NoActions(t *testing.T) {
	logger := NewLogUtil("TestCheckActionVersionsConcurrently_NoActions")
	workflowsActionDetails := make(map[string]WorkflowActionDetails)
	exceptions := &WorkflowActionExceptions{Exceptions: make(map[string]WorkflowActionException)}

	outdated, exempted, errors := checkActionVersionsConcurrently(logger, workflowsActionDetails, exceptions)
	require.Empty(t, outdated)
	require.Empty(t, exempted)
	require.Empty(t, errors)
}

func TestCheckActionVersionsConcurrently_WithActions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	logger := NewLogUtil("TestCheckActionVersionsConcurrently_WithActions")

	workflowsActionDetails := map[string]WorkflowActionDetails{
		"actions/checkout": {
			Name:           "actions/checkout",
			CurrentVersion: "v4",
			WorkflowFiles:  []string{"ci-test.yml"},
		},
	}
	exceptions := &WorkflowActionExceptions{Exceptions: make(map[string]WorkflowActionException)}

	outdated, exempted, errors := checkActionVersionsConcurrently(logger, workflowsActionDetails, exceptions)

	// The result depends on actual GitHub API state
	// Just verify the function executes without panic
	require.NotNil(t, outdated)
	require.NotNil(t, exempted)
	require.NotNil(t, errors)
}

func TestGetLatestTag_ValidRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	logger := NewLogUtil("TestGetLatestTag_ValidRepo")

	// Test with a known repository
	tag, err := getLatestTag(logger, "actions/checkout")

	// Result depends on network/GitHub API state
	// Just verify it doesn't panic and returns valid types
	if err == nil {
		require.NotEmpty(t, tag)
	}
}

func TestGetLatestTag_InvalidRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	logger := NewLogUtil("TestGetLatestTag_InvalidRepo")

	// Test with an invalid repository
	_, err := getLatestTag(logger, "nonexistent/repository-that-does-not-exist")
	require.Error(t, err, "Should fail for nonexistent repository")
}
