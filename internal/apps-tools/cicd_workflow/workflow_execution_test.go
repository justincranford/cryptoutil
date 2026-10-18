// Copyright (c) 2025-2026 Justin Cranford.
//
//

package cicd_workflow

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// makeWorkflowsDir creates a temp dir with .github/workflows/ci-<name>.yml files
// and returns the abs path to the .github/workflows directory.
func makeWorkflowsDir(t *testing.T, workflowNames ...string) string {
	t.Helper()

	tempDir := t.TempDir()
	wfDir := filepath.Join(tempDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "workflows")
	err := os.MkdirAll(wfDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
	require.NoError(t, err)

	names := workflowNames
	if len(names) == 0 {
		names = []string{workflowTestNameMock}
	}

	for _, name := range names {
		content := "name: Mock Workflow\non:\n  push:\n"
		err = os.WriteFile(
			filepath.Join(wfDir, "ci-"+name+".yml"),
			[]byte(content),
			cryptoutilSharedMagic.FilePermOwnerReadWriteGroupRead,
		)
		require.NoError(t, err)
	}

	return wfDir
}

// TestRunWithWorkflowsDir_MissingDir tests when the workflows directory doesn't exist.
func TestRunWithWorkflowsDir_MissingDir(t *testing.T) {
	t.Parallel()

	result := runWithWorkflowsDir([]string{workflowTestArgWorkflowsMock}, "/nonexistent/workflows/dir")
	require.Equal(t, 1, result)
}

// TestRunWithWorkflowsDir_HelpFlag tests the --help flag path.
func TestRunWithWorkflowsDir_HelpFlag(t *testing.T) {
	t.Parallel()

	wfDir := makeWorkflowsDir(t, workflowTestNameQuality, workflowTestNameCoverage)
	result := runWithWorkflowsDir([]string{"-help"}, wfDir)
	require.Equal(t, 0, result)
}

// TestRunWithWorkflowsDir_NoWorkflowsFlag tests run with no -workflows flag specified.
func TestRunWithWorkflowsDir_NoWorkflowsFlag(t *testing.T) {
	t.Parallel()

	wfDir := makeWorkflowsDir(t)
	result := runWithWorkflowsDir([]string{}, wfDir)
	require.Equal(t, 1, result)
}

// TestRunWithWorkflowsDir_InvalidWorkflow tests when all specified workflows are invalid.
func TestRunWithWorkflowsDir_InvalidWorkflow(t *testing.T) {
	t.Parallel()

	wfDir := makeWorkflowsDir(t)
	result := runWithWorkflowsDir([]string{"-workflows=nonexistent_workflow_xyz"}, wfDir)
	require.Equal(t, 1, result)
}

// TestRunWithWorkflowsDir_DryRun tests dry-run mode.
func TestRunWithWorkflowsDir_DryRun(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	wfDir := makeWorkflowsDir(t, workflowTestNameMock)
	outputDir := filepath.Join(tempDir, "output")

	result := runWithWorkflowsDir([]string{
		workflowTestArgWorkflowsMock,
		workflowTestArgDryRun,
		"-output=" + outputDir,
	}, wfDir)

	require.Equal(t, 0, result)
}

// TestRunWithWorkflowsDir_DryRun_DastWorkflow tests dry-run with the dast workflow (special event).
func TestRunWithWorkflowsDir_DryRun_DastWorkflow(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	wfDir := makeWorkflowsDir(t, cryptoutilSharedMagic.WorkflowNameDAST)
	outputDir := filepath.Join(tempDir, "output")

	result := runWithWorkflowsDir([]string{
		"-workflows=dast",
		workflowTestArgDryRun,
		"-output=" + outputDir,
	}, wfDir)

	require.Equal(t, 0, result)
}

// TestRunWithWorkflowsDir_DryRun_WithActArgs tests dry-run with extra act arguments.
func TestRunWithWorkflowsDir_DryRun_WithActArgs(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	wfDir := makeWorkflowsDir(t, workflowTestNameQuality)
	outputDir := filepath.Join(tempDir, "output")

	result := runWithWorkflowsDir([]string{
		"-workflows=quality",
		workflowTestArgDryRun,
		"-output=" + outputDir,
		workflowTestArgVerbose,
	}, wfDir)

	require.Equal(t, 0, result)
}

// TestRunWithWorkflowsDir_NonDryRun_UsesInjectedCommand exercises execution without external binaries.
func TestRunWithWorkflowsDir_NonDryRun_UsesInjectedCommand(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	wfDir := makeWorkflowsDir(t, workflowTestNameMock)
	outputDir := filepath.Join(tempDir, "output")

	dependencies := newWorkflowDependencies()
	dependencies.commandContextFn = func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
		return workflowTestProcess(t, ctx, false)
	}

	result := runWithWorkflowsDirAndDependencies([]string{
		workflowTestArgWorkflowsMock,
		"-output=" + outputDir,
	}, wfDir, dependencies)

	require.Equal(t, 0, result)
}

// TestRunWithWorkflowsDir_NonDryRun_ActNotFound tests non-dry-run when act binary doesn't exist.
func TestRunWithWorkflowsDir_NonDryRun_ActNotFound(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	wfDir := makeWorkflowsDir(t, workflowTestNameMock)
	outputDir := filepath.Join(tempDir, "output")

	result := runWithWorkflowsDir([]string{
		workflowTestArgWorkflowsMock,
		"-act-path=/nonexistent/bin/fake-act-xyz",
		"-output=" + outputDir,
	}, wfDir)

	// Should fail since act binary doesn't exist.
	require.Equal(t, 1, result)
}

// TestRunWithWorkflowsDir_OutputDirCreationFails tests when output dir creation fails.
func TestRunWithWorkflowsDir_OutputDirCreationFails(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	blockingFile := filepath.Join(tempDir, workflowTestNameMock)
	require.NoError(t, os.WriteFile(blockingFile, nil, cryptoutilSharedMagic.FilePermOwnerReadWriteGroupRead))

	wfDir := makeWorkflowsDir(t, workflowTestNameMock)

	// A regular file as the parent path forces directory creation to fail on all platforms.
	result := runWithWorkflowsDir([]string{
		workflowTestArgWorkflowsMock,
		workflowTestArgDryRun,
		"-output=" + filepath.Join(blockingFile, workflowTestNameQuality),
	}, wfDir)

	require.Equal(t, 1, result)
}

// TestWorkflow_EntryPoint tests the Workflow() public entry point via the "run" subcommand.
func TestWorkflow_EntryPoint(t *testing.T) {
	t.Parallel()

	// Workflow requires args[0] as program name and args[1] as subcommand.
	// Test the "run" subcommand error path (invalid workflows flag).
	result := Workflow([]string{cryptoutilSharedMagic.CICDCmdDirWorkflow, cryptoutilSharedMagic.WorkflowSubCmdRun, "-workflows=nonexistent"}, nil, nil, nil)
	// Returns 1 because no valid workflow named "nonexistent" exists.
	require.Equal(t, 1, result)
}

// TestWorkflow_NoArgs tests Workflow with fewer than 2 args (no subcommand provided).
func TestWorkflow_NoArgs(t *testing.T) {
	t.Parallel()

	result := Workflow([]string{cryptoutilSharedMagic.CICDCmdDirWorkflow}, nil, nil, io.Discard)
	require.Equal(t, 1, result)
}

// TestWorkflow_CleanupSubcommand_Success tests cleanup subcommand success path via fn-param injection.
func TestWorkflow_CleanupSubcommand_Success(t *testing.T) {
	t.Parallel()

	stubCleanupFn := func(_ []string, _ io.Writer) int { return 0 }

	result := workflow([]string{cryptoutilSharedMagic.CICDCmdDirWorkflow, cryptoutilSharedMagic.WorkflowSubCmdCleanup}, nil, nil, nil, stubCleanupFn)
	require.Equal(t, 0, result)
}

// TestWorkflow_CleanupSubcommand_Error tests cleanup subcommand error path via fn-param injection.
func TestWorkflow_CleanupSubcommand_Error(t *testing.T) {
	t.Parallel()

	stubCleanupFn := func(_ []string, _ io.Writer) int { return 1 }

	result := workflow([]string{cryptoutilSharedMagic.CICDCmdDirWorkflow, cryptoutilSharedMagic.WorkflowSubCmdCleanup}, nil, nil, nil, stubCleanupFn)
	require.Equal(t, 1, result)
}

// TestWorkflow_UnknownSubcommand tests the default case for unknown subcommands.
func TestWorkflow_UnknownSubcommand(t *testing.T) {
	t.Parallel()

	result := Workflow([]string{cryptoutilSharedMagic.CICDCmdDirWorkflow, "invalid-cmd"}, nil, nil, io.Discard)
	require.Equal(t, 1, result)
}

// TestDefaultCleanup_ParseError tests defaultCleanup with an invalid flag (ParseArgs failure).
func TestDefaultCleanup_ParseError(t *testing.T) {
	t.Parallel()

	result := defaultCleanup([]string{"--xyzzy-invalid-flag=bad"}, io.Discard)
	require.Equal(t, 1, result)
}

// TestPrintHelp tests the printHelp function.
func TestPrintHelp(t *testing.T) {
	t.Parallel()

	workflows := map[string]WorkflowConfig{
		workflowTestNameQuality:  {Description: workflowTestDescriptionQualityChecks},
		workflowTestNameCoverage: {Description: workflowTestDescriptionCoverageCollection},
	}

	// printHelp writes to os.Stdout directly; just verify it doesn't panic.
	require.NotPanics(t, func() {
		printHelp(workflows)
	})
}

// TestPrintHelp_Empty tests printHelp with an empty workflow map.
func TestPrintHelp_Empty(t *testing.T) {
	t.Parallel()

	require.NotPanics(t, func() {
		printHelp(map[string]WorkflowConfig{})
	})
}

// TestParseWorkflowNames tests workflow name parsing.
func TestParseWorkflowNames(t *testing.T) {
	t.Parallel()

	available := map[string]WorkflowConfig{
		workflowTestNameQuality:                {Description: workflowTestDescriptionQualityChecks},
		workflowTestNameCoverage:               {Description: workflowTestTitleCoverage},
		cryptoutilSharedMagic.WorkflowNameDAST: {Description: "DAST"},
	}

	tests := []struct {
		name          string
		input         string
		expectedNames []string
	}{
		{
			name:          "single valid workflow",
			input:         workflowTestNameQuality,
			expectedNames: []string{workflowTestNameQuality},
		},
		{
			name:          "multiple valid workflows",
			input:         "quality,coverage",
			expectedNames: []string{workflowTestNameQuality, workflowTestNameCoverage},
		},
		{
			name:          "mixed valid and invalid",
			input:         "quality,nonexistent,coverage",
			expectedNames: []string{workflowTestNameQuality, workflowTestNameCoverage},
		},
		{
			name:          "all invalid",
			input:         "invalid1,invalid2",
			expectedNames: []string{},
		},
		{
			name:          "with whitespace",
			input:         "quality, coverage",
			expectedNames: []string{workflowTestNameQuality, workflowTestNameCoverage},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := parseWorkflowNames(tc.input, available)

			resultNames := make([]string, 0, len(result))
			for _, wf := range result {
				resultNames = append(resultNames, wf.Name)
			}

			for _, expected := range tc.expectedNames {
				require.Contains(t, resultNames, expected)
			}

			require.Equal(t, len(tc.expectedNames), len(result))
		})
	}
}

// TestPrintExecutiveSummaryHeader tests the header printing function.
