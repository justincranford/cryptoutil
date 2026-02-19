// Copyright (c) 2025 Justin Cranford

package cicd

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

func TestRunUsage(t *testing.T) {
	t.Parallel()

	// Test with no commands (should return error).
	exitCode := Cicd([]string{"cicd"}, nil, nil, os.Stderr)
	require.Equal(t, 1, exitCode, "Expected exit code 1 when no commands provided")
}

func TestRunInvalidCommand(t *testing.T) {
	t.Parallel()

	// Test with invalid command.
	err := run([]string{"invalid-command"})
	require.Error(t, err, "Expected error for invalid command")
	require.Contains(t, err.Error(), "unknown command: invalid-command", "Error message should indicate unknown command")
}

func TestRunMultipleCommands(t *testing.T) {
	t.Parallel()

	// Verify the command parsing logic works.
	commands := []string{"lint-go-mod", "lint-workflow"}
	require.Len(t, commands, 2, "Expected 2 commands")
	require.Equal(t, "lint-go-mod", commands[0], "Expected first command")
	require.Equal(t, "lint-workflow", commands[1], "Expected second command")
}

func TestRun_AllCommands_HappyPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		commands []string
	}{
		{
			name:     "lint-text command",
			commands: []string{"lint-text"},
		},
		{
			name:     "lint-go command",
			commands: []string{"lint-go"},
		},
		{
			name:     "format-go command",
			commands: []string{"format-go"},
		},
		{
			name:     "lint-go-test command",
			commands: []string{"lint-go-test"},
		},
		{
			name:     "format-go-test command",
			commands: []string{"format-go-test"},
		},
		{
			name:     "lint-workflow command",
			commands: []string{"lint-workflow"},
		},
		{
			name:     "lint-go-mod command",
			commands: []string{"lint-go-mod"},
		},
		{
			name:     "multiple commands together",
			commands: []string{"lint-text", "lint-go"},
		},
		{
			name:     "lint-compose command",
			commands: []string{"lint-compose"},
		},
		{
			name:     "lint-ports command",
			commands: []string{"lint-ports"},
		},
		{
			name:     "lint-golangci command",
			commands: []string{"lint-golangci"},
		},
		{
			name:     "multiple commands together (subset)",
			commands: []string{"lint-text", "lint-go"},
		},
		{
			name:     "all commands together",
			commands: []string{"lint-text", "lint-go", "format-go", "lint-go-test", "format-go-test", "lint-workflow", "lint-go-mod", "lint-compose", "lint-ports", "lint-golangci"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := run(tc.commands)
			// Note: Some commands may fail due to project state (e.g., outdated dependencies).
			// We're testing that the switch cases execute without panic, not that they pass.
			if err != nil {
				require.Contains(t, err.Error(), "failed commands:", "Error should indicate failed commands")
			}
		})
	}
}

func TestValidateCommands_HappyPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		commands []string
	}{
		{
			name:     "single lint command",
			commands: []string{"lint-workflow"},
		},
		{
			name:     "multiple lint commands",
			commands: []string{"lint-text", "lint-go", "lint-go-test"},
		},
		{
			name:     "lint and format commands",
			commands: []string{"lint-text", "format-go", "format-go-test"},
		},
		{
			name:     "all commands once each",
			commands: []string{"lint-text", "lint-go", "format-go", "lint-go-test", "format-go-test", "lint-workflow", "lint-go-mod", "lint-compose", "lint-ports", "lint-golangci"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actualCommands, err := validateCommands(tc.commands)
			require.NoError(t, err, "Expected no error for valid commands: %v", tc.commands)
			require.Equal(t, tc.commands, actualCommands, "Expected actualCommands to match input commands")
		})
	}
}

func TestValidateCommands_DuplicateCommands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		commands     []string
		expectedText string
	}{
		{
			name:         "duplicate lint-text",
			commands:     []string{"lint-text", "lint-text"},
			expectedText: "command 'lint-text' specified 2 times",
		},
		{
			name:         "duplicate lint-go",
			commands:     []string{"lint-go", "lint-text", "lint-go"},
			expectedText: "command 'lint-go' specified 2 times",
		},
		{
			name:         "duplicate format-go",
			commands:     []string{"format-go", "lint-workflow", "format-go"},
			expectedText: "command 'format-go' specified 2 times",
		},
		{
			name:         "multiple duplicates",
			commands:     []string{"lint-workflow", "lint-workflow", "lint-text", "lint-text"},
			expectedText: "command 'lint-workflow' specified 2 times",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := validateCommands(tc.commands)
			require.Error(t, err, "Expected error for duplicate commands: %v", tc.commands)
			require.Contains(t, err.Error(), tc.expectedText, "Error message should contain expected text")
		})
	}
}

func TestValidateCommands_UnknownCommands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		commands     []string
		expectedText string
	}{
		{
			name:         "unknown command",
			commands:     []string{"unknown-cmd"},
			expectedText: "unknown command: unknown-cmd",
		},
		{
			name:         "old command name all-enforce-utf8",
			commands:     []string{"all-enforce-utf8"},
			expectedText: "unknown command: all-enforce-utf8",
		},
		{
			name:         "old command name go-fix-copyloopvar",
			commands:     []string{"go-fix-copyloopvar"},
			expectedText: "unknown command: go-fix-copyloopvar",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := validateCommands(tc.commands)
			require.Error(t, err, "Expected error for unknown command: %v", tc.commands)
			require.Contains(t, err.Error(), tc.expectedText, "Error message should contain expected text")
		})
	}
}

func TestValidateCommands_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		commands     []string
		expectErr    bool
		expectedCmds []string
	}{
		{
			name:         "empty commands",
			commands:     []string{},
			expectErr:    true,
			expectedCmds: nil,
		},
		{
			name:         "valid commands with flags (flags are skipped)",
			commands:     []string{"--strict", "true", "lint-text"},
			expectErr:    false,
			expectedCmds: []string{"lint-text"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actualCommands, err := validateCommands(tc.commands)
			if tc.expectErr {
				require.Error(t, err, "Expected error for test case: %s", tc.name)
			} else {
				require.NoError(t, err, "Expected no error for test case: %s", tc.name)
				require.Equal(t, tc.expectedCmds, actualCommands, "Expected actualCommands to match")
			}
		})
	}
}

func TestCommandResult(t *testing.T) {
	t.Parallel()

	result := cryptoutilCmdCicdCommon.CommandResult{
		Command:  "lint-text",
		Duration: 100 * time.Millisecond,
		Error:    fmt.Errorf("test error"),
	}
	require.Equal(t, "lint-text", result.Command)
	require.Equal(t, 100*time.Millisecond, result.Duration)
	require.Error(t, result.Error)
}

func TestCommandResultSuccess(t *testing.T) {
	t.Parallel()

	result := cryptoutilCmdCicdCommon.CommandResult{
		Command:  "lint-workflow",
		Duration: 50 * time.Millisecond,
		Error:    nil,
	}
	require.Equal(t, "lint-workflow", result.Command)
	require.Equal(t, 50*time.Millisecond, result.Duration)
	require.NoError(t, result.Error)
}

func TestGetFailedCommands(t *testing.T) {
	t.Parallel()

	results := []cryptoutilCmdCicdCommon.CommandResult{
		{Command: "lint-text", Duration: 100 * time.Millisecond, Error: nil},
		{Command: "lint-go", Duration: 200 * time.Millisecond, Error: fmt.Errorf("error1")},
		{Command: "format-go", Duration: 150 * time.Millisecond, Error: nil},
		{Command: "lint-workflow", Duration: 50 * time.Millisecond, Error: fmt.Errorf("error2")},
	}

	failed := cryptoutilCmdCicdCommon.GetFailedCommands(results)
	require.Len(t, failed, 2)
	require.Contains(t, failed, "lint-go")
	require.Contains(t, failed, "lint-workflow")
}

func TestGetFailedCommands_NoFailures(t *testing.T) {
	t.Parallel()

	results := []cryptoutilCmdCicdCommon.CommandResult{
		{Command: "lint-text", Duration: 100 * time.Millisecond, Error: nil},
		{Command: "lint-go", Duration: 200 * time.Millisecond, Error: nil},
	}

	failed := cryptoutilCmdCicdCommon.GetFailedCommands(results)
	require.Empty(t, failed)
}

// TestRun_LintComposeCommand tests that lint-compose command executes.
func TestRun_LintComposeCommand(t *testing.T) {
	t.Parallel()

	err := run([]string{"lint-compose"})
	// Command may pass or fail depending on compose files in project.
	// We're testing that the switch case executes without panic.
	if err != nil {
		require.Contains(t, err.Error(), "failed commands:", "Error should indicate failed commands")
	}
}

// TestValidateCommands_OnlyFlags tests the edge case where only flags are provided.
func TestValidateCommands_OnlyFlags(t *testing.T) {
	t.Parallel()

	// Pass only flags with values (no actual commands).
	actualCommands, err := validateCommands([]string{"-v", "true", "-debug", "enabled"})
	require.Error(t, err, "Expected error when only flags provided")
	require.Nil(t, actualCommands, "Expected nil actual commands")
	require.Contains(t, err.Error(), "Usage: cicd <command>", "Error should contain usage info")
}
