// Copyright (c) 2025 Justin Cranford

package github_actions

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

const (
	testWorkflowWithActions = `name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
`
)

func TestLint_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger, FilterWorkflowFiles(map[string][]string{}))

	require.NoError(t, err, "Lint should succeed with no files")
}

func TestLint_NoWorkflowFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	// Files not in .github/workflows/.
	filesByExtension := map[string][]string{
		"go":   {"main.go"},
		"yaml": {"config.yaml"},
		"yml":  {"test.yml"},
	}

	err := Check(logger, FilterWorkflowFiles(filesByExtension))
	require.NoError(t, err, "Lint should succeed with no workflow files")
}

func TestFilterWorkflowFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    map[string][]string
		expected int
	}{
		{
			name:     "empty input",
			input:    map[string][]string{},
			expected: 0,
		},
		{
			name: "no workflow files",
			input: map[string][]string{
				"go":   {"main.go"},
				"yaml": {"config.yaml"},
			},
			expected: 0,
		},
		{
			name: "workflow files",
			input: map[string][]string{
				"yml":  {".github/workflows/ci.yml"},
				"yaml": {".github/workflows/cd.yaml"},
			},
			expected: 2,
		},
		{
			name: "mixed files",
			input: map[string][]string{
				"yml":  {".github/workflows/ci.yml"},
				"go":   {"main.go"},
				"yaml": {"config.yaml"},
			},
			expected: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := FilterWorkflowFiles(tc.input)
			require.Len(t, result, tc.expected)
		})
	}
}

func TestIsWorkflowFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"github workflow yml", ".github/workflows/ci.yml", true},
		{"github workflow yaml", ".github/workflows/ci.yaml", true},
		{"non-workflow yml", "config.yml", false},
		{"non-workflow yaml", "config.yaml", false},
		{"go file", "main.go", false},
		{"windows path", ".github\\workflows\\ci.yml", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := IsWorkflowFile(tc.path)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestValidateAndParseWorkflowFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workflowDir := filepath.Join(tmpDir, ".github", "workflows")

	err := os.MkdirAll(workflowDir, 0o755)
	require.NoError(t, err)

	workflowFile := filepath.Join(workflowDir, "ci.yml")
	err = os.WriteFile(workflowFile, []byte(testWorkflowWithActions), 0o600)
	require.NoError(t, err)

	actionDetails, validationErrors, err := validateAndParseWorkflowFile(workflowFile)

	require.NoError(t, err)
	require.Empty(t, validationErrors)
	require.Len(t, actionDetails, 2, "Should find 2 actions")
}

func TestContains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"contains", "hello world", "world", true},
		{"not contains", "hello world", "foo", false},
		{"empty substr", "hello", "", true},
		{"empty string", "", "foo", false},
		{"exact match", "foo", "foo", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := contains(tc.s, tc.substr)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestFindSubstring(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		s        string
		substr   string
		expected int
	}{
		{"found at beginning", "hello world", "hello", 0},
		{"found at end", "hello world", "world", 6},
		{"found in middle", "hello world", "lo wo", 3},
		{"not found", "hello world", "foo", -1},
		{"empty substr", "hello", "", 0},
		{"substr longer than string", "hi", "hello", -1},
		{"exact match", "foo", "foo", 0},
		{"multiple occurrences", "test test", "test", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := findSubstring(tc.s, tc.substr)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestLintGitHubWorkflows_NoActions(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()
	workflowDir := filepath.Join(tmpDir, ".github", "workflows")

	err := os.MkdirAll(workflowDir, 0o755)
	require.NoError(t, err)

	workflowFile := filepath.Join(workflowDir, "simple.yml")
	content := `name: Simple
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Hello World"
`
	err = os.WriteFile(workflowFile, []byte(content), 0o600)
	require.NoError(t, err)

	err = Check(logger, []string{workflowFile})
	require.NoError(t, err, "Lint should succeed with no actions")
}

func TestLintGitHubWorkflows_WithActions(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()
	workflowDir := filepath.Join(tmpDir, ".github", "workflows")

	err := os.MkdirAll(workflowDir, 0o755)
	require.NoError(t, err)

	workflowFile := filepath.Join(workflowDir, "ci.yml")
	err = os.WriteFile(workflowFile, []byte(testWorkflowWithActions), 0o600)
	require.NoError(t, err)

	err = Check(logger, []string{workflowFile})
	require.NoError(t, err, "Lint should succeed with actions (no outdated check)")
}

func TestLoadWorkflowActionExceptions_NotExists(t *testing.T) {
	// Note: Not parallel - uses relative path from current working directory.
	exceptions, err := loadWorkflowActionExceptions()
	require.NoError(t, err, "Should succeed when file doesn't exist")
	require.NotNil(t, exceptions)
	require.Empty(t, exceptions.Exceptions)
}

func TestLoadWorkflowActionExceptions_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)

	githubDir := filepath.Join(tmpDir, ".github")
	err = os.MkdirAll(githubDir, 0o755)
	require.NoError(t, err)

	exceptionsFile := filepath.Join(githubDir, "workflow-action-exceptions.json")
	err = os.WriteFile(exceptionsFile, []byte("invalid json"), 0o600)
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	defer func() {
		_ = os.Chdir(originalWd)
	}()

	exceptions, err := loadWorkflowActionExceptions()
	require.Error(t, err)
	require.Nil(t, exceptions)
	require.Contains(t, err.Error(), "failed to parse exceptions file")
}

func TestLoadWorkflowActionExceptions_ValidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)

	githubDir := filepath.Join(tmpDir, ".github")
	err = os.MkdirAll(githubDir, 0o755)
	require.NoError(t, err)

	exceptionsFile := filepath.Join(githubDir, "workflow-action-exceptions.json")
	content := `{"exceptions":{"actions/checkout":{"version":"v3","reason":"Test exception"}}}`
	err = os.WriteFile(exceptionsFile, []byte(content), 0o600)
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	defer func() {
		_ = os.Chdir(originalWd)
	}()

	exceptions, err := loadWorkflowActionExceptions()
	require.NoError(t, err)
	require.NotNil(t, exceptions)
	require.Len(t, exceptions.Exceptions, 1)
	require.Contains(t, exceptions.Exceptions, "actions/checkout")
	require.Equal(t, "v3", exceptions.Exceptions["actions/checkout"].Version)
	require.Equal(t, "Test exception", exceptions.Exceptions["actions/checkout"].Reason)
}

func TestLoadWorkflowActionExceptions_UnreadableFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.Chmod does not enforce POSIX permissions on Windows")
	}

	if os.Getuid() == 0 {
		t.Skip("Skipping test when running as root (root can read all files)")
	}

	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)

	githubDir := filepath.Join(tmpDir, ".github")
	err = os.MkdirAll(githubDir, 0o755)
	require.NoError(t, err)

	exceptionsFile := filepath.Join(githubDir, "workflow-action-exceptions.json")
	content := `{"exceptions":{}}`
	err = os.WriteFile(exceptionsFile, []byte(content), 0o000)
	require.NoError(t, err)

	defer func() {
		// Restore permissions so cleanup can delete the file.
		_ = os.Chmod(exceptionsFile, 0o644)
	}()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	defer func() {
		_ = os.Chdir(originalWd)
	}()

	exceptions, err := loadWorkflowActionExceptions()
	require.Error(t, err, "Should fail when file exists but is unreadable")
	require.Nil(t, exceptions)
	require.Contains(t, err.Error(), "failed to read exceptions file")
}

func TestValidateAndParseWorkflowFile_InvalidFile(t *testing.T) {
	t.Parallel()

	actionDetails, validationErrors, err := validateAndParseWorkflowFile("nonexistent.yml")
	require.Error(t, err)
	require.Nil(t, actionDetails)
	require.Nil(t, validationErrors)
	require.Contains(t, err.Error(), "failed to read workflow file")
}

func TestValidateAndParseWorkflowFile_MultipleActions(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workflowFile := filepath.Join(tmpDir, "test.yml")
	content := `name: Test
on: push
jobs:
  job1:
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
  job2:
    steps:
      - uses: actions/upload-artifact@v3
`
	err := os.WriteFile(workflowFile, []byte(content), 0o600)
	require.NoError(t, err)

	actionDetails, validationErrors, err := validateAndParseWorkflowFile(workflowFile)
	require.NoError(t, err)
	require.Empty(t, validationErrors)
	require.Len(t, actionDetails, 3)
	require.Contains(t, actionDetails, "actions/checkout@v4")
	require.Contains(t, actionDetails, "actions/setup-go@v5")
	require.Contains(t, actionDetails, "actions/upload-artifact@v3")
}

func TestCheckActionVersionsConcurrently_NoExceptions(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	actionDetails := map[string]WorkflowActionDetails{
		"actions/checkout@v4": {
			Name:           "actions/checkout",
			CurrentVersion: "v4",
			WorkflowFiles:  []string{"ci.yml"},
		},
	}
	exceptions := &WorkflowActionExceptions{Exceptions: make(map[string]WorkflowActionException)}

	outdated, exempted, errors := checkActionVersionsConcurrently(logger, actionDetails, exceptions)
	require.Empty(t, outdated)
	require.Empty(t, exempted)
	require.Empty(t, errors)
}

func TestCheckActionVersionsConcurrently_WithExemption(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	actionDetails := map[string]WorkflowActionDetails{
		"actions/checkout@v3": {
			Name:           "actions/checkout",
			CurrentVersion: "v3",
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
	require.Empty(t, outdated)
	require.Len(t, exempted, 1)
	require.Empty(t, errors)
	require.Equal(t, "actions/checkout", exempted[0].Name)
	require.Equal(t, "v3", exempted[0].CurrentVersion)
}

func TestValidateAndGetWorkflowActionsDetails_MultipleFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	workflowFile1 := filepath.Join(tmpDir, "ci.yml")
	content1 := `name: CI
on: push
jobs:
  build:
    steps:
      - uses: actions/checkout@v4
`
	err := os.WriteFile(workflowFile1, []byte(content1), 0o600)
	require.NoError(t, err)

	workflowFile2 := filepath.Join(tmpDir, "cd.yml")
	content2 := `name: CD
on: push
jobs:
  deploy:
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
`
	err = os.WriteFile(workflowFile2, []byte(content2), 0o600)
	require.NoError(t, err)

	actionDetails, err := validateAndGetWorkflowActionsDetails(logger, []string{workflowFile1, workflowFile2})
	require.NoError(t, err)
	require.Len(t, actionDetails, 2)
	require.Contains(t, actionDetails, "actions/checkout@v4")
	require.Contains(t, actionDetails, "actions/setup-go@v5")
	require.Len(t, actionDetails["actions/checkout@v4"].WorkflowFiles, 2, "Should merge workflow files for duplicate actions")
}
