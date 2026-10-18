// Copyright (c) 2025-2026 Justin Cranford.
package precommit_cicd_architecture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	testGoExecutable                  = "go"
	testCICDLintMainPath              = "cmd/cicd-lint/main.go"
	testLintTextCommand               = "lint-text"
	testFormatTextCommand             = "format-text"
	testUnsupportedStageError         = "unsupported stage"
	testExpectedLintOnlyBulkHookError = "expected exactly one lint-only bulk cicd hook"
)

func TestCheckAndLoadConfigErrors(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("precommit-cicd-architecture")
	require.ErrorContains(t, Check(logger), "precommit-cicd-architecture")

	rootDir := t.TempDir()
	require.ErrorContains(t, CheckInDir(logger, rootDir), "cannot read")

	configPath := filepath.Join(rootDir, preCommitConfigPath)
	require.NoError(t, os.WriteFile(configPath, []byte("repos: [broken"), cryptoutilSharedMagic.CacheFilePermissions))
	require.ErrorContains(t, CheckInDir(logger, rootDir), "cannot parse")
}

func TestExtractAndClassifyCommands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		commands []string
		kind     hookKind
	}{
		{name: "empty args"},
		{name: "missing binary", args: []string{testGoExecutable, cryptoutilSharedMagic.WorkflowSubCmdRun}},
		{name: "binary without commands", args: []string{testGoExecutable, cryptoutilSharedMagic.WorkflowSubCmdRun, testCICDLintMainPath}},
		{
			name:     "filters flags and invalid commands",
			args:     []string{testGoExecutable, cryptoutilSharedMagic.WorkflowSubCmdRun, testCICDLintMainPath, cryptoutilSharedMagic.FlagQuiet, testLintTextCommand, "not-a-command"},
			commands: []string{testLintTextCommand},
			kind:     hookKindLint,
		},
		{
			name:     "mixed lint and format commands",
			args:     []string{testGoExecutable, cryptoutilSharedMagic.WorkflowSubCmdRun, "./" + testCICDLintMainPath, testLintTextCommand, testFormatTextCommand},
			commands: []string{testLintTextCommand, testFormatTextCommand},
			kind:     hookKindMixed,
		},
		{
			name:     "format commands",
			args:     []string{testGoExecutable, cryptoutilSharedMagic.WorkflowSubCmdRun, testCICDLintMainPath, testFormatTextCommand},
			commands: []string{testFormatTextCommand},
			kind:     hookKindFormat,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			commands := extractCICDCommands(tc.args)
			require.Equal(t, tc.commands, commands)
			require.Equal(t, tc.kind, classifyCommands(commands))
		})
	}

	require.Equal(t, hookKindUnknown, classifyCommands([]string{"unclassified"}))
}

func TestDeriveStage(t *testing.T) {
	t.Parallel()

	require.Equal(t, stagePreCommit, deriveStage(nil))
	require.Equal(t, stagePrePush, deriveStage([]string{stagePrePush, stagePreCommit}))
}

func TestCollectBulkHooksSkipsNonCICDHooks(t *testing.T) {
	t.Parallel()

	cfg := &preCommitConfig{
		Repos: []preCommitRepo{{
			Hooks: []preCommitHook{
				{ID: "no-args"},
				{ID: "not-cicd", Args: []string{testGoExecutable, cryptoutilSharedMagic.WorkflowSubCmdRun}},
				{ID: "empty-commands", Args: []string{testGoExecutable, cryptoutilSharedMagic.WorkflowSubCmdRun, testCICDLintMainPath, cryptoutilSharedMagic.FlagQuiet}},
				{ID: "lint-hook", Args: []string{testGoExecutable, cryptoutilSharedMagic.WorkflowSubCmdRun, testCICDLintMainPath, testLintTextCommand}},
			},
		}},
	}

	hooks := collectBulkHooks(cfg)

	require.Len(t, hooks, 1)
	require.Equal(t, "lint-hook", hooks[0].ID)
	require.Equal(t, stagePreCommit, hooks[0].Stage)
}

func TestValidateBulkHooksPolicies(t *testing.T) {
	t.Parallel()

	buildValid := func() []bulkHookInfo {
		lintCommands := expectedCommandsForPrefix("lint-")
		formatCommands := expectedCommandsForPrefix("format-")

		return []bulkHookInfo{
			{ID: "pre-commit-lint", Stage: stagePreCommit, Commands: lintCommands, kind: hookKindLint},
			{ID: "pre-commit-format", Stage: stagePreCommit, RequireSerial: true, Commands: formatCommands, kind: hookKindFormat},
			{ID: "pre-push-lint", Stage: stagePrePush, Commands: lintCommands, kind: hookKindLint},
			{ID: "pre-push-format", Stage: stagePrePush, RequireSerial: true, Commands: formatCommands, kind: hookKindFormat},
		}
	}

	require.Empty(t, validateBulkHooks(buildValid()))

	tests := []struct {
		name  string
		hooks []bulkHookInfo
		want  string
	}{
		{
			name: testUnsupportedStageError,
			hooks: []bulkHookInfo{
				{ID: "wrong-stage", Stage: "manual", Commands: []string{testLintTextCommand}, kind: hookKindLint},
			},
			want: testUnsupportedStageError,
		},
		{
			name: "lint hook must run concurrently",
			hooks: []bulkHookInfo{
				{ID: "serial-lint", Stage: stagePreCommit, RequireSerial: true, Commands: []string{testLintTextCommand}, kind: hookKindLint},
			},
			want: "lint-only but require_serial=true",
		},
		{
			name: "format hook must run serially",
			hooks: []bulkHookInfo{
				{ID: "parallel-format", Stage: stagePreCommit, Commands: []string{testFormatTextCommand}, kind: hookKindFormat},
			},
			want: "format-only but require_serial=false",
		},
		{
			name: "duplicate stage hook",
			hooks: []bulkHookInfo{
				{ID: "lint-one", Stage: stagePreCommit, Commands: []string{testLintTextCommand}, kind: hookKindLint},
				{ID: "lint-two", Stage: stagePreCommit, Commands: []string{testLintTextCommand}, kind: hookKindLint},
			},
			want: testExpectedLintOnlyBulkHookError,
		},
		{
			name: "unknown kind is ignored",
			hooks: []bulkHookInfo{
				{ID: cryptoutilSharedMagic.UNKNOWN, Stage: stagePreCommit, Commands: []string{"misc"}, kind: hookKindUnknown},
			},
			want: testExpectedLintOnlyBulkHookError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			violations := validateBulkHooks(tc.hooks)
			require.NotEmpty(t, violations)
			require.Contains(t, strings.Join(violations, "\n"), tc.want)
		})
	}
}
