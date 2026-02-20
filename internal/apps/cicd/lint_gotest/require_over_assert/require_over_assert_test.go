// Copyright (c) 2025 Justin Cranford

package require_over_assert

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

func TestEnforceRequireOverAssert_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger, []string{})

	require.NoError(t, err)
}

func TestEnforceRequireOverAssert(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileContent string
		wantErr     bool
	}{
		{
			name: "valid_require_usage",
			fileContent: `package test

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func TestSomething(t *testing.T) {
	t.Parallel()
	require.NoError(t, nil)
}
`,
			wantErr: false,
		},
		{
			name: "invalid_assert_usage",
			fileContent: `package test

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
	t.Parallel()
	assert.NoError(t, nil)
}
`,
			wantErr: true,
		},
		{
			name: "assert_import_only",
			fileContent: `package test

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
	// No assert calls but assert import without require import
	t.Log("test")
}
`,
			wantErr: true,
		},
		{
			name: "both_imports_acceptable",
			fileContent: `package test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSomething(t *testing.T) {
	t.Parallel()
	// Both imported but not using assert
	require.NoError(t, nil)
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

func TestCheckAssertUsage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileContent string
		wantIssues  bool
	}{
		{
			name: "assert_noerror",
			fileContent: `package test
import "github.com/stretchr/testify/assert"
func Test(t *testing.T) { assert.NoError(t, nil) }
`,
			wantIssues: true,
		},
		{
			name: "assert_equal",
			fileContent: `package test
import "github.com/stretchr/testify/assert"
func Test(t *testing.T) { assert.Equal(t, 1, 1) }
`,
			wantIssues: true,
		},
		{
			name: "require_noerror",
			fileContent: `package test
import "github.com/stretchr/testify/require"
func Test(t *testing.T) { require.NoError(t, nil) }
`,
			wantIssues: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test_test.go")
			err := os.WriteFile(testFile, []byte(tc.fileContent), 0o600)
			require.NoError(t, err)

			issues := CheckAssertUsage(testFile)

			if tc.wantIssues {
				require.NotEmpty(t, issues)
			} else {
				require.Empty(t, issues)
			}
		})
	}
}
