// Copyright (c) 2025-2026 Justin Cranford.
package parallel_tests

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"

	"github.com/stretchr/testify/require"
)

func TestEnforceParallelTests_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckFiles(logger, []string{})

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
			name: testMissingParallel,
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
			err := os.WriteFile(testFile, []byte(tc.fileContent), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			testFiles := []string{testFile}

			err = CheckFiles(logger, testFiles)

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
			name: testMissingParallel,
			fileContent: `package test
import "testing"
func TestA(t *testing.T) { t.Log("test") }
`,
			wantIssues: true,
		},
		{
			name: "func_in_string_literal_not_flagged",
			fileContent: `package test
import "testing"
func TestA(t *testing.T) {
	t.Parallel()
	content := "func TestInsideString(t *testing.T) {}"
	_ = content
}
`,
			wantIssues: false,
		},
		{
			name: "nested_subtest_parallel_does_not_cover_parent",
			fileContent: `package test
import "testing"
func TestParent(t *testing.T) {
	t.Run("child", func(t *testing.T) { t.Parallel() })
}
`,
			wantIssues: true,
		},
		{
			name: "sequential_comment_exempts_function",
			fileContent: `package test
import "testing"
// Sequential: modifies package-level seam variable.
func TestB(t *testing.T) { t.Log("seam test") }
`,
			wantIssues: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test_test.go")
			err := os.WriteFile(testFile, []byte(tc.fileContent), cryptoutilSharedMagic.CacheFilePermissions)
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

func TestCheckParallelUsage_ReadFileError(t *testing.T) {
	t.Parallel()

	issues := CheckParallelUsage("/nonexistent/path/that/does/not/exist_test.go")
	require.NotEmpty(t, issues)
	require.Contains(t, issues[0], "Error reading file")
}

func TestCheckParallelUsage_ParseError(t *testing.T) {
	t.Parallel()

	testFile := filepath.Join(t.TempDir(), "invalid_test.go")
	require.NoError(t, os.WriteFile(testFile, []byte("package broken\nfunc Test( {\n"), cryptoutilSharedMagic.CacheFilePermissions))

	issues := CheckParallelUsage(testFile)

	require.Len(t, issues, 1)
	require.Contains(t, issues[0], "Error parsing file")
}

func TestTestingTParameter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
		want   string
	}{
		{name: "no parameters", source: "package p\nfunc TestNoParams() {}"},
		{name: "non-pointer", source: "package p\nfunc TestValue(t testing.T) {}"},
		{name: "wrong package", source: "package p\nfunc TestOther(t *other.T) {}"},
		{name: "unnamed parameter", source: "package p\nfunc TestUnnamed(*testing.T) {}"},
		{name: "testing parameter alias", source: "package p\nfunc TestNamed(check *testing.T) {}", want: "check"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file, err := parser.ParseFile(token.NewFileSet(), "test.go", tc.source, 0)
			require.NoError(t, err)
			require.Len(t, file.Decls, 1)

			function, ok := file.Decls[0].(*ast.FuncDecl)
			require.True(t, ok)
			require.Equal(t, tc.want, testingTParameter(function))
		})
	}

	require.Empty(t, testingTParameter(&ast.FuncDecl{Type: &ast.FuncType{}}))
}

func TestCheckWalkError(t *testing.T) {
	t.Parallel()

	err := checkWithWalkDir(
		cryptoutilCmdCicdCommon.NewLogger("test-parallel-tests"),
		func(_ string, _ fs.WalkDirFunc) error {
			return errors.New("injected walk failure")
		},
	)

	require.ErrorContains(t, err, "walking test files")
	require.ErrorContains(t, err, "injected walk failure")
}

func TestCheckWalkCallbackError(t *testing.T) {
	t.Parallel()

	err := checkWithWalkDir(
		cryptoutilCmdCicdCommon.NewLogger("test-parallel-tests"),
		func(root string, walkFn fs.WalkDirFunc) error {
			return walkFn(root, nil, errors.New("injected callback failure"))
		},
	)

	require.ErrorContains(t, err, "walking test files")
	require.ErrorContains(t, err, "injected callback failure")
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}

		dir = parent
	}
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_Integration(t *testing.T) {
	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping - cannot find project root")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(root))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-parallel-tests")

	err = Check(logger)
	require.NoError(t, err)
}
