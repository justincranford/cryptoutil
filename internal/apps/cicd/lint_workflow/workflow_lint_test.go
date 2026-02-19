// Copyright (c) 2025 Justin Cranford

package lint_workflow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)
func TestContains_EmptyStrings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"both empty", "", "", true},
		{"empty string non-empty substr", "", "a", false},
		{"non-empty string empty substr", "hello", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := contains(tc.s, tc.substr)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestFindSubstring_EmptyStrings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		s        string
		substr   string
		expected int
	}{
		{"both empty", "", "", 0},
		{"empty string non-empty substr", "", "a", -1},
		{"non-empty string empty substr", "hello", "", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := findSubstring(tc.s, tc.substr)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestLintGitHubWorkflows_MultipleErrors(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	workflowFiles := []string{
		"nonexistent1.yml",
		"nonexistent2.yml",
		"nonexistent3.yml",
	}

	err := lintGitHubWorkflows(logger, workflowFiles)
	require.Error(t, err)
	require.Contains(t, err.Error(), "workflow validation failed")
}

func TestValidateAndParseWorkflowFile_EmptyFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workflowFile := filepath.Join(tmpDir, "empty.yml")
	err := os.WriteFile(workflowFile, []byte(""), 0o600)
	require.NoError(t, err)

	actionDetails, validationErrors, err := validateAndParseWorkflowFile(workflowFile)
	require.NoError(t, err)
	require.Empty(t, validationErrors)
	require.Empty(t, actionDetails)
}

func TestValidateAndParseWorkflowFile_MalformedActions(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workflowFile := filepath.Join(tmpDir, "malformed.yml")
	content := `name: Malformed
on: push
jobs:
  test:
    steps:
      - uses: invalid-action
      - uses: actions/checkout
      - uses: actions/setup-go@
`
	err := os.WriteFile(workflowFile, []byte(content), 0o600)
	require.NoError(t, err)

	actionDetails, validationErrors, err := validateAndParseWorkflowFile(workflowFile)
	require.NoError(t, err)
	require.Empty(t, validationErrors)
	require.Empty(t, actionDetails, "Should not match malformed action references")
}

func TestCheckActionVersionsConcurrently_MultipleExemptions(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	actionDetails := map[string]WorkflowActionDetails{
		"actions/checkout@v3": {
			Name:           "actions/checkout",
			CurrentVersion: "v3",
			WorkflowFiles:  []string{"ci.yml"},
		},
		"actions/setup-go@v4": {
			Name:           "actions/setup-go",
			CurrentVersion: "v4",
			WorkflowFiles:  []string{"ci.yml"},
		},
	}
	exceptions := &WorkflowActionExceptions{
		Exceptions: map[string]WorkflowActionException{
			"actions/checkout": {
				Version: "v3",
				Reason:  "Exemption 1",
			},
			"actions/setup-go": {
				Version: "v4",
				Reason:  "Exemption 2",
			},
		},
	}

	outdated, exempted, errors := checkActionVersionsConcurrently(logger, actionDetails, exceptions)
	require.Empty(t, outdated)
	require.Len(t, exempted, 2)
	require.Empty(t, errors)
}

func TestIsWorkflowFile_CaseVariations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"uppercase YML", ".github/workflows/test.YML", false},
		{"uppercase YAML", ".github/workflows/test.YAML", false},
		{"mixed case yml", ".github/workflows/test.Yml", false},
		{"mixed case yaml", ".github/workflows/test.Yaml", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := isWorkflowFile(tc.path)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestFilterWorkflowFiles_DuplicateFiles(t *testing.T) {
	t.Parallel()

	filesByExtension := map[string][]string{
		"yml": {
			".github/workflows/ci.yml",
			".github/workflows/ci.yml",
			".github/workflows/cd.yml",
		},
	}

	result := filterWorkflowFiles(filesByExtension)
	require.Len(t, result, 3, "Should preserve duplicates as provided")
}

func TestValidateAndGetWorkflowActionsDetails_DuplicateActionsAcrossFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	workflowFile1 := filepath.Join(tmpDir, "ci1.yml")
	content1 := `name: CI1
on: push
jobs:
  build:
    steps:
      - uses: actions/checkout@v4
`
	err := os.WriteFile(workflowFile1, []byte(content1), 0o600)
	require.NoError(t, err)

	workflowFile2 := filepath.Join(tmpDir, "ci2.yml")
	content2 := `name: CI2
on: push
jobs:
  build:
    steps:
      - uses: actions/checkout@v4
`
	err = os.WriteFile(workflowFile2, []byte(content2), 0o600)
	require.NoError(t, err)

	actionDetails, err := validateAndGetWorkflowActionsDetails(logger, []string{workflowFile1, workflowFile2})
	require.NoError(t, err)
	require.Len(t, actionDetails, 1)
	require.Contains(t, actionDetails, "actions/checkout@v4")
	require.Len(t, actionDetails["actions/checkout@v4"].WorkflowFiles, 2, "Should merge workflow files list")
}

// TestLintGitHubWorkflows_WithExemptedActions tests the exempted actions reporting path
// by setting up an exceptions file and a workflow with matching exempted actions.
func TestLintGitHubWorkflows_WithExemptedActions(t *testing.T) {
	// Note: Not parallel - modifies working directory.
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	defer func() {
		_ = os.Chdir(originalWd)
	}()

	// Create .github directory with exceptions file.
	githubDir := filepath.Join(tmpDir, ".github")
	err = os.MkdirAll(githubDir, 0o755)
	require.NoError(t, err)

	// Create exceptions file with exempted action.
	exceptionsContent := `{
  "exceptions": {
    "actions/checkout": {
      "version": "v3",
      "reason": "Legacy compatibility required"
    }
  }
}`
	exceptionsFile := filepath.Join(githubDir, "workflow-action-exceptions.json")
	err = os.WriteFile(exceptionsFile, []byte(exceptionsContent), 0o600)
	require.NoError(t, err)

	// Create workflows directory with workflow using exempted action.
	workflowDir := filepath.Join(githubDir, "workflows")
	err = os.MkdirAll(workflowDir, 0o755)
	require.NoError(t, err)

	// Create workflow file using the exempted version.
	workflowContent := `name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`
	workflowFile := filepath.Join(workflowDir, "ci.yml")
	err = os.WriteFile(workflowFile, []byte(workflowContent), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = lintGitHubWorkflows(logger, []string{workflowFile})
	require.NoError(t, err, "Should succeed with exempted actions")
}

// TestLintGitHubWorkflows_SuccessPath tests the success message path when
// there are actions but none are exempted or outdated.
func TestLintGitHubWorkflows_SuccessPath(t *testing.T) {
	// Note: Not parallel - modifies working directory.
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	defer func() {
		_ = os.Chdir(originalWd)
	}()

	// Create .github/workflows directory.
	githubDir := filepath.Join(tmpDir, ".github")
	workflowDir := filepath.Join(githubDir, "workflows")
	err = os.MkdirAll(workflowDir, 0o755)
	require.NoError(t, err)

	// Create workflow file with actions (no exceptions file means no exemptions).
	workflowContent := `name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
`
	workflowFile := filepath.Join(workflowDir, "ci.yml")
	err = os.WriteFile(workflowFile, []byte(workflowContent), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = lintGitHubWorkflows(logger, []string{workflowFile})
	require.NoError(t, err, "Should succeed and print success message")
}

// TestLintGitHubWorkflows_ExemptedAndNonExemptedMixed tests when some actions
// are exempted and others are not.
func TestLintGitHubWorkflows_ExemptedAndNonExemptedMixed(t *testing.T) {
	// Note: Not parallel - modifies working directory.
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	defer func() {
		_ = os.Chdir(originalWd)
	}()

	// Create .github directory with exceptions file.
	githubDir := filepath.Join(tmpDir, ".github")
	err = os.MkdirAll(githubDir, 0o755)
	require.NoError(t, err)

	// Create exceptions file with one exempted action.
	exceptionsContent := `{
  "exceptions": {
    "actions/checkout": {
      "version": "v3",
      "reason": "Legacy compatibility"
    }
  }
}`
	exceptionsFile := filepath.Join(githubDir, "workflow-action-exceptions.json")
	err = os.WriteFile(exceptionsFile, []byte(exceptionsContent), 0o600)
	require.NoError(t, err)

	// Create workflows directory.
	workflowDir := filepath.Join(githubDir, "workflows")
	err = os.MkdirAll(workflowDir, 0o755)
	require.NoError(t, err)

	// Create workflow with both exempted and non-exempted actions.
	workflowContent := `name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v5
`
	workflowFile := filepath.Join(workflowDir, "ci.yml")
	err = os.WriteFile(workflowFile, []byte(workflowContent), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = lintGitHubWorkflows(logger, []string{workflowFile})
	require.NoError(t, err, "Should succeed with mixed exempted and non-exempted actions")
}
