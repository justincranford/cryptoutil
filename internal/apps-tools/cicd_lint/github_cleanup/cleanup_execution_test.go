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

const (
	testOldDate                 = "2000-01-01T00:00:00Z"
	testRunsResponse            = `{"total_count":1,"workflow_runs":[{"id":1,"status":"completed","conclusion":"failure","created_at":"2000-01-01T00:00:00Z","name":"test","workflow_id":2}]}`
	testArtifactsBody           = `{"total_count":1,"artifacts":[{"id":1,"name":"test","size_in_bytes":10,"created_at":"2000-01-01T00:00:00Z","expired":false}]}`
	testCachesBody              = `{"total_count":1,"actions_caches":[{"id":1,"key":"test","ref":"main","size_in_bytes":10,"last_accessed_at":"2000-01-01T00:00:00Z","created_at":"2000-01-01T00:00:00Z"}]}`
	testInvalidArguments        = "invalid arguments"
	testWorkflowRunsListFailure = "failed to list workflow runs"
	testInvalidJSON             = "{"
)

func newExecutionTestConfig(ghExecFn func(...string) ([]byte, error)) *CleanupConfig {
	cfg := NewDefaultConfig(cryptoutilCmdCicdCommon.NewLogger("test"))
	cfg.ghExecFn = ghExecFn

	return cfg
}

func TestCleanupWithCleaners(t *testing.T) {
	t.Parallel()

	cleanerError := errors.New("injected cleaner failure")
	tests := []struct {
		name    string
		args    []string
		cleaner func(*CleanupConfig) error
		wantErr string
	}{
		{
			name: "runs cleaners and aggregates errors",
			cleaner: func(*CleanupConfig) error {
				return cleanerError
			},
			wantErr: "cleanup-one: injected cleaner failure",
		},
		{
			name:    "rejects invalid arguments",
			args:    []string{coverageInvalidArgument},
			cleaner: func(*CleanupConfig) error { return nil },
			wantErr: testInvalidArguments,
		},
		{
			name:    "successful cleaner",
			cleaner: func(*CleanupConfig) error { return nil },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			called := 0
			cleaners := []struct {
				name    string
				cleaner CleanerFunc
			}{
				{name: "cleanup-one", cleaner: func(cfg *CleanupConfig) error {
					called++

					return tc.cleaner(cfg)
				}},
				{name: "cleanup-two", cleaner: func(*CleanupConfig) error {
					called++

					return nil
				}},
			}

			err := cleanupWithCleaners(cryptoutilCmdCicdCommon.NewLogger("test"), tc.args, cleaners)
			if tc.wantErr == "" {
				require.NoError(t, err)
				require.Equal(t, 2, called)

				return
			}

			require.ErrorContains(t, err, tc.wantErr)

			if tc.args == nil {
				require.Equal(t, 2, called)
			} else {
				require.Zero(t, called)
			}
		})
	}
}

func TestMainWithConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		subCommand string
		args       []string
		wantCode   int
		wantText   string
	}{
		{name: "runs cleanup runs", subCommand: cleanupSubcommandRuns},
		{name: "runs cleanup artifacts", subCommand: cleanupSubcommandArtifacts},
		{name: "runs cleanup caches", subCommand: cleanupSubcommandCaches},
		{name: "runs all cleanup operations", subCommand: cleanupSubcommandAll},
		{name: "unknown command", subCommand: cryptoutilSharedMagic.STRATEGY_UNKNOWN, wantCode: 1, wantText: "Unknown"},
		{name: testInvalidArguments, subCommand: cleanupSubcommandRuns, args: []string{coverageInvalidArgument}, wantCode: 1, wantText: unknownFlagPrefix},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var calls int

			cfg := newExecutionTestConfig(func(args ...string) ([]byte, error) {
				calls++

				joinedArgs := strings.Join(args, " ")
				switch {
				case strings.Contains(joinedArgs, "/actions/runs"):
					return []byte(`{"total_count":0,"workflow_runs":[]}`), nil
				case strings.Contains(joinedArgs, "/actions/artifacts"):
					return []byte(`{"total_count":0,"artifacts":[]}`), nil
				case strings.Contains(joinedArgs, "/actions/cache/usage"):
					return []byte(`{"total_count":0,"actions_caches":[]}`), nil
				default:
					return nil, fmt.Errorf("unexpected gh arguments: %s", joinedArgs)
				}
			})

			var stderr strings.Builder

			code := mainWithConfig(tc.subCommand, tc.args, &stderr, cfg)
			require.Equal(t, tc.wantCode, code)

			if tc.wantText != "" {
				require.Contains(t, stderr.String(), tc.wantText)
			}

			if tc.wantCode == 0 {
				expectedCalls := 1
				if tc.subCommand == cleanupSubcommandAll {
					expectedCalls = len([]string{cleanupSubcommandRuns, cleanupSubcommandArtifacts, cleanupSubcommandCaches})
				}

				require.Equal(t, expectedCalls, calls)
			} else {
				require.Zero(t, calls)
			}
		})
	}
}

func TestCleanupOperations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		operation    func(*CleanupConfig) error
		listNeedle   string
		response     string
		deleteNeedle string
	}{
		{
			name: "runs deletion success",
			operation: func(cfg *CleanupConfig) error {
				cfg.Confirm = true
				cfg.KeepMinRuns = 0

				return CleanupRuns(cfg)
			},
			listNeedle:   "/actions/runs",
			response:     testRunsResponse,
			deleteNeedle: "/actions/runs/1",
		},
		{
			name: "artifact deletion success",
			operation: func(cfg *CleanupConfig) error {
				cfg.Confirm = true

				return CleanupArtifacts(cfg)
			},
			listNeedle:   "/actions/artifacts",
			response:     testArtifactsBody,
			deleteNeedle: "/actions/artifacts/1",
		},
		{
			name: "cache deletion success",
			operation: func(cfg *CleanupConfig) error {
				cfg.Confirm = true

				return CleanupCaches(cfg)
			},
			listNeedle:   "/actions/cache/usage",
			response:     testCachesBody,
			deleteNeedle: "/actions/caches/1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var deleted bool

			cfg := newExecutionTestConfig(func(args ...string) ([]byte, error) {
				command := strings.Join(args, " ")
				if strings.Contains(command, "-X DELETE") {
					deleted = true

					return nil, nil
				}

				require.Contains(t, command, tc.listNeedle)

				return []byte(tc.response), nil
			})

			require.NoError(t, tc.operation(cfg))
			require.True(t, deleted)
		})
	}
}

func TestCleanupRunFiltering(t *testing.T) {
	t.Parallel()

	oldDate := testOldDate
	recentDate := time.Now().UTC().AddDate(0, 0, -1).Format(time.RFC3339)
	body := fmt.Sprintf(
		`{"total_count":4,"workflow_runs":[{"id":1,"conclusion":"success","created_at":%q,"name":"workflow"},{"id":2,"conclusion":"success","created_at":%q,"name":"workflow"},{"id":3,"conclusion":"failure","created_at":%q,"name":"workflow"},{"id":4,"conclusion":"failure","created_at":"bad-date","name":"workflow"}]}`,
		oldDate, oldDate, recentDate,
	)

	var deleteIDs []string

	cfg := newExecutionTestConfig(func(args ...string) ([]byte, error) {
		command := strings.Join(args, " ")
		if strings.Contains(command, "-X DELETE") {
			deleteIDs = append(deleteIDs, command)

			return nil, nil
		}

		return []byte(body), nil
	})
	cfg.Confirm = true
	cfg.KeepMinRuns = 1

	require.NoError(t, CleanupRuns(cfg))
	require.Len(t, deleteIDs, 1)
	require.Contains(t, deleteIDs[0], "/actions/runs/2")
}

func TestCleanupOperationFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		operation func(*CleanupConfig) error
		response  string
		wantErr   string
	}{
		{
			name: "runs list command failure",
			operation: func(cfg *CleanupConfig) error {
				return CleanupRuns(cfg)
			},
			wantErr: testWorkflowRunsListFailure,
		},
		{
			name: "runs response parse failure",
			operation: func(cfg *CleanupConfig) error {
				return CleanupRuns(cfg)
			},
			response: testInvalidJSON,
			wantErr:  testWorkflowRunsListFailure,
		},
		{
			name: "artifact response parse failure",
			operation: func(cfg *CleanupConfig) error {
				return CleanupArtifacts(cfg)
			},
			response: testInvalidJSON,
			wantErr:  "failed to list artifacts",
		},
		{
			name: "cache response parse failure",
			operation: func(cfg *CleanupConfig) error {
				return CleanupCaches(cfg)
			},
			response: testInvalidJSON,
			wantErr:  "failed to list caches",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := newExecutionTestConfig(func(...string) ([]byte, error) {
				if tc.response == "" {
					return nil, errors.New("injected gh failure")
				}

				return []byte(tc.response), nil
			})

			require.ErrorContains(t, tc.operation(cfg), tc.wantErr)
		})
	}
}

func TestCleanupDeletionFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		operation func(*CleanupConfig) error
		response  string
	}{
		{
			name: "run delete failure",
			operation: func(cfg *CleanupConfig) error {
				cfg.Confirm = true
				cfg.KeepMinRuns = 0

				return CleanupRuns(cfg)
			},
			response: testRunsResponse,
		},
		{
			name: "artifact delete failure",
			operation: func(cfg *CleanupConfig) error {
				cfg.Confirm = true

				return CleanupArtifacts(cfg)
			},
			response: testArtifactsBody,
		},
		{
			name: "cache delete failure",
			operation: func(cfg *CleanupConfig) error {
				cfg.Confirm = true

				return CleanupCaches(cfg)
			},
			response: testCachesBody,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := newExecutionTestConfig(func(args ...string) ([]byte, error) {
				if strings.Contains(strings.Join(args, " "), "-X DELETE") {
					return nil, errors.New("delete failed")
				}

				return []byte(tc.response), nil
			})

			require.Error(t, tc.operation(cfg))
		})
	}
}

func TestMainReportsCleanupFailure(t *testing.T) {
	t.Parallel()

	cfg := newExecutionTestConfig(func(...string) ([]byte, error) {
		return nil, errors.New("list failed")
	})

	var stderr strings.Builder

	require.Equal(t, 1, mainWithConfig(cleanupSubcommandRuns, nil, &stderr, cfg))
	require.Contains(t, stderr.String(), "Error:")
}

// Sequential: overrides PATH to force a missing gh executable.
func TestGhExecMissingExecutable(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	_, err := ghExec(cryptoutilSharedMagic.CLIVersionCommand)
	require.ErrorContains(t, err, "gh version failed")
}

func TestMainEntryPoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		subCommand string
		args       []string
		wantCode   int
	}{
		{name: "unknown subcommand", subCommand: cryptoutilSharedMagic.STRATEGY_UNKNOWN, wantCode: 1},
		{name: "invalid options", subCommand: cleanupSubcommandRuns, args: []string{coverageInvalidArgument}, wantCode: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stderr strings.Builder
			require.Equal(t, tc.wantCode, Main(tc.subCommand, tc.args, &stderr))
			require.NotEmpty(t, stderr.String())
		})
	}
}
