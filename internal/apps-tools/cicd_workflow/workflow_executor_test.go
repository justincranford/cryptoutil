// Copyright (c) 2025-2026 Justin Cranford.
//
//

package cicd_workflow

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	workflowTestChildFailureEnv               = "CICD_WORKFLOW_TEST_CHILD_FAILURE"
	workflowTestNameMock                      = "mock"
	workflowTestNameQuality                   = "quality"
	workflowTestNameCoverage                  = "coverage"
	workflowTestTitleCoverage                 = "Coverage"
	workflowTestDescriptionQualityChecks      = "Quality checks"
	workflowTestDescriptionCoverageCollection = "Coverage collection"
	workflowTestArgWorkflowsMock              = "-workflows=mock"
	workflowTestArgDryRun                     = "-dry-run"
	workflowTestArgVerbose                    = "-act-args=--verbose"
	workflowTestCustomOutputDir               = "custom-output"
	workflowTestCaseWithWarnings              = "with warnings"
	workflowTestNameReadFailure               = "read_failure"
	workflowTestNameFirst                     = "first"
	workflowTestNameSecond                    = "second"
	workflowTestNameThird                     = "third"
	workflowTestTitleQuality                  = "Quality"
)

func TestRun_DirectCall(t *testing.T) {
	t.Parallel()

	result := run([]string{})
	require.Equal(t, 1, result)
}

func TestRunWithWorkflowsDir_ParseFlagError(t *testing.T) {
	t.Parallel()

	result := runWithWorkflowsDir([]string{"--invalid-flag-xyz=bad"}, makeWorkflowsDir(t, workflowTestNameMock))
	require.Equal(t, 1, result)
}

func TestRunWithWorkflowsDir_NonDryRun_ProcessResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		workflowName   string
		commandFails   bool
		wantExitStatus int
		wantEvent      string
	}{
		{
			name:           "successful regular workflow",
			workflowName:   workflowTestNameMock,
			wantExitStatus: 0,
			wantEvent:      cryptoutilSharedMagic.EventTypePush,
		},
		{
			name:           "failed regular workflow",
			workflowName:   workflowTestNameMock,
			commandFails:   true,
			wantExitStatus: 1,
			wantEvent:      cryptoutilSharedMagic.EventTypePush,
		},
		{
			name:           "DAST workflow uses dispatch event",
			workflowName:   cryptoutilSharedMagic.WorkflowNameDAST,
			wantExitStatus: 0,
			wantEvent:      cryptoutilSharedMagic.EventTypeWorkflowDispatch,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dependencies := newWorkflowDependencies()

			var actualArgs []string

			dependencies.commandContextFn = func(ctx context.Context, _ string, args ...string) *exec.Cmd {
				actualArgs = append([]string(nil), args...)

				return workflowTestProcess(t, ctx, tc.commandFails)
			}

			result := runWithWorkflowsDirAndDependencies([]string{
				"-workflows=" + tc.workflowName,
				workflowTestArgVerbose,
				"-output=" + filepath.Join(t.TempDir(), "output"),
			}, makeWorkflowsDir(t, tc.workflowName), dependencies)

			require.Equal(t, tc.wantExitStatus, result)
			require.NotEmpty(t, actualArgs)
			require.Equal(t, tc.wantEvent, actualArgs[0])
			require.Contains(t, actualArgs, "--verbose")
		})
	}
}

func TestRunWithWorkflowsDir_OutputSetupFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		failDirectory bool
	}{
		{name: "output directory creation failure", failDirectory: true},
		{name: "combined log open failure"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dependencies := newWorkflowDependencies()
			if tc.failDirectory {
				dependencies.makeDirectoryFn = func(string, os.FileMode) error {
					return errors.New("injected directory creation failure")
				}
			} else {
				dependencies.openFileFn = func(string, int, os.FileMode) (*os.File, error) {
					return nil, errors.New("injected combined log open failure")
				}
			}

			result := runWithWorkflowsDirAndDependencies([]string{
				workflowTestArgWorkflowsMock,
				workflowTestArgDryRun,
				"-output=" + filepath.Join(t.TempDir(), "output"),
			}, makeWorkflowsDir(t, workflowTestNameMock), dependencies)

			require.Equal(t, 1, result)
		})
	}
}

func TestExecuteWorkflow_OpenLogFailure(t *testing.T) {
	t.Parallel()

	dependencies := newWorkflowDependencies()
	dependencies.openFileFn = func(string, int, os.FileMode) (*os.File, error) {
		return nil, errors.New("injected workflow log open failure")
	}

	result := executeWorkflowWithDependencies(
		WorkflowExecution{Name: workflowTestNameMock}, nil, t.TempDir(), false, "unused", "", dependencies,
	)

	require.False(t, result.Success)
	require.Contains(t, result.ErrorMessages, "injected workflow log open failure")
}

func TestExecuteWorkflow_PipeSetupFailure(t *testing.T) {
	t.Parallel()

	dependencies := newWorkflowDependencies()
	dependencies.setupCmdPipesFn = func(*exec.Cmd) (io.ReadCloser, io.ReadCloser, error) {
		return nil, nil, errors.New("injected command pipe setup failure")
	}

	result := executeWorkflowWithDependencies(
		WorkflowExecution{Name: workflowTestNameMock}, nil, t.TempDir(), false, "unused", "", dependencies,
	)

	require.False(t, result.Success)
	require.Contains(t, result.ErrorMessages, "injected command pipe setup failure")
}

func TestSetupCommandPipes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		configure   func(*exec.Cmd)
		wantErrText string
	}{
		{name: "creates both pipes"},
		{
			name: "stdout already configured",
			configure: func(command *exec.Cmd) {
				command.Stdout = io.Discard
			},
			wantErrText: "create stdout pipe",
		},
		{
			name: "stderr already configured",
			configure: func(command *exec.Cmd) {
				command.Stderr = io.Discard
			},
			wantErrText: "create stderr pipe",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			command := &exec.Cmd{}
			if tc.configure != nil {
				tc.configure(command)
			}

			stdout, stderr, err := setupCommandPipes(command)
			if tc.wantErrText != "" {
				require.ErrorContains(t, err, tc.wantErrText)
				require.Nil(t, stdout)
				require.Nil(t, stderr)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, stdout)
			require.NotNil(t, stderr)
			require.NoError(t, stdout.Close())
			require.NoError(t, stderr.Close())

			stdoutWriter, ok := command.Stdout.(*os.File)
			require.True(t, ok)

			stdoutCloseErr := stdoutWriter.Close()
			require.NoError(t, stdoutCloseErr)

			stderrWriter, ok := command.Stderr.(*os.File)
			require.True(t, ok)

			stderrCloseErr := stderrWriter.Close()
			require.NoError(t, stderrCloseErr)
		})
	}
}

func TestRunWithWorkflowsDir_CloseFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		failFileNameWith string
	}{
		{name: "workflow log close warning", failFileNameWith: "mock-"},
		{name: "combined log close warning", failFileNameWith: "combined-"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dependencies := newWorkflowDependencies()

			var injectedFailure bool

			dependencies.commandContextFn = func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
				return workflowTestProcess(t, ctx, false)
			}
			dependencies.closeFileFn = func(file *os.File) error {
				if strings.HasPrefix(filepath.Base(file.Name()), tc.failFileNameWith) {
					injectedFailure = true
					_ = file.Close()

					return errors.New("injected file close failure")
				}

				return file.Close()
			}

			result := runWithWorkflowsDirAndDependencies([]string{
				workflowTestArgWorkflowsMock,
				"-output=" + filepath.Join(t.TempDir(), "output"),
			}, makeWorkflowsDir(t, workflowTestNameMock), dependencies)

			require.Equal(t, 0, result)
			require.True(t, injectedFailure)
		})
	}
}

func TestWorkflowSubprocessHelper(t *testing.T) {
	t.Parallel()

	value, ok := os.LookupEnv(workflowTestChildFailureEnv)
	if !ok {
		return
	}

	shouldFail, err := strconv.ParseBool(value)
	require.NoError(t, err)

	if shouldFail {
		t.Fatal("injected workflow subprocess failure")
	}
}

func workflowTestProcess(t *testing.T, ctx context.Context, shouldFail bool) *exec.Cmd {
	t.Helper()

	executable, err := os.Executable()
	require.NoError(t, err)

	command := exec.CommandContext(ctx, executable, "-test.run=^TestWorkflowSubprocessHelper$")
	envPrefix := workflowTestChildFailureEnv + "="

	command.Env = make([]string, 0, len(os.Environ())+1)
	for _, envVar := range os.Environ() {
		if !strings.HasPrefix(envVar, envPrefix) {
			command.Env = append(command.Env, envVar)
		}
	}

	command.Env = append(command.Env, envPrefix+strconv.FormatBool(shouldFail))

	return command
}

func TestGetWorkflowDescription_FilePaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		workflowName string
		content      string
		want         string
	}{
		{name: "workflow name field", workflowName: "mytest", content: "name: My Test Workflow\non:\n  push:\n", want: "My Test Workflow"},
		{name: "fallback when name is absent", workflowName: "noname", content: "description: no name here\non:\n  push:\n", want: "Noname workflow"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			workflowFile := filepath.Join(t.TempDir(), "ci-"+tc.workflowName+".yml")
			err := os.WriteFile(workflowFile, []byte(tc.content), cryptoutilSharedMagic.FilePermOwnerReadWriteGroupRead)
			require.NoError(t, err)

			result := getWorkflowDescription(workflowFile, tc.workflowName, os.ReadFile)

			require.Equal(t, tc.want, result)
		})
	}
}
