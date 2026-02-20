// Copyright (c) 2025 Justin Cranford

package github_actions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)
func TestLint_WithActualWorkflow(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()
	workflowDir := filepath.Join(tmpDir, ".github", "workflows")

	err := os.MkdirAll(workflowDir, 0o755)
	require.NoError(t, err)

	workflowFile := filepath.Join(workflowDir, "test.yml")
	content := `name: Test
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - run: go test ./...
`
	err = os.WriteFile(workflowFile, []byte(content), 0o600)
	require.NoError(t, err)

	filesByExtension := map[string][]string{
		"yml": {workflowFile},
	}

	err = Check(logger, FilterWorkflowFiles(filesByExtension))
	require.NoError(t, err, "Should succeed with valid workflow")
}

func TestValidateAndGetWorkflowActionsDetails_EmptyFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	actionDetails, err := validateAndGetWorkflowActionsDetails(logger, []string{})
	require.NoError(t, err)
	require.Empty(t, actionDetails)
}

func TestValidateAndGetWorkflowActionsDetails_FileReadError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	actionDetails, err := validateAndGetWorkflowActionsDetails(logger, []string{"nonexistent.yml"})
	require.Error(t, err)
	require.Nil(t, actionDetails)
	require.Contains(t, err.Error(), "workflow validation errors")
}

func TestLintGitHubWorkflows_EmptyFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger, []string{})
	require.NoError(t, err)
}

func TestLintGitHubWorkflows_ValidationError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger, []string{"nonexistent.yml"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "workflow validation failed")
}

func TestFilterWorkflowFiles_MixedExtensions(t *testing.T) {
	t.Parallel()

	filesByExtension := map[string][]string{
		"yml": {
			".github/workflows/ci.yml",
			".github/workflows/cd.yml",
			"config.yml",
		},
		"yaml": {
			".github/workflows/test.yaml",
			"settings.yaml",
		},
		"go": {"main.go"},
	}

	result := FilterWorkflowFiles(filesByExtension)
	require.Len(t, result, 3)
	require.Contains(t, result, ".github/workflows/ci.yml")
	require.Contains(t, result, ".github/workflows/cd.yml")
	require.Contains(t, result, ".github/workflows/test.yaml")
	require.NotContains(t, result, "config.yml")
	require.NotContains(t, result, "settings.yaml")
	require.NotContains(t, result, "main.go")
}

func TestIsWorkflowFile_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"short path yml", ".yml", false},
		{"short path yaml", ".yaml", false},
		{"github workflows only", ".github/workflows/", false},
		{"mixed slashes forward", ".github/workflows/test.yml", true},
		{"mixed slashes backward", ".github\\workflows\\test.yml", true},
		{"partial match", "my.github/workflows/test.yml", true},
		{"no extension", ".github/workflows/test", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := IsWorkflowFile(tc.path)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestCheckActionVersionsConcurrently_EmptyActions(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	actionDetails := map[string]WorkflowActionDetails{}
	exceptions := &WorkflowActionExceptions{Exceptions: make(map[string]WorkflowActionException)}

	outdated, exempted, errors := checkActionVersionsConcurrently(logger, actionDetails, exceptions)
	require.Empty(t, outdated)
	require.Empty(t, exempted)
	require.Empty(t, errors)
}

func TestValidateAndParseWorkflowFile_NoActions(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workflowFile := filepath.Join(tmpDir, "simple.yml")
	content := `name: Simple
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Hello"
`
	err := os.WriteFile(workflowFile, []byte(content), 0o600)
	require.NoError(t, err)

	actionDetails, validationErrors, err := validateAndParseWorkflowFile(workflowFile)
	require.NoError(t, err)
	require.Empty(t, validationErrors)
	require.Empty(t, actionDetails)
}

func TestValidateAndParseWorkflowFile_ComplexVersions(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workflowFile := filepath.Join(tmpDir, "versions.yml")
	content := `name: Versions
on: push
jobs:
  test:
    steps:
      - uses: actions/checkout@v4.1.0
      - uses: actions/setup-go@v5-rc1
      - uses: custom/action@1.2.3-alpha
`
	err := os.WriteFile(workflowFile, []byte(content), 0o600)
	require.NoError(t, err)

	actionDetails, validationErrors, err := validateAndParseWorkflowFile(workflowFile)
	require.NoError(t, err)
	require.Empty(t, validationErrors)
	require.Len(t, actionDetails, 3)
	require.Contains(t, actionDetails, "actions/checkout@v4.1.0")
	require.Contains(t, actionDetails, "actions/setup-go@v5-rc1")
	require.Contains(t, actionDetails, "custom/action@1.2.3-alpha")
}

func TestLint_ErrorPath(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {".github/workflows/nonexistent.yml"},
	}

	err := Check(logger, FilterWorkflowFiles(filesByExtension))
	require.Error(t, err)
}

func TestCheckActionVersionsConcurrently_NonExemptedVersion(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	actionDetails := map[string]WorkflowActionDetails{
		"actions/checkout@v4": {
			Name:           "actions/checkout",
			CurrentVersion: "v4",
			WorkflowFiles:  []string{"ci.yml"},
		},
	}
	exceptions := &WorkflowActionExceptions{
		Exceptions: map[string]WorkflowActionException{
			"actions/checkout": {
				Version: "v3",
				Reason:  "Testing",
			},
		},
	}

	outdated, exempted, errors := checkActionVersionsConcurrently(logger, actionDetails, exceptions)
	require.Empty(t, outdated, "Should have no outdated actions")
	require.Empty(t, exempted, "Should not be exempted (version mismatch)")
	// Version mismatch with exception triggers a stale-exception warning.
	require.Len(t, errors, 1, "Should have 1 stale-exception warning")
	require.Contains(t, errors[0], "actions/checkout@v4")
	require.Contains(t, errors[0], "v3")
}

func TestIsWorkflowFile_LongPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"deep nested yml", "repo/.github/workflows/ci/test.yml", true},
		{"deep nested yaml", "repo/.github/workflows/deploy/prod.yaml", true},
		{"yml but not workflows", "repo/.github/config.yml", false},
		{"yaml but not workflows", "repo/.github/settings.yaml", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := IsWorkflowFile(tc.path)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestFilterWorkflowFiles_EmptyExtensionKeys(t *testing.T) {
	t.Parallel()

	filesByExtension := map[string][]string{
		"go": {"main.go"},
		"md": {"README.md"},
	}

	result := FilterWorkflowFiles(filesByExtension)
	require.Empty(t, result, "Should return empty list when no yml/yaml files")
}

func TestFilterWorkflowFiles_NilInput(t *testing.T) {
	t.Parallel()

	result := FilterWorkflowFiles(nil)
	require.Empty(t, result, "Should handle nil input gracefully")
}

func TestLint_WithValidationError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {".github/workflows/invalid.yml"},
	}

	err := Check(logger, FilterWorkflowFiles(filesByExtension))
	require.Error(t, err)
	require.Contains(t, err.Error(), "lint-workflow failed")
}

func TestValidateAndGetWorkflowActionsDetails_PartialFailures(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	workflowFile1 := filepath.Join(tmpDir, "valid.yml")
	content1 := `name: Valid
on: push
jobs:
  build:
    steps:
      - uses: actions/checkout@v4
`
	err := os.WriteFile(workflowFile1, []byte(content1), 0o600)
	require.NoError(t, err)

	workflowFiles := []string{workflowFile1, "nonexistent.yml"}

	actionDetails, err := validateAndGetWorkflowActionsDetails(logger, workflowFiles)
	require.Error(t, err)
	require.Nil(t, actionDetails)
	require.Contains(t, err.Error(), "workflow validation errors")
}

func TestLintGitHubWorkflows_ExceptionLoadWarning(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	defer func() {
		_ = os.Chdir(originalWd)
	}()

	githubDir := filepath.Join(tmpDir, ".github")
	err = os.MkdirAll(githubDir, 0o755)
	require.NoError(t, err)

	exceptionsFile := filepath.Join(githubDir, "workflow-action-exceptions.json")
	err = os.WriteFile(exceptionsFile, []byte("invalid json"), 0o600)
	require.NoError(t, err)

	workflowDir := filepath.Join(githubDir, "workflows")
	err = os.MkdirAll(workflowDir, 0o755)
	require.NoError(t, err)

	workflowFile := filepath.Join(workflowDir, "test.yml")
	content := `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Hello"
`
	err = os.WriteFile(workflowFile, []byte(content), 0o600)
	require.NoError(t, err)

	err = Check(logger, []string{workflowFile})
	require.NoError(t, err, "Should succeed with invalid exceptions file (warning only)")
}
