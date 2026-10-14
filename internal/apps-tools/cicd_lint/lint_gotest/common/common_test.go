// Copyright (c) 2025-2026 Justin Cranford.
package common

import (
	"testing"

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
			input:     []string{testUserTestGo, "repo_test.go"},
			wantCount: 2,
		},
		{
			name:      "exclude_cicd",
			input:     []string{testUserTestGo, "cicd_test.go"},
			wantCount: 1,
		},
		{
			name:      "exclude_testmain",
			input:     []string{testUserTestGo, "testmain_test.go"},
			wantCount: 1,
		},
		{
			name:      "exclude_e2e",
			input:     []string{testUserTestGo, "e2e_test.go"},
			wantCount: 1,
		},
		{
			name:      "exclude_sessions",
			input:     []string{testUserTestGo, "sessions_test.go"},
			wantCount: 1,
		},
		{
			name:      "exclude_admin",
			input:     []string{testUserTestGo, "admin_test.go"},
			wantCount: 1,
		},
		{
			name:      "exclude_lint_gotest_path",
			input:     []string{testUserTestGo, "lint_gotest/patterns_test.go"},
			wantCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := FilterExcludedTestFiles(tc.input)

			require.Len(t, result, tc.wantCount)
		})
	}
}
