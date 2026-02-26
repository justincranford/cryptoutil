// Copyright (c) 2025 Justin Cranford
//
//

//nolint:errcheck // Test infrastructure uses os.Pipe/Close/ReadFrom without error checking
package common

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPrintExecutionSummary(t *testing.T) {
	results := []CommandResult{
		{Command: "test-command-1", Duration: cryptoutilSharedMagic.JoseJAMaxMaterials * time.Millisecond, Error: nil},
		{Command: "test-command-2", Duration: 200 * time.Millisecond, Error: errors.New("test error")},
		{Command: "test-command-3", Duration: 150 * time.Millisecond, Error: nil},
	}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	PrintExecutionSummary(results, cryptoutilSharedMagic.TestDefaultRateLimitServiceIP * time.Millisecond)

	if err := w.Close(); err != nil {
		t.Logf("Warning: failed to close write pipe: %v", err)
	}

	os.Stderr = oldStderr

	var buf bytes.Buffer

	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.Contains(t, output, "EXECUTION SUMMARY", "Output should contain summary header")
	require.Contains(t, output, "test-command-1", "Output should contain command 1")
	require.Contains(t, output, "test-command-2", "Output should contain command 2")
	require.Contains(t, output, "test-command-3", "Output should contain command 3")
	require.Contains(t, output, cryptoutilSharedMagic.StatusSuccess, "Output should contain success marker")
	require.Contains(t, output, cryptoutilSharedMagic.StatusFailed, "Output should contain failure marker")
	require.Contains(t, output, "Total: 3 commands", "Output should show total count")
	require.Contains(t, output, "Passed: 2", "Output should show passed count")
	require.Contains(t, output, "Failed: 1", "Output should show failed count")
	require.Contains(t, output, "0.50s", "Output should show total duration")
}

func TestPrintExecutionSummary_AllSuccess(t *testing.T) {
	results := []CommandResult{
		{Command: "cmd1", Duration: cryptoutilSharedMagic.IMMaxUsernameLength * time.Millisecond, Error: nil},
		{Command: "cmd2", Duration: 75 * time.Millisecond, Error: nil},
	}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	PrintExecutionSummary(results, 150*time.Millisecond)

	if err := w.Close(); err != nil {
		t.Logf("Warning: failed to close write pipe: %v", err)
	}

	os.Stderr = oldStderr

	var buf bytes.Buffer

	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.Contains(t, output, "Passed: 2", "All commands should pass")
	require.Contains(t, output, "Failed: 0", "No failed commands")
}

func TestPrintExecutionSummary_AllFailure(t *testing.T) {
	results := []CommandResult{
		{Command: "cmd1", Duration: cryptoutilSharedMagic.IMMaxUsernameLength * time.Millisecond, Error: errors.New("error1")},
		{Command: "cmd2", Duration: 75 * time.Millisecond, Error: errors.New("error2")},
	}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	PrintExecutionSummary(results, 150*time.Millisecond)

	if err := w.Close(); err != nil {
		t.Logf("Warning: failed to close write pipe: %v", err)
	}

	os.Stderr = oldStderr

	var buf bytes.Buffer

	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.Contains(t, output, "Passed: 0", "No passed commands")
	require.Contains(t, output, "Failed: 2", "All commands should fail")
}

func TestPrintExecutionSummary_Empty(t *testing.T) {
	results := []CommandResult{}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	PrintExecutionSummary(results, 0)

	if err := w.Close(); err != nil {
		t.Logf("Warning: failed to close write pipe: %v", err)
	}

	os.Stderr = oldStderr

	var buf bytes.Buffer

	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.Contains(t, output, "Total: 0 commands", "Should handle empty results")
	require.Contains(t, output, "Passed: 0", "No passed commands")
	require.Contains(t, output, "Failed: 0", "No failed commands")
}

func TestPrintCommandSeparator(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	PrintCommandSeparator()

	if err := w.Close(); err != nil {
		t.Logf("Warning: failed to close write pipe: %v", err)
	}

	os.Stderr = oldStderr

	var buf bytes.Buffer

	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.Contains(t, output, "=", "Output should contain separator characters")
	lines := strings.Split(output, "\n")
	require.GreaterOrEqual(t, len(lines), 2, "Should output multiple lines")
}

func TestCalculateStats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		results          []CommandResult
		totalDuration    time.Duration
		expectedTotal    int
		expectedPassed   int
		expectedFailed   int
		expectedDuration time.Duration
	}{
		{
			name: "mixed results",
			results: []CommandResult{
				{Command: "cmd1", Duration: cryptoutilSharedMagic.JoseJAMaxMaterials * time.Millisecond, Error: nil},
				{Command: "cmd2", Duration: 200 * time.Millisecond, Error: errors.New(cryptoutilSharedMagic.StringError)},
				{Command: "cmd3", Duration: 150 * time.Millisecond, Error: nil},
			},
			totalDuration:    cryptoutilSharedMagic.TestDefaultRateLimitServiceIP * time.Millisecond,
			expectedTotal:    3,
			expectedPassed:   2,
			expectedFailed:   1,
			expectedDuration: cryptoutilSharedMagic.TestDefaultRateLimitServiceIP * time.Millisecond,
		},
		{
			name: "all success",
			results: []CommandResult{
				{Command: "cmd1", Duration: cryptoutilSharedMagic.IMMaxUsernameLength * time.Millisecond, Error: nil},
				{Command: "cmd2", Duration: 75 * time.Millisecond, Error: nil},
			},
			totalDuration:    150 * time.Millisecond,
			expectedTotal:    2,
			expectedPassed:   2,
			expectedFailed:   0,
			expectedDuration: 150 * time.Millisecond,
		},
		{
			name: "all failures",
			results: []CommandResult{
				{Command: "cmd1", Duration: cryptoutilSharedMagic.IMMaxUsernameLength * time.Millisecond, Error: errors.New("err1")},
				{Command: "cmd2", Duration: 75 * time.Millisecond, Error: errors.New("err2")},
			},
			totalDuration:    150 * time.Millisecond,
			expectedTotal:    2,
			expectedPassed:   0,
			expectedFailed:   2,
			expectedDuration: 150 * time.Millisecond,
		},
		{
			name:             "empty results",
			results:          []CommandResult{},
			totalDuration:    0,
			expectedTotal:    0,
			expectedPassed:   0,
			expectedFailed:   0,
			expectedDuration: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stats := CalculateStats(tc.results, tc.totalDuration)

			require.Equal(t, tc.expectedTotal, stats.Total, "Total should match")
			require.Equal(t, tc.expectedPassed, stats.Passed, "Passed should match")
			require.Equal(t, tc.expectedFailed, stats.Failed, "Failed should match")
			require.Equal(t, tc.expectedDuration, stats.Duration, "Duration should match")
		})
	}
}

func TestHasFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		results  []CommandResult
		expected bool
	}{
		{
			name: "no failures",
			results: []CommandResult{
				{Command: "cmd1", Error: nil},
				{Command: "cmd2", Error: nil},
			},
			expected: false,
		},
		{
			name: "has failures",
			results: []CommandResult{
				{Command: "cmd1", Error: nil},
				{Command: "cmd2", Error: errors.New(cryptoutilSharedMagic.StringError)},
			},
			expected: true,
		},
		{
			name: "all failures",
			results: []CommandResult{
				{Command: "cmd1", Error: errors.New("err1")},
				{Command: "cmd2", Error: errors.New("err2")},
			},
			expected: true,
		},
		{
			name:     "empty results",
			results:  []CommandResult{},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := HasFailures(tc.results)

			require.Equal(t, tc.expected, result, "HasFailures should match expected")
		})
	}
}

func TestGetFailedCommands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		results  []CommandResult
		expected []string
	}{
		{
			name: "mixed results",
			results: []CommandResult{
				{Command: "cmd1", Error: nil},
				{Command: "cmd2", Error: errors.New(cryptoutilSharedMagic.StringError)},
				{Command: "cmd3", Error: nil},
				{Command: "cmd4", Error: errors.New("error2")},
			},
			expected: []string{"cmd2", "cmd4"},
		},
		{
			name: "no failures",
			results: []CommandResult{
				{Command: "cmd1", Error: nil},
				{Command: "cmd2", Error: nil},
			},
			expected: nil,
		},
		{
			name: "all failures",
			results: []CommandResult{
				{Command: "cmd1", Error: errors.New("err1")},
				{Command: "cmd2", Error: errors.New("err2")},
			},
			expected: []string{"cmd1", "cmd2"},
		},
		{
			name:     "empty results",
			results:  []CommandResult{},
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			failed := GetFailedCommands(tc.results)

			require.Equal(t, tc.expected, failed, "Failed commands should match expected")
		})
	}
}
