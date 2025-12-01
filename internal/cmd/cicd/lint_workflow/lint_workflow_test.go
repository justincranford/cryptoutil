// Copyright (c) 2025 Justin Cranford

package lint_workflow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
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
	content := `name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
`
	err = os.WriteFile(workflowFile, []byte(content), 0o600)
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
