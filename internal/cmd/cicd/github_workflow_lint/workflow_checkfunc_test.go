// Copyright (c) 2025 Justin Cranford
//
//

package github_workflow_lint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/cmd/cicd/common"
	cryptoutilTestutil "cryptoutil/internal/common/testutil"
)

func TestLint_NoActions(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("test_no_actions")
	tempDir := t.TempDir()

	// Workflow file with no actions (only run commands)
	workflowContent := `name: No Actions Workflow
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Run command
        run: echo "No actions used"
`
	workflowPath := cryptoutilTestutil.WriteTempFile(t, tempDir, "ci-test.yml", workflowContent)

	err := Lint(logger, []string{workflowPath})
	require.NoError(t, err, "Expected no error when no actions are used")
}

func TestLint_AllUpToDate(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("test_all_uptodate")
	tempDir := t.TempDir()

	// Create workflow file with current version of actions
	workflowContent := `name: Up To Date Actions
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
`
	workflowPath := cryptoutilTestutil.WriteTempFile(t, tempDir, "ci-uptodate.yml", workflowContent)

	// This test assumes GitHub API returns v4 and v5 as latest versions.
	// In real scenarios, this might be outdated. We test the logic flow, not actual API responses.
	err := Lint(logger, []string{workflowPath})

	// We expect either no error (if versions are current) or an outdated error (if they're old).
	// The test validates that the function completes without panicking.
	if err != nil {
		require.Contains(t, err.Error(), "outdated", "Expected error to mention outdated actions if versions are old")
	}
}

func TestLint_ExemptedActions(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("test_exempted")
	tempDir := t.TempDir()

	// Create exception file
	exceptionsContent := `{
  "exceptions": {
    "actions/checkout": {
      "allowed_version": "v3",
      "reason": "Testing older version"
    }
  }
}
`
	exceptionsPath := filepath.Join(tempDir, ".github", "workflows-outdated-action-exemptions.json")
	err := os.MkdirAll(filepath.Dir(exceptionsPath), 0o755)
	require.NoError(t, err, "Failed to create .github directory")

	err = os.WriteFile(exceptionsPath, []byte(exceptionsContent), 0o600)
	require.NoError(t, err, "Failed to write exceptions file")

	// Change to temp directory so exception file is found
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: failed to restore directory: %v", err)
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Create workflow file with exempted action
	workflowContent := `name: Exempted Actions
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`
	workflowPath := cryptoutilTestutil.WriteTempFile(t, tempDir, "ci-exempted.yml", workflowContent)

	err = Lint(logger, []string{workflowPath})

	// Exempted actions should not cause errors
	// (though they may trigger warnings about exemptions)
	if err != nil {
		require.NotContains(t, err.Error(), "actions/checkout", "Exempted action should not cause error")
	}
}

func TestLint_InvalidWorkflowFile(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("test_invalid")
	tempDir := t.TempDir()

	// Invalid YAML content - function doesn't strictly parse YAML, just does regex matching
	// So this won't cause an error, but it will have validation issues for missing name/logging
	invalidContent := `this is not valid YAML: {{{ [[[`
	invalidPath := cryptoutilTestutil.WriteTempFile(t, tempDir, "ci-invalid.yml", invalidContent)

	// Function logs validation issues but doesn't return error for malformed YAML
	err := Lint(logger, []string{invalidPath})
	// No error expected - function is lenient with malformed YAML (just logs warnings)
	require.NoError(t, err, "Function should not error on malformed YAML, only log warnings")
}

func TestLint_MissingFile(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("test_missing")

	nonexistentPath := "/nonexistent/path/to/workflow.yml"

	// Function filters workflow files and skips non-existent ones
	err := Lint(logger, []string{nonexistentPath})
	require.NoError(t, err, "Function should skip non-existent files, not error")
}

func TestLint_MultipleFiles(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("test_multiple")
	tempDir := t.TempDir()

	// Create multiple workflow files
	workflow1 := `name: Workflow 1
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`
	workflow2 := `name: Workflow 2
on: pull_request
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-go@v5
`
	path1 := cryptoutilTestutil.WriteTempFile(t, tempDir, "ci-workflow1.yml", workflow1)
	path2 := cryptoutilTestutil.WriteTempFile(t, tempDir, "ci-workflow2.yml", workflow2)

	err := Lint(logger, []string{path1, path2})

	// Should process multiple files without panicking
	if err != nil {
		require.Contains(t, err.Error(), "outdated", "Multi-file processing should work correctly")
	}
}

func TestLint_LoadExceptionsWarning(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("test_exceptions_warning")
	tempDir := t.TempDir()

	// Create invalid exception file (malformed JSON)
	exceptionsPath := filepath.Join(tempDir, ".github", "workflows-outdated-action-exemptions.json")
	err := os.MkdirAll(filepath.Dir(exceptionsPath), 0o755)
	require.NoError(t, err, "Failed to create .github directory")
	err = os.WriteFile(exceptionsPath, []byte(`{invalid json`), 0o600)
	require.NoError(t, err, "Failed to write invalid exception file")

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: failed to restore directory: %v", err)
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Create simple workflow file
	workflowContent := `name: Test Workflow
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Test step
        run: echo "test"
`
	workflowPath := cryptoutilTestutil.WriteTempFile(t, tempDir, "ci-test.yml", workflowContent)

	// Function should log warning about invalid exceptions file but continue
	err = Lint(logger, []string{workflowPath})
	require.NoError(t, err, "Should handle invalid exceptions file gracefully")
}
