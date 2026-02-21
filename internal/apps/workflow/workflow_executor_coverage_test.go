// Copyright (c) 2025 Justin Cranford
//
//

package workflow

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// varMu guards mutations of package-level function variables used for injection.
var varMu sync.Mutex

// TestRun_DirectCall exercises the run() wrapper (0% coverage baseline).
// Not parallel: run() uses ".github/workflows" relative to CWD; no .github here.
func TestRun_DirectCall(t *testing.T) {
	result := run([]string{})
	require.Equal(t, 1, result) // Fails: no .github/workflows in package dir.
}

// TestRunWithWorkflowsDir_ParseFlagError tests flag parse error path.
func TestRunWithWorkflowsDir_ParseFlagError(t *testing.T) {
	t.Parallel()

	wfDir := makeWorkflowsDir(t, "mock")
	// --invalid-flag-xyz is not registered; ContinueOnError returns error.
	result := runWithWorkflowsDir([]string{"--invalid-flag-xyz=bad"}, wfDir)
	require.Equal(t, 1, result)
}

// TestRunWithWorkflowsDir_NonDryRun_FalseBinary tests cmd.Wait failure path.
// /bin/false exits with code 1, causing result.Success=false and the failed footer.
func TestRunWithWorkflowsDir_NonDryRun_FalseBinary(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	wfDir := makeWorkflowsDir(t, "mock")
	outputDir := filepath.Join(tempDir, "output")

	result := runWithWorkflowsDir([]string{
		"-workflows=mock",
		"-act-path=/bin/false",
		"-act-args=extra-arg",
		"-output=" + outputDir,
	}, wfDir)

	require.Equal(t, 1, result) // /bin/false exits 1; TotalFailed > 0.
}

// TestRunWithWorkflowsDir_NonDryRun_DastWorkflow tests DAST event type selection.
// WorkflowNameDAST triggers workflow_dispatch event instead of push in non-dry-run.
func TestRunWithWorkflowsDir_NonDryRun_DastWorkflow(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	wfDir := makeWorkflowsDir(t, cryptoutilSharedMagic.WorkflowNameDAST)
	outputDir := filepath.Join(tempDir, "output")

	result := runWithWorkflowsDir([]string{
		"-workflows=" + cryptoutilSharedMagic.WorkflowNameDAST,
		"-act-path=/bin/false",
		"-output=" + outputDir,
	}, wfDir)

	// /bin/false exits 1 but DAST event selection code path is exercised.
	require.Equal(t, 1, result)
}

// TestRunWithWorkflowsDir_CombinedLogCreationFails tests the OpenFile failure path.
// Creates outputDir as read-only so the combined log file cannot be created.
func TestRunWithWorkflowsDir_CombinedLogCreationFails(t *testing.T) {
	t.Parallel()

	wfDir := makeWorkflowsDir(t, "mock")
	parentDir := t.TempDir()
	outputDir := filepath.Join(parentDir, "noaccess")

	err := os.MkdirAll(outputDir, cryptoutilSharedMagic.FilePermOwnerReadOnlyGroupOtherReadOnly)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.Chmod(outputDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
	})

	result := runWithWorkflowsDir([]string{
		"-workflows=mock",
		"-dry-run",
		"-output=" + outputDir,
	}, wfDir)

	require.Equal(t, 1, result)
}

// TestExecuteWorkflow_WorkflowLogCreationFailure tests the workflow log failure path.
// Valid combinedLog (non-nil) with non-existent outputDir: log file cannot be created.
func TestExecuteWorkflow_WorkflowLogCreationFailure(t *testing.T) {
	t.Parallel()

	tmpFile, err := os.CreateTemp("", "combined-log-test-*")
	require.NoError(t, err)

	t.Cleanup(func() { _ = tmpFile.Close() })

	wf := WorkflowExecution{Name: "testcoveragelog"}
	result := executeWorkflow(wf, tmpFile, "/totally/nonexistent/dir/xyz123", false, "/bin/echo", "")
	require.False(t, result.Success)
	require.NotEmpty(t, result.ErrorMessages)
}

// TestExecuteWorkflow_PipeSetupFailure tests the setupCmdPipes error path.
// Injects a mock that returns an error from setupCmdPipes.
// Not parallel: modifies package-level setupCmdPipes variable.
func TestExecuteWorkflow_PipeSetupFailure(t *testing.T) {
	varMu.Lock()

	origSetupCmdPipes := setupCmdPipes
	setupCmdPipes = func(_ *exec.Cmd) (io.ReadCloser, io.ReadCloser, error) {
		return nil, nil, fmt.Errorf("mock pipe setup error")
	}

	varMu.Unlock()

	t.Cleanup(func() {
		varMu.Lock()

		setupCmdPipes = origSetupCmdPipes

		varMu.Unlock()
	})

	tempDir := t.TempDir()
	wfDir := makeWorkflowsDir(t, "mock")
	outputDir := filepath.Join(tempDir, "output")

	result := runWithWorkflowsDir([]string{
		"-workflows=mock",
		"-act-path=/bin/echo",
		"-output=" + outputDir,
	}, wfDir)

	require.Equal(t, 1, result)
}

// TestExecuteWorkflow_WorkflowLogCloseError tests the deferred workflowLog.Close error.
// Injects a doCloseFile mock that returns an error on the first call.
// Not parallel: modifies package-level doCloseFile variable.
func TestExecuteWorkflow_WorkflowLogCloseError(t *testing.T) {
	var closeCallCount int

	varMu.Lock()

	origDoCloseFile := doCloseFile
	closeCallCount = 0
	doCloseFile = func(f *os.File) error {
		closeCallCount++
		if closeCallCount == 1 {
			// First call is for workflowLog inside executeWorkflow.
			return fmt.Errorf("mock workflow log close error")
		}
		// Subsequent calls use real close.
		return f.Close()
	}

	varMu.Unlock()

	t.Cleanup(func() {
		varMu.Lock()

		doCloseFile = origDoCloseFile

		varMu.Unlock()
	})

	tempDir := t.TempDir()
	wfDir := makeWorkflowsDir(t, "mock")
	outputDir := filepath.Join(tempDir, "output")

	// /bin/echo exits 0 so the workflow succeeds despite the close error warning.
	result := runWithWorkflowsDir([]string{
		"-workflows=mock",
		"-act-path=/bin/echo",
		"-output=" + outputDir,
	}, wfDir)

	require.Equal(t, 0, result)
}

// TestExecuteWorkflow_CombinedLogCloseError tests the deferred combinedLog.Close error.
// Exercises the doCloseFile path for combinedLog in workflow.go.
// Not parallel: modifies package-level doCloseFile variable.
func TestExecuteWorkflow_CombinedLogCloseError(t *testing.T) {
	var closeCallCount int

	varMu.Lock()

	origDoCloseFile := doCloseFile
	closeCallCount = 0
	doCloseFile = func(f *os.File) error {
		closeCallCount++
		if closeCallCount == 2 { //nolint:mnd // 2nd close call = combinedLog.
			return fmt.Errorf("mock combined log close error")
		}

		return f.Close()
	}

	varMu.Unlock()

	t.Cleanup(func() {
		varMu.Lock()

		doCloseFile = origDoCloseFile

		varMu.Unlock()
	})

	tempDir := t.TempDir()
	wfDir := makeWorkflowsDir(t, "mock")
	outputDir := filepath.Join(tempDir, "output")

	result := runWithWorkflowsDir([]string{
		"-workflows=mock",
		"-act-path=/bin/echo",
		"-output=" + outputDir,
	}, wfDir)

	// Run should succeed even though combinedLog.Close() failed.
	require.Equal(t, 0, result)
}

// TestGetWorkflowDescription_WithFile exercises the file-reading path.
// Not parallel: requires CWD change via os.Chdir.
func TestGetWorkflowDescription_WithFile(t *testing.T) {
	tempDir := t.TempDir()
	wfDir := filepath.Join(tempDir, ".github", "workflows")
	err := os.MkdirAll(wfDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
	require.NoError(t, err)

	// File with a name field.
	err = os.WriteFile(
		filepath.Join(wfDir, "ci-mytest.yml"),
		[]byte("name: My Test Workflow\non:\n  push:\n"),
		cryptoutilSharedMagic.FilePermOwnerReadWriteGroupRead,
	)
	require.NoError(t, err)

	// File without a name field (triggers the fallback in the loop body).
	err = os.WriteFile(
		filepath.Join(wfDir, "ci-noname.yml"),
		[]byte("description: no name here\non:\n  push:\n"),
		cryptoutilSharedMagic.FilePermOwnerReadWriteGroupRead,
	)
	require.NoError(t, err)

	origDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	t.Cleanup(func() { _ = os.Chdir(origDir) })

	// File with name field -> should return the name from the file.
	result := getWorkflowDescription("mytest")
	require.Equal(t, "My Test Workflow", result)

	// File without name field -> returns title-cased fallback.
	result = getWorkflowDescription("noname")
	require.True(t, strings.Contains(strings.ToLower(result), "noname"),
		"expected fallback description containing 'noname', got: %s", result)
}
