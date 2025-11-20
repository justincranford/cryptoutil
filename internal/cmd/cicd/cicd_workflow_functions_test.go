// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"cryptoutil/internal/cmd/cicd/common"
)

// TestIsOutdated_EdgeCases tests version comparison logic.
func TestIsOutdated_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		latest   string
		expected bool
	}{
		{"same version", "v4.0.0", "v4.0.0", false},
		{"different patch", "v4.0.0", "v4.0.1", true},
		{"different minor", "v4.0.0", "v4.1.0", true},
		{"major version pin same", "v4", "v4.0.0", false},
		{"major version pin different", "v4", "v5.0.0", true},
		{"main branch", "main", "v5.0.0", false},
		{"master branch", "master", "v5.0.0", false},
		{"variable reference", "${{ env.VERSION }}", "v5.0.0", false},
		{"specific vs newer", "v3.2.1", "v3.2.2", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOutdated(tt.current, tt.latest)
			require.Equal(t, tt.expected, result, "isOutdated(%q, %q) = %v, want %v", tt.current, tt.latest, result, tt.expected)
		})
	}
}

// TestLoadWorkflowActionExceptions tests loading exceptions file.
func TestLoadWorkflowActionExceptions(t *testing.T) {
	// Test with no file - should return empty exceptions
	exceptions, err := loadWorkflowActionExceptions()
	require.NoError(t, err)
	require.NotNil(t, exceptions)
	require.NotNil(t, exceptions.Exceptions)
	require.Empty(t, exceptions.Exceptions)
}

// TestLoadWorkflowActionExceptions_WithFile tests loading valid exceptions file.
func TestLoadWorkflowActionExceptions_WithFile(t *testing.T) {
	// Create temporary exceptions file
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	//nolint:errcheck // Best effort to restore directory
	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create .github directory
	githubDir := filepath.Join(tmpDir, ".github")
	err = os.MkdirAll(githubDir, 0o755)
	require.NoError(t, err)

	// Create exceptions file with correct filename: workflows-outdated-action-exemptions.json
	exceptionsFile := filepath.Join(githubDir, "workflows-outdated-action-exemptions.json")
	exceptionsContent := `{
		"exceptions": {
			"actions/checkout": {
				"allowed_versions": ["v3", "v4"]
			}
		}
	}`
	err = os.WriteFile(exceptionsFile, []byte(exceptionsContent), 0o600)
	require.NoError(t, err)

	exceptions, err := loadWorkflowActionExceptions()
	require.NoError(t, err)
	require.NotNil(t, exceptions)
	require.Contains(t, exceptions.Exceptions, "actions/checkout")
	require.Contains(t, exceptions.Exceptions["actions/checkout"].AllowedVersions, "v3")
	require.Contains(t, exceptions.Exceptions["actions/checkout"].AllowedVersions, "v4")
}

// TestLoadWorkflowActionExceptions_InvalidJSON tests handling of corrupt file.
func TestLoadWorkflowActionExceptions_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	//nolint:errcheck // Best effort to restore directory
	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create .github directory
	githubDir := filepath.Join(tmpDir, ".github")
	err = os.MkdirAll(githubDir, 0o755)
	require.NoError(t, err)

	// Create invalid JSON file with correct filename: workflows-outdated-action-exemptions.json
	exceptionsFile := filepath.Join(githubDir, "workflows-outdated-action-exemptions.json")
	err = os.WriteFile(exceptionsFile, []byte("invalid json"), 0o600)
	require.NoError(t, err)

	_, err = loadWorkflowActionExceptions()
	require.Error(t, err)
	require.Contains(t, err.Error(), "unmarshal")
}

// TestParseWorkflowFile_WithActions tests workflow file parsing.
func TestParseWorkflowFile_WithActions(t *testing.T) {
	tmpDir := t.TempDir()

	workflowFile := filepath.Join(tmpDir, "test.yml")
	workflowContent := `name: Test
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - name: Test
        run: go test ./...
`
	err := os.WriteFile(workflowFile, []byte(workflowContent), 0o600)
	require.NoError(t, err)

	actions, err := parseWorkflowFile(workflowFile)
	require.NoError(t, err)
	require.Len(t, actions, 2)

	// Check first action
	require.Equal(t, "actions/checkout", actions[0].Name)
	require.Equal(t, "v4", actions[0].CurrentVersion)
	require.Contains(t, actions[0].WorkflowFiles, "test.yml")

	// Check second action
	require.Equal(t, "actions/setup-go", actions[1].Name)
	require.Equal(t, "v5", actions[1].CurrentVersion)
	require.Contains(t, actions[1].WorkflowFiles, "test.yml")
}

// TestParseWorkflowFile_NoActions tests workflow without actions.
func TestParseWorkflowFile_NoActions(t *testing.T) {
	tmpDir := t.TempDir()

	workflowFile := filepath.Join(tmpDir, "test.yml")
	workflowContent := `name: Test
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Test
        run: echo "test"
`
	err := os.WriteFile(workflowFile, []byte(workflowContent), 0o600)
	require.NoError(t, err)

	actions, err := parseWorkflowFile(workflowFile)
	require.NoError(t, err)
	require.Empty(t, actions)
}

// TestParseWorkflowFile_FileNotFound tests missing file handling.
func TestParseWorkflowFile_FileNotFound(t *testing.T) {
	_, err := parseWorkflowFile("/nonexistent/file.yml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read")
}

// TestCheckActionVersionsConcurrently_Coverage tests concurrent checking.
func TestCheckActionVersionsConcurrently_Coverage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	logger := common.NewLogger("TestCheckActionVersionsConcurrently_Coverage")

	// Test with empty map
	outdated, exempted, errors := checkActionVersionsConcurrently(logger, map[string]WorkflowActionDetails{}, &WorkflowActionExceptions{Exceptions: make(map[string]WorkflowActionException)})
	require.Empty(t, outdated)
	require.Empty(t, exempted)
	require.Empty(t, errors)

	// Test with exempted action
	actionMap := map[string]WorkflowActionDetails{
		"actions/checkout@v3": {
			Name:           "actions/checkout",
			CurrentVersion: "v3",
			WorkflowFiles:  []string{"test.yml"},
		},
	}

	exceptions := &WorkflowActionExceptions{
		Exceptions: map[string]WorkflowActionException{
			"actions/checkout": {
				AllowedVersions: []string{"v3"},
			},
		},
	}

	outdated, exempted, errors = checkActionVersionsConcurrently(logger, actionMap, exceptions)
	require.Empty(t, outdated)
	require.Len(t, exempted, 1)
	require.Empty(t, errors)
	require.Equal(t, "actions/checkout", exempted[0].Name)
}

// TestCheckActionVersionsConcurrently_NonExempted tests non-exempted actions.
func TestCheckActionVersionsConcurrently_NonExempted(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	logger := common.NewLogger("TestCheckActionVersionsConcurrently_NonExempted")

	actionMap := map[string]WorkflowActionDetails{
		"actions/checkout@v4": {
			Name:           "actions/checkout",
			CurrentVersion: "v4",
			WorkflowFiles:  []string{"test.yml"},
		},
	}

	exceptions := &WorkflowActionExceptions{
		Exceptions: make(map[string]WorkflowActionException),
	}

	outdated, exempted, errors := checkActionVersionsConcurrently(logger, actionMap, exceptions)

	// Should have checked the action (may or may not be outdated)
	require.Empty(t, exempted)
	// Either outdated or no errors
	require.True(t, len(outdated) > 0 || len(errors) == 0)
}
