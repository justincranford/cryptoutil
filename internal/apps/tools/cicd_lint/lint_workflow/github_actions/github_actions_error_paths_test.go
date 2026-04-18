// Copyright (c) 2025 Justin Cranford

package github_actions

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"

	"github.com/stretchr/testify/require"
)

// TestCheck_OutdatedActionsFound covers the len(outdated) > 0 branch in CheckInDir.
func TestCheck_OutdatedActionsFound(t *testing.T) {
	t.Parallel()

	stubCheckVersionsFn := func(_ *cryptoutilCmdCicdCommon.Logger, details map[string]WorkflowActionDetails, _ *WorkflowActionExceptions) ([]WorkflowActionDetails, []WorkflowActionDetails, []string) {
		var outdated []WorkflowActionDetails //nolint:prealloc // range map, size unknown

		for _, d := range details {
			d.LatestVersion = "v99.0.0"
			outdated = append(outdated, d)
		}

		return outdated, nil, nil
	}

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
	require.NoError(t, os.WriteFile(workflowFile, []byte(workflowContent), cryptoutilSharedMagic.CacheFilePermissions))

	err := checkInDir(logger, []string{workflowFile}, tmpDir, stubCheckVersionsFn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "found")
	require.Contains(t, err.Error(), "outdated GitHub Actions")
}
