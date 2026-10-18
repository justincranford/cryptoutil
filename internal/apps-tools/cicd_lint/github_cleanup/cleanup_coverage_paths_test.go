// Copyright (c) 2025-2026 Justin Cranford.
package github_cleanup

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const coverageInvalidArgument = "--unknown"

func TestCleanupPublicAPIRejectsInvalidArguments(t *testing.T) {
	t.Parallel()

	err := Cleanup(cryptoutilCmdCicdCommon.NewLogger("test"), []string{coverageInvalidArgument})

	require.ErrorContains(t, err, "github-cleanup: invalid arguments")
}

func TestCleanupArtifactAndCacheDryRunFiltering(t *testing.T) {
	t.Parallel()

	recentDate := time.Now().UTC().Format(time.RFC3339)
	tests := []struct {
		name      string
		operation func(*CleanupConfig) error
		response  string
	}{
		{
			name:      "artifacts",
			operation: CleanupArtifacts,
			response: fmt.Sprintf(
				`{"total_count":3,"artifacts":[{"id":17,"size_in_bytes":32,"created_at":"invalid-date"},{"id":18,"size_in_bytes":64,"created_at":%q},{"id":19,"size_in_bytes":96,"created_at":%q}]}`,
				recentDate,
				testOldDate,
			),
		},
		{
			name:      "caches",
			operation: CleanupCaches,
			response: fmt.Sprintf(
				`{"total_count":3,"actions_caches":[{"id":27,"size_in_bytes":32,"last_accessed_at":"invalid-date"},{"id":28,"size_in_bytes":64,"last_accessed_at":%q},{"id":29,"size_in_bytes":96,"last_accessed_at":%q}]}`,
				recentDate,
				testOldDate,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var deleteCount int

			cfg := NewDefaultConfig(cryptoutilCmdCicdCommon.NewLogger("test"))
			cfg.ghExecFn = func(args ...string) ([]byte, error) {
				if strings.Contains(strings.Join(args, " "), "-X DELETE") {
					deleteCount++
				}

				return []byte(tc.response), nil
			}

			require.NoError(t, tc.operation(cfg))
			require.Zero(t, deleteCount)
		})
	}
}

func TestCleanupRunsDryRun(t *testing.T) {
	t.Parallel()

	var deleteCount int

	cfg := NewDefaultConfig(cryptoutilCmdCicdCommon.NewLogger("test"))
	cfg.KeepMinRuns = 0
	cfg.ghExecFn = func(args ...string) ([]byte, error) {
		if strings.Contains(strings.Join(args, " "), "-X DELETE") {
			deleteCount++
		}

		return []byte(testRunsResponse), nil
	}

	require.NoError(t, CleanupRuns(cfg))
	require.Zero(t, deleteCount)
}

func TestCleanupAllAggregatesOperationFailures(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig(cryptoutilCmdCicdCommon.NewLogger("test"))
	cfg.ghExecFn = func(...string) ([]byte, error) {
		return nil, errors.New("injected list failure")
	}

	err := CleanupAll(cfg)

	require.ErrorContains(t, err, "runs: failed to list workflow runs")
	require.ErrorContains(t, err, "artifacts: failed to list artifacts")
	require.ErrorContains(t, err, "caches: failed to list caches")
}

// Sequential: temporarily isolates PATH so executeGH reaches ghExec's missing-binary path.
func TestExecuteGHUsesDefaultExecutor(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	cfg := NewDefaultConfig(cryptoutilCmdCicdCommon.NewLogger("test"))
	cfg.ghExecFn = nil

	_, err := cfg.executeGH(cryptoutilSharedMagic.CLIVersionCommand)

	require.ErrorContains(t, err, "gh version failed")
}
