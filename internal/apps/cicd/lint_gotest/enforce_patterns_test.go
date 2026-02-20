// Copyright (c) 2025 Justin Cranford

package lint_gotest

import (
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintGoTestCommon "cryptoutil/internal/apps/cicd/lint_gotest/common"
	lintGoTestNoHardcodedPasswords "cryptoutil/internal/apps/cicd/lint_gotest/no_hardcoded_passwords"
	lintGoTestParallelTests "cryptoutil/internal/apps/cicd/lint_gotest/parallel_tests"
	lintGoTestRequireOverAssert "cryptoutil/internal/apps/cicd/lint_gotest/require_over_assert"

	"github.com/stretchr/testify/require"
)

func TestFilterExcludedTestFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     []string
		wantCount int
	}{
		{
			name:      "no_exclusions",
			input:     []string{"user_test.go", "repo_test.go"},
			wantCount: 2,
		},
		{
			name:      "exclude_cicd",
			input:     []string{"user_test.go", "cicd_test.go"},
			wantCount: 1,
		},
		{
			name:      "exclude_testmain",
			input:     []string{"user_test.go", "testmain_test.go"},
			wantCount: 1,
		},
		{
			name:      "exclude_e2e",
			input:     []string{"user_test.go", "e2e_test.go"},
			wantCount: 1,
		},
		{
			name:      "exclude_sessions",
			input:     []string{"user_test.go", "sessions_test.go"},
			wantCount: 1,
		},
		{
			name:      "exclude_admin",
			input:     []string{"user_test.go", "admin_test.go"},
			wantCount: 1,
		},
		{
			name:      "exclude_lint_gotest_path",
			input:     []string{"user_test.go", "lint_gotest/patterns_test.go"},
			wantCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := lintGoTestCommon.FilterExcludedTestFiles(tc.input)

			require.Len(t, result, tc.wantCount)
		})
	}
}

func TestEnforceRequireOverAssert_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := lintGoTestRequireOverAssert.Check(logger, []string{})

	require.NoError(t, err)
}

func TestEnforceParallelTests_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := lintGoTestParallelTests.Check(logger, []string{})

	require.NoError(t, err)
}

func TestEnforceHardcodedPasswords_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := lintGoTestNoHardcodedPasswords.Check(logger, []string{})

	require.NoError(t, err)
}
