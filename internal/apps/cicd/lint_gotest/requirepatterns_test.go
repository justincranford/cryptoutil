// Copyright (c) 2025 Justin Cranford

package lint_gotest

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

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

			err = enforceRequireOverAssert(logger, testFiles)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
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

			err = enforceParallelTests(logger, testFiles)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEnforceHardcodedPasswords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileContent string
		wantErr     bool
	}{
		{
			name: "valid_dynamic_password",
			fileContent: `package test

import (
	"testing"
	googleUuid "github.com/google/uuid"
)

func TestSomething(t *testing.T) {
	t.Parallel()
	password := googleUuid.NewV7().String()
	_ = password
}
`,
			wantErr: false,
		},
		{
			name: "invalid_hardcoded_password",
			fileContent: `package test

import "testing"

func TestSomething(t *testing.T) {
	t.Parallel()
	password := "test123"
	_ = password
}
`,
			wantErr: true,
		},
		{
			name: "invalid_hardcoded_secret",
			fileContent: `package test

import "testing"

func TestSomething(t *testing.T) {
	t.Parallel()
	secret := "secret"
	_ = secret
}
`,
			wantErr: true,
		},
		{
			name: "no_passwords",
			fileContent: `package test

import "testing"

func TestSomething(t *testing.T) {
	t.Parallel()
	value := "some value"
	_ = value
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

			err = enforceHardcodedPasswords(logger, testFiles)

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

			issues := checkAssertUsage(testFile)

			if tc.wantIssues {
				require.NotEmpty(t, issues)
			} else {
				require.Empty(t, issues)
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

			issues := checkParallelUsage(testFile)

			if tc.wantIssues {
				require.NotEmpty(t, issues)
			} else {
				require.Empty(t, issues)
			}
		})
	}
}

func TestCheckHardcodedPasswords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileContent string
		wantIssues  bool
	}{
		{
			name:        "hardcoded_password",
			fileContent: `password := "test123"`,
			wantIssues:  true,
		},
		{
			name:        "hardcoded_password_alt",
			fileContent: `password := "password"`,
			wantIssues:  true,
		},
		{
			name:        "hardcoded_secret",
			fileContent: `secret := "secret"`,
			wantIssues:  true,
		},
		{
			name:        "no_hardcoded",
			fileContent: `value := "other value"`,
			wantIssues:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test_test.go")
			err := os.WriteFile(testFile, []byte(tc.fileContent), 0o600)
			require.NoError(t, err)

			issues := checkHardcodedPasswords(testFile)

			if tc.wantIssues {
				require.NotEmpty(t, issues)
			} else {
				require.Empty(t, issues)
			}
		})
	}
}

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

			result := filterExcludedTestFiles(tc.input)

			require.Len(t, result, tc.wantCount)
		})
	}
}

func TestEnforceRequireOverAssert_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := enforceRequireOverAssert(logger, []string{})

	require.NoError(t, err)
}

func TestEnforceParallelTests_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := enforceParallelTests(logger, []string{})

	require.NoError(t, err)
}

func TestEnforceHardcodedPasswords_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := enforceHardcodedPasswords(logger, []string{})

	require.NoError(t, err)
}
