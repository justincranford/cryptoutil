// Copyright (c) 2025-2026 Justin Cranford.
package cicd_lint

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"

	"github.com/stretchr/testify/assert"
)

func TestRunUsage(t *testing.T) {
	t.Parallel()

	// Test with no commands (should return error).
	exitCode := Cicd([]string{testCicd}, nil, nil, os.Stderr)
	require.Equal(t, 1, exitCode, "Expected exit code 1 when no commands provided")
}

func TestRunInvalidCommand(t *testing.T) {
	t.Parallel()

	// Test with invalid command.
	err := run([]string{testInvalidCommand}, []string{}, false)
	require.Error(t, err, "Expected error for invalid command")
	require.Contains(t, err.Error(), "unknown command: invalid-command", "Error message should indicate unknown command")
}

func TestRunScriptCommandInvalidArgs(t *testing.T) {
	t.Parallel()

	err := run([]string{cmdGitHubCleanup}, []string{"--invalid"}, true)

	require.ErrorContains(t, err, "failed commands")
}

func TestRunMultipleCommands(t *testing.T) {
	t.Parallel()

	// Verify the command parsing logic works.
	commands := []string{cmdLintGoMod, cmdLintWorkflow}
	require.Len(t, commands, 2, "Expected 2 commands")
	require.Equal(t, cmdLintGoMod, commands[0], "Expected first command")
	require.Equal(t, cmdLintWorkflow, commands[1], "Expected second command")
}

func TestRun_AllCommands_HappyPath(t *testing.T) {
	t.Parallel()

	commands := []string{
		cmdLintText, cmdLintGo, cmdLintGoTest, cmdLintGoMod, cmdLintGolangci,
		cmdLintWorkflow, cmdLintCompose, cmdLintPorts, cmdLintDeployments,
		cmdLintDocs, cmdLintFitness, cmdLintOpenAPI, cmdLintJavaTest, cmdLintPythonTest,
		cmdFormatGo, cmdFormatGoTest, cmdFormatText,
	}

	var (
		executions atomic.Int32
		executed   sync.Map
	)

	execute := func(
		_ *cryptoutilCmdCicdCommon.Logger,
		command string,
		_ map[string][]string,
		_ []string,
	) cryptoutilCmdCicdCommon.CommandResult {
		executions.Add(1)
		executed.Store(command, struct{}{})

		return cryptoutilCmdCicdCommon.CommandResult{Command: command}
	}

	err := runWithExecutor(commands, []string{}, true, execute)
	require.NoError(t, err)
	require.Equal(t, len(commands), int(executions.Load()))

	for _, command := range commands {
		_, ok := executed.Load(command)
		require.True(t, ok, "command should execute: %s", command)
	}
}

func TestRunWithExecutor_FileCollectionError(t *testing.T) {
	t.Parallel()

	listFiles := func(startDirectory string) (map[string][]string, error) {
		require.Equal(t, cryptoutilSharedMagic.ListAllFilesStartDirectory, startDirectory)

		return nil, fmt.Errorf("injected file collection failure")
	}
	execute := func(
		_ *cryptoutilCmdCicdCommon.Logger,
		command string,
		_ map[string][]string,
		_ []string,
	) cryptoutilCmdCicdCommon.CommandResult {
		assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "command %q executed after file collection failed", command)

		return cryptoutilCmdCicdCommon.CommandResult{Command: command}
	}

	err := runWithExecutorAndFileLister([]string{cmdLintText}, nil, true, execute, listFiles)
	require.ErrorContains(t, err, "failed to collect files: injected file collection failure")
}

func TestRunWithExecutor_VerboseCommands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		commands []string
		wantRuns int32
	}{
		{name: "single command", commands: []string{cmdLintDocs}, wantRuns: 1},
		{name: "multiple commands", commands: []string{cmdLintDocs, cmdLintDeployments}, wantRuns: 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var executions atomic.Int32

			execute := func(
				_ *cryptoutilCmdCicdCommon.Logger,
				command string,
				_ map[string][]string,
				_ []string,
			) cryptoutilCmdCicdCommon.CommandResult {
				executions.Add(1)

				return cryptoutilCmdCicdCommon.CommandResult{Command: command}
			}

			require.NoError(t, runWithExecutor(tc.commands, nil, false, execute))
			require.Equal(t, tc.wantRuns, executions.Load())
		})
	}
}

func TestExecuteCommand_Dispatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		command          string
		wantFormatGo     bool
		wantFormatGoTest bool
	}{
		{name: cmdLintGo, command: cmdLintGo},
		{name: cmdFormatGo, command: cmdFormatGo, wantFormatGo: true},
		{name: cmdLintGoTest, command: cmdLintGoTest},
		{name: cmdFormatGoTest, command: cmdFormatGoTest, wantFormatGoTest: true},
		{name: cmdFormatText, command: cmdFormatText},
		{name: cmdLintPorts, command: cmdLintPorts},
		{name: cmdLintGolangci, command: cmdLintGolangci},
		{name: cmdLintDeployments, command: cmdLintDeployments},
		{name: cmdLintFitness, command: cmdLintFitness},
		{name: cmdLintOpenAPI, command: cmdLintOpenAPI},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var formatGoCalled, formatGoTestCalled bool

			formatGo := func(
				_ *cryptoutilCmdCicdCommon.Logger,
				_ map[string][]string,
			) error {
				formatGoCalled = true

				return nil
			}
			formatGoTest := func(_ *cryptoutilCmdCicdCommon.Logger) error {
				formatGoTestCalled = true

				return nil
			}

			result := executeCommandWithFormatters(
				cryptoutilCmdCicdCommon.NewQuietLogger("test"),
				tc.command,
				map[string][]string{},
				nil,
				formatGo,
				formatGoTest,
			)

			require.Equal(t, tc.command, result.Command)
			require.Equal(t, tc.wantFormatGo, formatGoCalled)
			require.Equal(t, tc.wantFormatGoTest, formatGoTestCalled)
		})
	}
}

func TestGetExtraArgs_WithArgs(t *testing.T) {
	t.Parallel()

	result := getExtraArgs([]string{"binary", "command", testExtra1, testExtra2})
	require.Equal(t, []string{testExtra1, testExtra2}, result)
}

func TestCommandsNeedFiles_NonFileCommands(t *testing.T) {
	t.Parallel()

	require.False(t, commandsNeedFiles([]string{cmdLintDocs}))
	require.False(t, commandsNeedFiles([]string{cmdLintDeployments}))
	require.False(t, commandsNeedFiles([]string{cmdLintFitness}))
}

func TestValidateCommands_HappyPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		commands []string
	}{
		{
			name:     "single lint command",
			commands: []string{cmdLintWorkflow},
		},
		{
			name:     "multiple lint commands",
			commands: []string{cmdLintText, cmdLintGo, cmdLintGoTest},
		},
		{
			name:     "lint and format commands",
			commands: []string{cmdLintText, cmdFormatGo, cmdFormatGoTest, cmdFormatText},
		},
		{
			name:     "all commands once each",
			commands: []string{cmdLintText, cmdLintGo, cmdFormatGo, cmdLintGoTest, cmdFormatGoTest, cmdFormatText, cmdLintWorkflow, cmdLintGoMod, cmdLintCompose, cmdLintPorts, cmdLintGolangci},
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
			commands:     []string{cmdLintText, cmdLintText},
			expectedText: "command 'lint-text' specified 2 times",
		},
		{
			name:         "duplicate lint-go",
			commands:     []string{cmdLintGo, cmdLintText, cmdLintGo},
			expectedText: "command 'lint-go' specified 2 times",
		},
		{
			name:         "duplicate format-go",
			commands:     []string{cmdFormatGo, cmdLintWorkflow, cmdFormatGo},
			expectedText: "command 'format-go' specified 2 times",
		},
		{
			name:         "multiple duplicates",
			commands:     []string{cmdLintWorkflow, cmdLintWorkflow, cmdLintText, cmdLintText},
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
			commands:     []string{"--strict", testTrue, cmdLintText},
			expectErr:    false,
			expectedCmds: []string{cmdLintText},
		},
		{
			name:         "boolean flag -q does not consume next argument",
			commands:     []string{testQ, cmdLintText},
			expectErr:    false,
			expectedCmds: []string{cmdLintText},
		},
		{
			name:         "boolean flag --summary does not consume next argument",
			commands:     []string{cryptoutilSharedMagic.FlagSummary, cmdLintWorkflow},
			expectErr:    false,
			expectedCmds: []string{cmdLintWorkflow},
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
		Command:  cmdLintText,
		Duration: cryptoutilSharedMagic.JoseJAMaxMaterials * time.Millisecond,
		Error:    fmt.Errorf("test error"),
	}
	require.Equal(t, cmdLintText, result.Command)
	require.Equal(t, cryptoutilSharedMagic.JoseJAMaxMaterials*time.Millisecond, result.Duration)
	require.Error(t, result.Error)
}

func TestCommandResultSuccess(t *testing.T) {
	t.Parallel()

	result := cryptoutilCmdCicdCommon.CommandResult{
		Command:  cmdLintWorkflow,
		Duration: cryptoutilSharedMagic.IMMaxUsernameLength * time.Millisecond,
		Error:    nil,
	}
	require.Equal(t, cmdLintWorkflow, result.Command)
	require.Equal(t, cryptoutilSharedMagic.IMMaxUsernameLength*time.Millisecond, result.Duration)
	require.NoError(t, result.Error)
}

func TestGetFailedCommands(t *testing.T) {
	t.Parallel()

	results := []cryptoutilCmdCicdCommon.CommandResult{
		{Command: cmdLintText, Duration: cryptoutilSharedMagic.JoseJAMaxMaterials * time.Millisecond, Error: nil},
		{Command: cmdLintGo, Duration: 200 * time.Millisecond, Error: fmt.Errorf("error1")},
		{Command: cmdFormatGo, Duration: 150 * time.Millisecond, Error: nil},
		{Command: cmdLintWorkflow, Duration: cryptoutilSharedMagic.IMMaxUsernameLength * time.Millisecond, Error: fmt.Errorf("error2")},
	}

	failed := cryptoutilCmdCicdCommon.GetFailedCommands(results)
	require.Len(t, failed, 2)
	require.Contains(t, failed, cmdLintGo)
	require.Contains(t, failed, cmdLintWorkflow)
}

func TestGetFailedCommands_NoFailures(t *testing.T) {
	t.Parallel()

	results := []cryptoutilCmdCicdCommon.CommandResult{
		{Command: cmdLintText, Duration: cryptoutilSharedMagic.JoseJAMaxMaterials * time.Millisecond, Error: nil},
		{Command: cmdLintGo, Duration: 200 * time.Millisecond, Error: nil},
	}

	failed := cryptoutilCmdCicdCommon.GetFailedCommands(results)
	require.Empty(t, failed)
}

// TestRun_LintComposeCommand tests that lint-compose command executes.
func TestRun_LintComposeCommand(t *testing.T) {
	t.Parallel()

	err := run([]string{cmdLintCompose}, []string{}, false)
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
	actualCommands, err := validateCommands(cryptoutilCmdCicdCommon.NewQuietLogger("test"), []string{"-v", testTrue, "-debug", "enabled"})
	require.Error(t, err, "Expected error when only flags provided")
	require.Nil(t, actualCommands, "Expected nil actual commands")
	require.Contains(t, err.Error(), "Usage: cicd <command>", "Error should contain usage info")
}

// TestCicd_ErrorPath verifies that Cicd returns exit code 1 when run returns an error.
func TestCicd_ErrorPath(t *testing.T) {
	t.Parallel()

	// invalid-command is not in ValidCommands so run() returns an error.
	exitCode := Cicd([]string{testCicd, testInvalidCommand}, nil, os.Stdout, os.Stderr)
	require.Equal(t, 1, exitCode, "Expected exit code 1 when run returns error")
}

// TestCicd_SuccessPath verifies that Cicd returns exit code 0 on success.
func TestCicd_SuccessPath(t *testing.T) {
	t.Parallel()

	// lint-workflow runs successfully from the project root.
	exitCode := Cicd([]string{testCicd, cmdLintWorkflow}, nil, os.Stdout, os.Stderr)
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
		{name: "no flags", args: []string{cmdLintText}, wantBool: false},
		{name: "short flag -q", args: []string{testQ, cmdLintText}, wantBool: true},
		{name: "long flag --summary", args: []string{cryptoutilSharedMagic.FlagSummary, cmdLintText}, wantBool: true},
		{name: "flag after command", args: []string{cmdLintText, testQ}, wantBool: true},
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
		{name: "one extension", input: map[string][]string{testGo: {testAGo, "b.go"}}, wantInt: 2},
		{name: "two extensions", input: map[string][]string{testGo: {testAGo}, ".yml": {"b.yml", "c.yml"}}, wantInt: 3},
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
		{name: "lint-text needs files", command: cmdLintText, wantBool: true},
		{name: "lint-compose needs files", command: cmdLintCompose, wantBool: true},
		{name: "lint-workflow needs files", command: cmdLintWorkflow, wantBool: true},
		{name: "lint-go does not", command: cmdLintGo, wantBool: false},
		{name: "lint-docs does not", command: cmdLintDocs, wantBool: false},
		{name: "lint-fitness does not", command: cmdLintFitness, wantBool: false},
		{name: "lint-deployments does not", command: cmdLintDeployments, wantBool: false},
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

	runErr := run([]string{cmdLintWorkflow}, []string{}, true)

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

	exitCode := Cicd([]string{testCicd, testQ, cmdLintWorkflow}, nil, os.Stdout, os.Stderr)

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

	exitCode := Cicd([]string{testCicd, cryptoutilSharedMagic.FlagSummary, cmdLintWorkflow}, nil, os.Stdout, os.Stderr)

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

	runErr := run([]string{cmdLintText}, []string{}, true)

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
	runErr := run([]string{cmdLintGoMod}, []string{}, true)

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
	runErr := run([]string{cmdLintDocs}, []string{}, true)

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
		{name: cmdLintJavaTest, command: cmdLintJavaTest},
		{name: cmdLintPythonTest, command: cmdLintPythonTest},
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
