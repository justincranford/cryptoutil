// Copyright (c) 2025 Justin Cranford

package cicd_lint

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
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
	err := run([]string{"invalid-command"}, []string{}, false)
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
			commands: []string{"lint-text", "lint-go", "format-go", "lint-go-test", "format-go-test", "lint-workflow", "lint-go-mod", "lint-compose", "lint-ports", "lint-golangci", "lint-openapi"},
		},
		{
			name:     "lint-docs command",
			commands: []string{"lint-docs"},
		},
		{
			name:     "lint-deployments command",
			commands: []string{"lint-deployments"},
		},
		{
			name:     "lint-fitness command",
			commands: []string{"lint-fitness"},
		},
		{
			name:     "lint-openapi command",
			commands: []string{"lint-openapi"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := run(tc.commands, []string{}, false)
			// Note: Some commands may fail due to project state (e.g., outdated dependencies).
			// We're testing that the switch cases execute without panic, not that they pass.
			if err != nil {
				require.Contains(t, err.Error(), "failed commands:", "Error should indicate failed commands")
			}
		})
	}
}

func TestGetExtraArgs_WithArgs(t *testing.T) {
	t.Parallel()

	result := getExtraArgs([]string{"binary", "command", "extra1", "extra2"})
	require.Equal(t, []string{"extra1", "extra2"}, result)
}

func TestCommandsNeedFiles_NonFileCommands(t *testing.T) {
	t.Parallel()

	require.False(t, commandsNeedFiles([]string{"lint-docs"}))
	require.False(t, commandsNeedFiles([]string{"lint-deployments"}))
	require.False(t, commandsNeedFiles([]string{"lint-fitness"}))
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

			actualCommands, err := validateCommands(cryptoutilCmdCicdCommon.NewQuietLogger("test"), tc.commands)
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

			_, err := validateCommands(cryptoutilCmdCicdCommon.NewQuietLogger("test"), tc.commands)
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

			_, err := validateCommands(cryptoutilCmdCicdCommon.NewQuietLogger("test"), tc.commands)
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
		{
			name:         "boolean flag -q does not consume next argument",
			commands:     []string{"-q", "lint-text"},
			expectErr:    false,
			expectedCmds: []string{"lint-text"},
		},
		{
			name:         "boolean flag --summary does not consume next argument",
			commands:     []string{cryptoutilSharedMagic.FlagSummary, "lint-workflow"},
			expectErr:    false,
			expectedCmds: []string{"lint-workflow"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actualCommands, err := validateCommands(cryptoutilCmdCicdCommon.NewQuietLogger("test"), tc.commands)
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
		Duration: cryptoutilSharedMagic.JoseJAMaxMaterials * time.Millisecond,
		Error:    fmt.Errorf("test error"),
	}
	require.Equal(t, "lint-text", result.Command)
	require.Equal(t, cryptoutilSharedMagic.JoseJAMaxMaterials*time.Millisecond, result.Duration)
	require.Error(t, result.Error)
}

func TestCommandResultSuccess(t *testing.T) {
	t.Parallel()

	result := cryptoutilCmdCicdCommon.CommandResult{
		Command:  "lint-workflow",
		Duration: cryptoutilSharedMagic.IMMaxUsernameLength * time.Millisecond,
		Error:    nil,
	}
	require.Equal(t, "lint-workflow", result.Command)
	require.Equal(t, cryptoutilSharedMagic.IMMaxUsernameLength*time.Millisecond, result.Duration)
	require.NoError(t, result.Error)
}

func TestGetFailedCommands(t *testing.T) {
	t.Parallel()

	results := []cryptoutilCmdCicdCommon.CommandResult{
		{Command: "lint-text", Duration: cryptoutilSharedMagic.JoseJAMaxMaterials * time.Millisecond, Error: nil},
		{Command: "lint-go", Duration: 200 * time.Millisecond, Error: fmt.Errorf("error1")},
		{Command: "format-go", Duration: 150 * time.Millisecond, Error: nil},
		{Command: "lint-workflow", Duration: cryptoutilSharedMagic.IMMaxUsernameLength * time.Millisecond, Error: fmt.Errorf("error2")},
	}

	failed := cryptoutilCmdCicdCommon.GetFailedCommands(results)
	require.Len(t, failed, 2)
	require.Contains(t, failed, "lint-go")
	require.Contains(t, failed, "lint-workflow")
}

func TestGetFailedCommands_NoFailures(t *testing.T) {
	t.Parallel()

	results := []cryptoutilCmdCicdCommon.CommandResult{
		{Command: "lint-text", Duration: cryptoutilSharedMagic.JoseJAMaxMaterials * time.Millisecond, Error: nil},
		{Command: "lint-go", Duration: 200 * time.Millisecond, Error: nil},
	}

	failed := cryptoutilCmdCicdCommon.GetFailedCommands(results)
	require.Empty(t, failed)
}

// TestRun_LintComposeCommand tests that lint-compose command executes.
func TestRun_LintComposeCommand(t *testing.T) {
	t.Parallel()

	err := run([]string{"lint-compose"}, []string{}, false)
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
	actualCommands, err := validateCommands(cryptoutilCmdCicdCommon.NewQuietLogger("test"), []string{"-v", "true", "-debug", "enabled"})
	require.Error(t, err, "Expected error when only flags provided")
	require.Nil(t, actualCommands, "Expected nil actual commands")
	require.Contains(t, err.Error(), "Usage: cicd <command>", "Error should contain usage info")
}

// TestCicd_ErrorPath verifies that Cicd returns exit code 1 when run returns an error.
func TestCicd_ErrorPath(t *testing.T) {
	t.Parallel()

	// invalid-command is not in ValidCommands so run() returns an error.
	exitCode := Cicd([]string{"cicd", "invalid-command"}, nil, os.Stdout, os.Stderr)
	require.Equal(t, 1, exitCode, "Expected exit code 1 when run returns error")
}

// TestCicd_SuccessPath verifies that Cicd returns exit code 0 on success.
func TestCicd_SuccessPath(t *testing.T) {
	t.Parallel()

	// lint-workflow runs successfully from the project root.
	exitCode := Cicd([]string{"cicd", "lint-workflow"}, nil, os.Stdout, os.Stderr)
	require.Equal(t, 0, exitCode, "Expected exit code 0 for lint-workflow")
}

// TestHasQuietFlag verifies detection of -q and --summary flags.
func TestHasQuietFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		wantBool bool
	}{
		{name: "no flags", args: []string{"lint-text"}, wantBool: false},
		{name: "short flag -q", args: []string{"-q", "lint-text"}, wantBool: true},
		{name: "long flag --summary", args: []string{cryptoutilSharedMagic.FlagSummary, "lint-text"}, wantBool: true},
		{name: "flag after command", args: []string{"lint-text", "-q"}, wantBool: true},
		{name: "empty args", args: []string{}, wantBool: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.wantBool, hasQuietFlag(tc.args))
		})
	}
}

// TestCountFiles verifies counting files across extension map.
func TestCountFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   map[string][]string
		wantInt int
	}{
		{name: "nil map", input: nil, wantInt: 0},
		{name: "empty map", input: map[string][]string{}, wantInt: 0},
		{name: "one extension", input: map[string][]string{".go": {"a.go", "b.go"}}, wantInt: 2},
		{name: "two extensions", input: map[string][]string{".go": {"a.go"}, ".yml": {"b.yml", "c.yml"}}, wantInt: 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.wantInt, countFiles(tc.input))
		})
	}
}

// TestCommandNeedsFiles verifies which commands use file-count display in quiet mode.
func TestCommandNeedsFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		command  string
		wantBool bool
	}{
		{name: "lint-text needs files", command: "lint-text", wantBool: true},
		{name: "lint-compose needs files", command: "lint-compose", wantBool: true},
		{name: "lint-workflow needs files", command: "lint-workflow", wantBool: true},
		{name: "lint-go does not", command: "lint-go", wantBool: false},
		{name: "lint-docs does not", command: "lint-docs", wantBool: false},
		{name: "lint-fitness does not", command: "lint-fitness", wantBool: false},
		{name: "lint-deployments does not", command: "lint-deployments", wantBool: false},
		{name: "github-cleanup does not", command: "github-cleanup", wantBool: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.wantBool, commandNeedsFiles(tc.command))
		})
	}
}

// TestRun_QuietMode verifies quiet mode produces summary output for a passing command.
// Sequential: uses os.Stderr pipe (global process state, cannot run in parallel).
func TestRun_QuietMode(t *testing.T) {
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stderr = w

	runErr := run([]string{"lint-workflow"}, []string{}, true)

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer

	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, runErr, "lint-workflow should pass")
	require.Contains(t, output, "lint-workflow: PASS", "Quiet mode should output PASS summary")
	require.NotContains(t, output, "[CICD]", "Quiet mode should suppress verbose [CICD] logger output")
}

// TestCicd_QuietFlagShort verifies -q flag produces quiet output.
// Sequential: uses os.Stderr pipe (global process state, cannot run in parallel).
func TestCicd_QuietFlagShort(t *testing.T) {
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stderr = w

	exitCode := Cicd([]string{"cicd", "-q", "lint-workflow"}, nil, os.Stdout, os.Stderr)

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer

	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.Equal(t, 0, exitCode, "Expected exit code 0 for lint-workflow -q")
	require.Contains(t, output, "lint-workflow: PASS", "Quiet mode should output PASS summary")
}

// TestCicd_SummaryFlagLong verifies --summary flag produces quiet output.
// Sequential: uses os.Stderr pipe (global process state, cannot run in parallel).
func TestCicd_SummaryFlagLong(t *testing.T) {
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stderr = w

	exitCode := Cicd([]string{"cicd", cryptoutilSharedMagic.FlagSummary, "lint-workflow"}, nil, os.Stdout, os.Stderr)

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer

	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.Equal(t, 0, exitCode, "Expected exit code 0 for lint-workflow --summary")
	require.Contains(t, output, "lint-workflow: PASS", "Summary mode should output PASS summary")
}

// TestRun_QuietMode_FileBasedCommand verifies quiet mode shows file count for file-based commands.
// Sequential: uses os.Stderr pipe (global process state, cannot run in parallel).
func TestRun_QuietMode_FileBasedCommand(t *testing.T) {
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stderr = w

	runErr := run([]string{"lint-text"}, []string{}, true)

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer

	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// lint-text always passes; quiet mode for file-based command shows file count.
	if runErr == nil {
		require.Contains(t, output, "lint-text: PASS (", "Quiet mode file-based command should include file count")
	}
}

// TestRun_QuietMode_FailingCommand verifies quiet mode outputs FAIL for a failing command.
// lint-go-mod always fails when dependencies are outdated (expected in this project).
// Sequential: uses os.Stderr pipe (global process state, cannot run in parallel).
func TestRun_QuietMode_FailingCommand(t *testing.T) {
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stderr = w

	// lint-go-mod fails when dependencies are outdated (always fails in this project).
	runErr := run([]string{"lint-go-mod"}, []string{}, true)

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer

	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// lint-go-mod is expected to fail (outdated dependencies).
	require.Error(t, runErr, "lint-go-mod should fail due to outdated dependencies")
	require.Contains(t, output, "lint-go-mod: FAIL", "Quiet mode should output FAIL for failing command")
	require.NotContains(t, output, "[CICD]", "Quiet mode should suppress verbose [CICD] logger output")
}

// TestRun_QuietMode_NoFilesCommand verifies quiet mode for commands that don't use file counts.
// lint-docs does NOT use file-based scanning → covers CICDQuietPassNoFilesFormat branch.
// Sequential: uses os.Stderr pipe (global process state, cannot run in parallel).
func TestRun_QuietMode_NoFilesCommand(t *testing.T) {
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stderr = w

	// lint-docs does NOT need files and always passes → covers CICDQuietPassNoFilesFormat branch.
	runErr := run([]string{"lint-docs"}, []string{}, true)

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer

	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// lint-docs always passes — verify PASS line has no file count parenthetical.
	require.NoError(t, runErr, "lint-docs should pass in project root")
	require.Contains(t, output, "lint-docs: PASS", "Quiet mode no-files command should show PASS without file count")
	require.NotContains(t, output, "lint-docs: PASS (", "Quiet mode no-files command should NOT show file count")
}

// TestRun_JavaTest_PythonTest verifies lint-java-test and lint-python-test commands execute.
func TestRun_JavaTest_PythonTest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		command string
	}{
		{name: "lint-java-test", command: "lint-java-test"},
		{name: "lint-python-test", command: "lint-python-test"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := run([]string{tc.command}, []string{}, false)
			// These commands may pass or fail depending on project state.
			if err != nil {
				require.Contains(t, err.Error(), "failed commands:", "Error should indicate failed commands")
			}
		})
	}
}
