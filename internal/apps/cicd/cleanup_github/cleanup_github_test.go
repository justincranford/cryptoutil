// Copyright (c) 2025 Justin Cranford

package cleanup_github

import (
	"errors"
	"os/exec"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

const (
	testMaxAgeDays  = 7
	testKeepMinRuns = 5
)

func TestNewDefaultConfig(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	cfg := NewDefaultConfig(logger)

	require.NotNil(t, cfg, "Config should not be nil")
	require.Equal(t, defaultMaxAgeDays, cfg.MaxAgeDays, "Default max age days")
	require.Equal(t, defaultKeepMinRuns, cfg.KeepMinRuns, "Default keep min runs")
	require.False(t, cfg.Confirm, "Default confirm should be false (dry-run)")
	require.Empty(t, cfg.Repo, "Default repo should be empty (auto-detect)")
	require.NotNil(t, cfg.Logger, "Logger should be set")
}

func TestParseArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		wantConfirm bool
		wantMaxAge  int
		wantKeepMin int
		wantRepo    string
		wantErr     bool
	}{
		{
			name:        "no args (defaults)",
			args:        []string{},
			wantConfirm: false,
			wantMaxAge:  defaultMaxAgeDays,
			wantKeepMin: defaultKeepMinRuns,
			wantRepo:    "",
			wantErr:     false,
		},
		{
			name:        "confirm flag",
			args:        []string{"--confirm"},
			wantConfirm: true,
			wantMaxAge:  defaultMaxAgeDays,
			wantKeepMin: defaultKeepMinRuns,
			wantRepo:    "",
			wantErr:     false,
		},
		{
			name:        "max age days",
			args:        []string{"--max-age-days=7"},
			wantConfirm: false,
			wantMaxAge:  testMaxAgeDays,
			wantKeepMin: defaultKeepMinRuns,
			wantRepo:    "",
			wantErr:     false,
		},
		{
			name:        "keep min runs",
			args:        []string{"--keep-min-runs=5"},
			wantConfirm: false,
			wantMaxAge:  defaultMaxAgeDays,
			wantKeepMin: testKeepMinRuns,
			wantRepo:    "",
			wantErr:     false,
		},
		{
			name:        "repo flag",
			args:        []string{"--repo=owner/repo"},
			wantConfirm: false,
			wantMaxAge:  defaultMaxAgeDays,
			wantKeepMin: defaultKeepMinRuns,
			wantRepo:    "owner/repo",
			wantErr:     false,
		},
		{
			name:        "all flags combined",
			args:        []string{"--confirm", "--max-age-days=14", "--keep-min-runs=3", "--repo=test/repo"},
			wantConfirm: true,
			wantMaxAge:  14,
			wantKeepMin: 3,
			wantRepo:    "test/repo",
			wantErr:     false,
		},
		{
			name:    "invalid max age days (negative)",
			args:    []string{"--max-age-days=-1"},
			wantErr: true,
		},
		{
			name:    "invalid max age days (zero)",
			args:    []string{"--max-age-days=0"},
			wantErr: true,
		},
		{
			name:    "invalid max age days (non-numeric)",
			args:    []string{"--max-age-days=abc"},
			wantErr: true,
		},
		{
			name:    "invalid keep min runs (negative)",
			args:    []string{"--keep-min-runs=-1"},
			wantErr: true,
		},
		{
			name:    "unknown flag",
			args:    []string{"--unknown"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			cfg := NewDefaultConfig(logger)
			err := ParseArgs(tc.args, cfg)

			if tc.wantErr {
				require.Error(t, err, "Expected error for args: %v", tc.args)

				return
			}

			require.NoError(t, err, "Unexpected error for args: %v", tc.args)
			require.Equal(t, tc.wantConfirm, cfg.Confirm, "Confirm mismatch")
			require.Equal(t, tc.wantMaxAge, cfg.MaxAgeDays, "MaxAgeDays mismatch")
			require.Equal(t, tc.wantKeepMin, cfg.KeepMinRuns, "KeepMinRuns mismatch")
			require.Equal(t, tc.wantRepo, cfg.Repo, "Repo mismatch")
		})
	}
}

func TestParseArgs_KeepMinRunsZero(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	cfg := NewDefaultConfig(logger)
	err := ParseArgs([]string{"--keep-min-runs=0"}, cfg)

	require.NoError(t, err, "Zero keep-min-runs should be valid")
	require.Equal(t, 0, cfg.KeepMinRuns, "KeepMinRuns should be 0")
}

func TestRepoArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		repo     string
		expected []string
	}{
		{
			name:     "empty repo (auto-detect)",
			repo:     "",
			expected: nil,
		},
		{
			name:     "explicit repo",
			repo:     "owner/repo",
			expected: []string{"--repo", "owner/repo"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			cfg := NewDefaultConfig(logger)
			cfg.Repo = tc.repo

			result := repoArgs(cfg)
			require.Equal(t, tc.expected, result, "repoArgs mismatch")
		})
	}
}

func TestIsExitError_NonExitError(t *testing.T) {
	t.Parallel()

	var target *exec.ExitError

	simpleErr := errors.New("simple error")
	result := isExitError(simpleErr, &target)
	require.False(t, result, "Non-ExitError should return false")
	require.Nil(t, target, "Target should remain nil for non-ExitError")
}
