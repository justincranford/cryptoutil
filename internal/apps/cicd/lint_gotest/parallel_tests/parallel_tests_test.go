// Copyright (c) 2025 Justin Cranford

package parallel_tests

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

func TestEnforceParallelTests_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger, []string{})

	require.NoError(t, err)
}

func TestEnforceParallelTests(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileContent string
		wantErr     bool
	}{
		{
			name: "valid_with_parallel",
			fileContent: `package test

import "testing"

func TestSomething(t *testing.T) {
	t.Parallel()
	t.Log("test")
}
`,
			wantErr: false,
		},
		{
			name: "missing_parallel",
			fileContent: `package test

import "testing"

func TestSomething(t *testing.T) {
	t.Log("test without parallel")
}
`,
			wantErr: true,
		},
		{
			name: "no_test_functions",
			fileContent: `package test

import "testing"

func helperFunc(t *testing.T) {
	t.Log("helper")
}
`,
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test_test.go")
			err := os.WriteFile(testFile, []byte(tc.fileContent), 0o600)
			require.NoError(t, err)

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			testFiles := []string{testFile}

			err = Check(logger, testFiles)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckParallelUsage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileContent string
		wantIssues  bool
	}{
		{
			name: "has_parallel",
			fileContent: `package test
import "testing"
func TestA(t *testing.T) { t.Parallel() }
`,
			wantIssues: false,
		},
		{
			name: "missing_parallel",
			fileContent: `package test
import "testing"
func TestA(t *testing.T) { t.Log("test") }
`,
			wantIssues: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test_test.go")
			err := os.WriteFile(testFile, []byte(tc.fileContent), 0o600)
			require.NoError(t, err)

			issues := CheckParallelUsage(testFile)

			if tc.wantIssues {
				require.NotEmpty(t, issues)
			} else {
				require.Empty(t, issues)
			}
		})
	}
}
