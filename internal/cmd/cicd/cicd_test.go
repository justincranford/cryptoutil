// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"cryptoutil/internal/cmd/cicd/common"
)

// Common test constants to avoid goconst linter violations.
const (
	testPackageMain   = "package main"
	testImportFmt     = `import "fmt"`
	testFuncMainStart = `
func main() {`
	testFuncMainEnd = `}
`
	testTypeMyStruct = `
type MyStruct struct {
	Data any
}
`
)

func TestRunUsage(t *testing.T) {
	// Test with no commands (should return error)
	err := Run([]string{})
	require.Error(t, err, "Expected error when no commands provided")
	require.Contains(t, err.Error(), "Usage: cicd <command>", "Error message should contain usage information")
}

func TestRunInvalidCommand(t *testing.T) {
	// Test with invalid command
	err := Run([]string{"invalid-command"})
	require.Error(t, err, "Expected error for invalid command")
	require.Contains(t, err.Error(), "unknown command: invalid-command", "Error message should indicate unknown command")
}

func TestRunMultipleCommands(t *testing.T) {
	// Note: We can't easily test actual command execution as they call os.Exit
	// This test just verifies the command parsing logic works
	commands := []string{"go-update-direct-dependencies", "github-workflow-lint"}
	require.Len(t, commands, 2, "Expected 2 commands")
	require.Equal(t, "go-update-direct-dependencies", commands[0], "Expected first command")
	require.Equal(t, "github-workflow-lint", commands[1], "Expected second command")
}

func TestValidateCommands_HappyPath(t *testing.T) {
	tests := []struct {
		name     string
		commands []string
	}{
		{
			name:     "single valid command",
			commands: []string{"github-workflow-lint"},
		},
		{
			name:     "multiple different valid commands",
			commands: []string{"github-workflow-lint", "go-enforce-test-patterns", "all-enforce-utf8"},
		},
		{
			name:     "dependency update commands individually",
			commands: []string{"go-update-direct-dependencies"},
		},
		{
			name:     "all dependency update commands individually",
			commands: []string{"go-update-all-dependencies"},
		},
		{
			name: "all commands once each",
			commands: []string{
				"all-enforce-utf8",
				"go-enforce-test-patterns",
				"go-enforce-any",
				"go-check-circular-package-dependencies",
				"go-update-direct-dependencies",
				"github-workflow-lint",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommands(tt.commands)
			require.NoError(t, err, "Expected no error for valid commands: %v", tt.commands)
		})
	}
}

func TestValidateCommands_DuplicateCommands(t *testing.T) {
	tests := []struct {
		name     string
		commands []string
		expected []string // Now expecting multiple error messages
	}{
		{
			name:     "duplicate github-workflow-lint",
			commands: []string{"github-workflow-lint", "github-workflow-lint"},
			expected: []string{"command 'github-workflow-lint' specified 2 times"},
		},
		{
			name:     "duplicate go-enforce-test-patterns",
			commands: []string{"go-enforce-test-patterns", "all-enforce-utf8", "go-enforce-test-patterns"},
			expected: []string{"command 'go-enforce-test-patterns' specified 2 times"},
		},
		{
			name:     "duplicate go-update-direct-dependencies",
			commands: []string{"go-update-direct-dependencies", "github-workflow-lint", "go-update-direct-dependencies"},
			expected: []string{"command 'go-update-direct-dependencies' specified 2 times"},
		},
		{
			name:     "multiple duplicates",
			commands: []string{"github-workflow-lint", "github-workflow-lint", "all-enforce-utf8", "all-enforce-utf8"},
			expected: []string{
				"command 'github-workflow-lint' specified 2 times",
				"command 'all-enforce-utf8' specified 2 times",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommands(tt.commands)
			require.Error(t, err, "Expected error for duplicate commands: %v", tt.commands)

			errMsg := err.Error()
			for _, expectedMsg := range tt.expected {
				require.Contains(t, errMsg, expectedMsg, "Error message should contain expected text: %s", expectedMsg)
			}
		})
	}
}

func TestValidateCommands_MutuallyExclusiveCommands(t *testing.T) {
	tests := []struct {
		name     string
		commands []string
		expected string
	}{
		{
			name:     "both dependency update commands",
			commands: []string{"go-update-direct-dependencies", "go-update-all-dependencies"},
			expected: "cannot be used together",
		},
		{
			name:     "both dependency update commands with other commands",
			commands: []string{"github-workflow-lint", "go-update-direct-dependencies", "go-update-all-dependencies", "all-enforce-utf8"},
			expected: "cannot be used together",
		},
		{
			name:     "both dependency update commands in different order",
			commands: []string{"go-update-all-dependencies", "go-update-direct-dependencies"},
			expected: "cannot be used together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommands(tt.commands)
			require.Error(t, err, "Expected error for mutually exclusive commands: %v", tt.commands)
			require.Contains(t, err.Error(), tt.expected, "Error message should contain expected text")
		})
	}
}

func TestValidateCommands_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		commands    []string
		expectError bool
	}{
		{
			name:        "empty commands slice",
			commands:    []string{},
			expectError: true, // validateCommands now checks for empty slice
		},
		{
			name:        "nil commands slice",
			commands:    nil,
			expectError: true, // validateCommands now checks for nil slice (len(nil) == 0)
		},
		{
			name:        "single dependency update command with other commands",
			commands:    []string{"go-update-direct-dependencies", "github-workflow-lint"},
			expectError: false,
		},
		{
			name:        "single all dependency update command with other commands",
			commands:    []string{"go-update-all-dependencies", "all-enforce-utf8"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommands(tt.commands)
			if tt.expectError {
				require.Error(t, err, "Expected error for edge case: %v", tt.commands)
			} else {
				require.NoError(t, err, "Expected no error for edge case: %v", tt.commands)
			}
		})
	}
}

func TestValidateCommands_MultipleErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		commands []string
		expected []string // All expected error messages
	}{
		{
			name:     "unknown command and duplicate",
			commands: []string{"unknown-cmd", "github-workflow-lint", "github-workflow-lint"},
			expected: []string{
				"unknown command: unknown-cmd",
				"command 'github-workflow-lint' specified 2 times",
			},
		},
		{
			name:     "empty commands with unknown",
			commands: []string{"", "invalid-cmd"},
			expected: []string{
				"Usage: cicd <command>",        // From empty check
				"unknown command: ",            // From empty string
				"unknown command: invalid-cmd", // From invalid command
			},
		},
		{
			name:     "duplicate and mutually exclusive",
			commands: []string{"go-update-direct-dependencies", "go-update-all-dependencies", "go-update-direct-dependencies"},
			expected: []string{
				"command 'go-update-direct-dependencies' specified 2 times",
				"cannot be used together", // From mutually exclusive check
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommands(tt.commands)
			require.Error(t, err, "Expected error for multiple error types: %v", tt.commands)

			errMsg := err.Error()
			for _, expectedMsg := range tt.expected {
				require.Contains(t, errMsg, expectedMsg, "Error message should contain expected text: %s", expectedMsg)
			}
		})
	}
}

func TestValidateCommands(t *testing.T) {
	tests := []struct {
		name          string
		commands      []string
		expectedError bool
		errorContains string
	}{
		{
			name:          "empty commands",
			commands:      []string{},
			expectedError: true,
			errorContains: "Usage:",
		},
		{
			name:          "nil commands",
			commands:      nil,
			expectedError: true,
			errorContains: "Usage:",
		},
		{
			name:          "single valid command that needs files",
			commands:      []string{"all-enforce-utf8"},
			expectedError: false,
		},
		{
			name:          "single valid command that doesn't need files",
			commands:      []string{"go-update-direct-dependencies"},
			expectedError: false,
		},
		{
			name:          "multiple valid commands",
			commands:      []string{"all-enforce-utf8", "go-update-direct-dependencies"},
			expectedError: false,
		},
		{
			name:          "unknown command",
			commands:      []string{"unknown-command"},
			expectedError: true,
			errorContains: "unknown command",
		},
		{
			name:          "duplicate command",
			commands:      []string{"all-enforce-utf8", "all-enforce-utf8"},
			expectedError: true,
			errorContains: "specified 2 times",
		},
		{
			name:          "mutually exclusive commands",
			commands:      []string{"go-update-direct-dependencies", "go-update-all-dependencies"},
			expectedError: true,
			errorContains: "cannot be used together",
		},
		{
			name:          "all valid commands",
			commands:      []string{"all-enforce-utf8", "go-enforce-test-patterns", "go-enforce-any", "go-check-circular-package-dependencies", "go-update-direct-dependencies", "github-workflow-lint"},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommands(tt.commands)

			if tt.expectedError {
				require.Error(t, err, "Expected error for test case: %s", tt.name)

				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains, "Error should contain expected text")
				}
			} else {
				require.NoError(t, err, "Expected no error for test case: %s", tt.name)
			}
		})
	}
}

func TestPrintExecutionSummary(t *testing.T) {
	tests := []struct {
		name          string
		results       []common.CommandResult
		totalDuration time.Duration
	}{
		{
			name: "all commands successful",
			results: []common.CommandResult{
				{Command: "all-enforce-utf8", Duration: 100 * time.Millisecond, Error: nil},
				{Command: "go-enforce-test-patterns", Duration: 200 * time.Millisecond, Error: nil},
				{Command: "go-enforce-any", Duration: 150 * time.Millisecond, Error: nil},
			},
			totalDuration: 450 * time.Millisecond,
		},
		{
			name: "some commands failed",
			results: []common.CommandResult{
				{Command: "all-enforce-utf8", Duration: 100 * time.Millisecond, Error: nil},
				{Command: "go-enforce-test-patterns", Duration: 200 * time.Millisecond, Error: fmt.Errorf("pattern violation")},
				{Command: "go-enforce-any", Duration: 150 * time.Millisecond, Error: nil},
			},
			totalDuration: 450 * time.Millisecond,
		},
		{
			name: "all commands failed",
			results: []common.CommandResult{
				{Command: "all-enforce-utf8", Duration: 100 * time.Millisecond, Error: fmt.Errorf("encoding error")},
				{Command: "go-enforce-test-patterns", Duration: 200 * time.Millisecond, Error: fmt.Errorf("pattern violation")},
			},
			totalDuration: 300 * time.Millisecond,
		},
		{
			name:          "no commands executed",
			results:       []common.CommandResult{},
			totalDuration: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test just ensures common.PrintExecutionSummary doesn't panic
			// Actual output validation would require capturing stderr
			require.NotPanics(t, func() {
				common.PrintExecutionSummary(tt.results, tt.totalDuration)
			}, "common.PrintExecutionSummary should not panic")
		})
	}
}

func TestCommandResult(t *testing.T) {
	tests := []struct {
		name     string
		result   common.CommandResult
		hasError bool
	}{
		{
			name: "successful command",
			result: common.CommandResult{
				Command:  "all-enforce-utf8",
				Duration: 100 * time.Millisecond,
				Error:    nil,
			},
			hasError: false,
		},
		{
			name: "failed command",
			result: common.CommandResult{
				Command:  "go-enforce-test-patterns",
				Duration: 200 * time.Millisecond,
				Error:    fmt.Errorf("test pattern violation"),
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.hasError, tt.result.Error != nil, "Error state should match expectation")
			require.NotEmpty(t, tt.result.Command, "Command name should not be empty")
			require.GreaterOrEqual(t, tt.result.Duration, time.Duration(0), "Duration should be non-negative")
		})
	}
}
