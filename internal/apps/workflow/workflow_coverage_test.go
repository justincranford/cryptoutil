// Copyright (c) 2025 Justin Cranford
//
//

package workflow

import (
"os"
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
wfDir := filepath.Join(tempDir, ".github", "workflows")
err := os.MkdirAll(wfDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
require.NoError(t, err)

names := workflowNames
if len(names) == 0 {
names = []string{"mock"}
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

result := runWithWorkflowsDir([]string{"-workflows=mock"}, "/nonexistent/workflows/dir")
require.Equal(t, 1, result)
}

// TestRunWithWorkflowsDir_HelpFlag tests the --help flag path.
func TestRunWithWorkflowsDir_HelpFlag(t *testing.T) {
t.Parallel()

wfDir := makeWorkflowsDir(t, "quality", "coverage")
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
wfDir := makeWorkflowsDir(t, "mock")
outputDir := filepath.Join(tempDir, "output")

result := runWithWorkflowsDir([]string{
"-workflows=mock",
"-dry-run",
"-output=" + outputDir,
}, wfDir)

require.Equal(t, 0, result)
}

// TestRunWithWorkflowsDir_DryRun_DastWorkflow tests dry-run with the dast workflow (special event).
func TestRunWithWorkflowsDir_DryRun_DastWorkflow(t *testing.T) {
t.Parallel()

tempDir := t.TempDir()
wfDir := makeWorkflowsDir(t, "dast")
outputDir := filepath.Join(tempDir, "output")

result := runWithWorkflowsDir([]string{
"-workflows=dast",
"-dry-run",
"-output=" + outputDir,
}, wfDir)

require.Equal(t, 0, result)
}

// TestRunWithWorkflowsDir_DryRun_WithActArgs tests dry-run with extra act arguments.
func TestRunWithWorkflowsDir_DryRun_WithActArgs(t *testing.T) {
t.Parallel()

tempDir := t.TempDir()
wfDir := makeWorkflowsDir(t, "quality")
outputDir := filepath.Join(tempDir, "output")

result := runWithWorkflowsDir([]string{
"-workflows=quality",
"-dry-run",
"-output=" + outputDir,
"-act-args=--verbose",
}, wfDir)

require.Equal(t, 0, result)
}

// TestRunWithWorkflowsDir_NonDryRun_EchoAct tests non-dry-run mode using echo as a fake act binary.
func TestRunWithWorkflowsDir_NonDryRun_EchoAct(t *testing.T) {
t.Parallel()

tempDir := t.TempDir()
wfDir := makeWorkflowsDir(t, "mock")
outputDir := filepath.Join(tempDir, "output")

// /bin/echo just prints args and exits 0 - no real act needed.
result := runWithWorkflowsDir([]string{
"-workflows=mock",
"-act-path=/bin/echo",
"-output=" + outputDir,
}, wfDir)

// /bin/echo always exits 0.
require.Equal(t, 0, result)
}

// TestRunWithWorkflowsDir_NonDryRun_ActNotFound tests non-dry-run when act binary doesn't exist.
func TestRunWithWorkflowsDir_NonDryRun_ActNotFound(t *testing.T) {
t.Parallel()

tempDir := t.TempDir()
wfDir := makeWorkflowsDir(t, "mock")
outputDir := filepath.Join(tempDir, "output")

result := runWithWorkflowsDir([]string{
"-workflows=mock",
"-act-path=/nonexistent/bin/fake-act-xyz",
"-output=" + outputDir,
}, wfDir)

// Should fail since act binary doesn't exist.
require.Equal(t, 1, result)
}

// TestRunWithWorkflowsDir_OutputDirCreationFails tests when output dir creation fails.
func TestRunWithWorkflowsDir_OutputDirCreationFails(t *testing.T) {
t.Parallel()

wfDir := makeWorkflowsDir(t, "mock")

// Use an unwritable path to force MkdirAll failure.
result := runWithWorkflowsDir([]string{
"-workflows=mock",
"-dry-run",
"-output=/root/cannot_write_here_xyz/output",
}, wfDir)

// Should fail since we can't create the output dir.
require.Equal(t, 1, result)
}

// TestWorkflow_EntryPoint tests the Workflow() public entry point.
func TestWorkflow_EntryPoint(t *testing.T) {
t.Parallel()

// Workflow requires args[0] as program name.
// Test the error path (invalid workflows dir) via the Workflow function itself.
// Since Workflow calls run() which uses ".github/workflows" directly,
// and test CWD is the package dir (no .github/workflows there),
// this will exercise the error path.
result := Workflow([]string{"workflow", "-workflows=nonexistent"}, nil, nil, nil)
// Either 1 (no .github/workflows from package dir) or 1 (no valid workflow).
require.Equal(t, 1, result)
}

// TestPrintHelp tests the printHelp function.
func TestPrintHelp(t *testing.T) {
t.Parallel()

workflows := map[string]WorkflowConfig{
"quality":  {Description: "Quality checks"},
"coverage": {Description: "Coverage collection"},
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
"quality":  {Description: "Quality checks"},
"coverage": {Description: "Coverage"},
"dast":     {Description: "DAST"},
}

tests := []struct {
name          string
input         string
expectedNames []string
}{
{
name:          "single valid workflow",
input:         "quality",
expectedNames: []string{"quality"},
},
{
name:          "multiple valid workflows",
input:         "quality,coverage",
expectedNames: []string{"quality", "coverage"},
},
{
name:          "mixed valid and invalid",
input:         "quality,nonexistent,coverage",
expectedNames: []string{"quality", "coverage"},
},
{
name:          "all invalid",
input:         "invalid1,invalid2",
expectedNames: []string{},
},
{
name:          "with whitespace",
input:         "quality, coverage",
expectedNames: []string{"quality", "coverage"},
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
