// Copyright (c) 2025 Justin Cranford

package lint_workflow

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
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
	err := Lint(logger, map[string][]string{})

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

	err := Lint(logger, filesByExtension)
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

			result := filterWorkflowFiles(tc.input)
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

			result := isWorkflowFile(tc.path)
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

	err = lintGitHubWorkflows(logger, []string{workflowFile})
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

	err = lintGitHubWorkflows(logger, []string{workflowFile})
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

	err = Lint(logger, filesByExtension)
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
	err := lintGitHubWorkflows(logger, []string{})
	require.NoError(t, err)
}

func TestLintGitHubWorkflows_ValidationError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := lintGitHubWorkflows(logger, []string{"nonexistent.yml"})
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

	result := filterWorkflowFiles(filesByExtension)
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

			result := isWorkflowFile(tc.path)
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

	err := Lint(logger, filesByExtension)
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
	require.Empty(t, outdated)
	require.Empty(t, exempted)
	require.Empty(t, errors)
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

			result := isWorkflowFile(tc.path)
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

	result := filterWorkflowFiles(filesByExtension)
	require.Empty(t, result, "Should return empty list when no yml/yaml files")
}

func TestFilterWorkflowFiles_NilInput(t *testing.T) {
	t.Parallel()

	result := filterWorkflowFiles(nil)
	require.Empty(t, result, "Should handle nil input gracefully")
}

func TestLint_WithValidationError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {".github/workflows/invalid.yml"},
	}

	err := Lint(logger, filesByExtension)
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

	err = lintGitHubWorkflows(logger, []string{workflowFile})
	require.NoError(t, err, "Should succeed with invalid exceptions file (warning only)")
}

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
