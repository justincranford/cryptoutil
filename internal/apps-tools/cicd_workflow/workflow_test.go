// Copyright (c) 2025-2026 Justin Cranford.
//
//

package cicd_workflow

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestGetAvailableWorkflows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		createDir   bool
		files       []string
		wantNames   []string
		wantErrText string
	}{
		{
			name:      "filters workflow files and ignores directories",
			createDir: true,
			files: []string{
				"ci-quality.yml",
				"ci-coverage.yml",
				"ci-dast.yml",
				"ci-load.yml",
				"ci-race.yml",
				"not-ci-file.yml",
				"ci-test.txt",
			},
			wantNames: []string{
				workflowTestNameQuality,
				workflowTestNameCoverage,
				cryptoutilSharedMagic.WorkflowNameDAST,
				cryptoutilSharedMagic.WorkflowNameLoad,
				"race",
			},
		},
		{name: "missing workflow directory", wantErrText: "failed to read workflows directory"},
		{name: "empty workflow directory", createDir: true, wantErrText: "no workflows found"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			workflowsDir := filepath.Join(t.TempDir(), "workflows")
			if tc.createDir {
				err := os.MkdirAll(workflowsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
				require.NoError(t, err)
			}

			for _, filename := range tc.files {
				err := os.WriteFile(
					filepath.Join(workflowsDir, filename),
					[]byte("name: Test Workflow"),
					cryptoutilSharedMagic.FilePermOwnerReadWriteOnly,
				)
				require.NoError(t, err)
			}

			if len(tc.files) > 0 {
				err := os.MkdirAll(filepath.Join(workflowsDir, "subdir"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
				require.NoError(t, err)
			}

			workflows, err := getAvailableWorkflows(workflowsDir, os.ReadFile)
			if tc.wantErrText != "" {
				require.ErrorContains(t, err, tc.wantErrText)
				require.Nil(t, workflows)

				return
			}

			require.NoError(t, err)
			require.Len(t, workflows, len(tc.wantNames))

			for _, workflowName := range tc.wantNames {
				workflow, ok := workflows[workflowName]
				require.True(t, ok)
				require.Equal(t, "Test Workflow", workflow.Description)
			}

			require.NotContains(t, workflows, "not-ci-file")
			require.NotContains(t, workflows, "test")
		})
	}
}
