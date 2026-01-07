// Copyright (c) 2025 Justin Cranford
//
//

package workflow

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestStatusBadge tests the status badge generation.
func TestStatusBadge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		success  bool
		expected string
	}{
		{
			name:     "Success badge",
			success:  true,
			expected: "✅ SUCCESS",
		},
		{
			name:     "Failure badge",
			success:  false,
			expected: "❌ FAILED",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := statusBadge(tc.success)

			require.Equal(t, tc.expected, result, "Status badge should match expected value")
		})
	}
}

// TestContains tests the slice contains function.
func TestContains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "Item exists in slice",
			slice:    []string{"quality", "coverage", "dast"},
			item:     "coverage",
			expected: true,
		},
		{
			name:     "Item does not exist in slice",
			slice:    []string{"quality", "coverage", "dast"},
			item:     "unknown",
			expected: false,
		},
		{
			name:     "Empty slice",
			slice:    []string{},
			item:     "quality",
			expected: false,
		},
		{
			name:     "Nil slice",
			slice:    nil,
			item:     "quality",
			expected: false,
		},
		{
			name:     "Item at start of slice",
			slice:    []string{"first", "second", "third"},
			item:     "first",
			expected: true,
		},
		{
			name:     "Item at end of slice",
			slice:    []string{"first", "second", "third"},
			item:     "third",
			expected: true,
		},
		{
			name:     "Case sensitive check - exact match",
			slice:    []string{"Quality", "Coverage"},
			item:     "Quality",
			expected: true,
		},
		{
			name:     "Case sensitive check - no match",
			slice:    []string{"Quality", "Coverage"},
			item:     "quality",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := contains(tc.slice, tc.item)

			require.Equal(t, tc.expected, result, "Contains result should match expected value")
		})
	}
}

// TestGetWorkflowFile tests the workflow file path generation.
func TestGetWorkflowFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		workflowName string
		expected     string
	}{
		{
			name:         "Quality workflow",
			workflowName: "quality",
			expected:     ".github/workflows/ci-quality.yml",
		},
		{
			name:         "Coverage workflow",
			workflowName: "coverage",
			expected:     ".github/workflows/ci-coverage.yml",
		},
		{
			name:         "DAST workflow",
			workflowName: "dast",
			expected:     ".github/workflows/ci-dast.yml",
		},
		{
			name:         "E2E workflow",
			workflowName: "e2e",
			expected:     ".github/workflows/ci-e2e.yml",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := getWorkflowFile(tc.workflowName)

			require.Equal(t, tc.expected, result, "Workflow file path should match expected value")
		})
	}
}

// TestGetWorkflowLogFile tests the workflow log file path generation with timestamp.
func TestGetWorkflowLogFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		outputDir    string
		workflowName string
	}{
		{
			name:         "Quality workflow in default output dir",
			outputDir:    "test-output",
			workflowName: "quality",
		},
		{
			name:         "Coverage workflow in custom dir",
			outputDir:    "/tmp/output",
			workflowName: "coverage",
		},
		{
			name:         "DAST workflow in relative dir",
			outputDir:    "../output",
			workflowName: "dast",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := getWorkflowLogFile(tc.outputDir, tc.workflowName)

			// Verify format: outputDir/workflowName-YYYY-MM-DD_HH-MM-SS.log
			// Normalize paths for cross-platform comparison
			resultDir := filepath.Dir(result)
			expectedDir := filepath.Clean(tc.outputDir)

			require.Equal(t, expectedDir, resultDir, "Log file should be in output directory")
			require.True(t, strings.Contains(result, tc.workflowName), "Log file should contain workflow name")
			require.True(t, strings.HasSuffix(result, ".log"), "Log file should have .log extension")

			currentYear := time.Now().UTC().Format("2006")
			require.True(t, strings.Contains(result, "-"+currentYear+"-"), "Log file should contain timestamp with year")
		})
	}
}
