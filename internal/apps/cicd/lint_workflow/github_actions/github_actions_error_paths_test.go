// Copyright (c) 2025 Justin Cranford

package github_actions

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

// TestCheck_OutdatedActionsFound covers the len(outdated) > 0 branch in Check.
// NOT parallel — modifies package-level injectable var.
func TestCheck_OutdatedActionsFound(t *testing.T) {
	original := checkActionVersionsConcurrentlyFn
	checkActionVersionsConcurrentlyFn = func(_ *cryptoutilCmdCicdCommon.Logger, details map[string]WorkflowActionDetails, _ *WorkflowActionExceptions) ([]WorkflowActionDetails, []WorkflowActionDetails, []string) {
		var outdated []WorkflowActionDetails
		for _, d := range details {
			d.LatestVersion = "v99.0.0"
			outdated = append(outdated, d)
		}

		return outdated, nil, nil
	}

	defer func() { checkActionVersionsConcurrentlyFn = original }()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Create a minimal workflow file with at least one action.
	tmpDir := t.TempDir()
	workflowContent := `name: test
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`

	workflowFile := filepath.Join(tmpDir, "test.yml")
	require.NoError(t, os.WriteFile(workflowFile, []byte(workflowContent), 0o600))

	err := Check(logger, []string{workflowFile})
	require.Error(t, err)
	require.Contains(t, err.Error(), "found")
	require.Contains(t, err.Error(), "outdated GitHub Actions")
}
