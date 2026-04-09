// Copyright (c) 2025 Justin Cranford

package format_gotest_test

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilCmdCicdFormatGotest "cryptoutil/internal/apps/tools/cicd_lint/format_gotest"
)

const contentHelperNeedingFix = `package example

import "testing"

func setupTest(t *testing.T) {
	doSomething()
}

func doSomething() {}
`

const contentBenchmarkHelper = `package example

import "testing"

func setupBenchmark(b *testing.B) {
	doSomething()
}

func doSomething() {}

func BenchmarkExample(b *testing.B) {
	setupBenchmark(b)
}
`

const contentPointerReceiver = `package example

import "testing"

type TestSuite struct{}

func (s *TestSuite) setupHelper(t *testing.T) {
	doSomething()
}

func doSomething() {}

func TestExample(t *testing.T) {
	t.Parallel()
	s := &TestSuite{}
	s.setupHelper(t)
}
`

const contentMixedStatements = `package example

import "testing"

func setupHelper(t *testing.T) {
	x := 1
	_ = x
	doSomething()
	t.Log("setup")
}

func doSomething() {}

func TestExample(t *testing.T) {
	t.Parallel()
	setupHelper(t)
}
`

const contentHelperWithoutTestingT = `package example

import "testing"

func setupData() string {
	return "data"
}

func TestExample(t *testing.T) {
	t.Parallel()
	data := setupData()
	_ = data
}
`

const contentNonPointerTestingT = `package example

import "testing"

func setupByValue(t testing.T) {
	doSomething()
}

func doSomething() {}

func TestExample(t *testing.T) {
	t.Parallel()
	setupByValue(*t)
}
`

const contentNoParams = `package example

import "testing"

func setupNoParams() {
	doSomething()
}

func doSomething() {}

func TestExample(t *testing.T) {
	t.Parallel()
	setupNoParams()
}
`

const contentNonTestingPointer = `package example

import "testing"

type MyType struct{ Name string }

func setupWithPointer(m *MyType) {
	doSomething()
}

func doSomething() {}

func TestExample(t *testing.T) {
	t.Parallel()
	m := &MyType{Name: "test"}
	setupWithPointer(m)
}
`

const contentArrayPointer = `package example

import "testing"

func setupWithArray(arr *[3]int) {
	doSomething()
}

func doSomething() {}

func TestExample(t *testing.T) {
	t.Parallel()
	arr := [3]int{1, 2, 3}
	setupWithArray(&arr)
}
`

const contentInterfaceMethod = `package example

import "testing"

type TestInterface interface {
	Setup(t *testing.T)
}
`

const contentAlreadyHasHelper = `package example

import "testing"

func setupTest(t *testing.T) {
	t.Helper()
	// setup code
}
`

const contentInvalidGo = `package example

func invalidSyntax( {
	// missing closing paren
}
`

func TestFormatDir_HelperDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		content    string
		wantHelper bool
		helperStr  string
	}{
		{name: "adds helper to function with testing.T", content: contentHelperNeedingFix, wantHelper: true, helperStr: ".Helper()"},
		{name: "adds helper to benchmark function", content: contentBenchmarkHelper, wantHelper: true, helperStr: "b.Helper"},
		{name: "adds helper to pointer receiver method", content: contentPointerReceiver, wantHelper: true, helperStr: "t.Helper"},
		{name: "adds helper with mixed statements", content: contentMixedStatements, wantHelper: true, helperStr: "t.Helper"},
		{name: "skips function without testing.T param", content: contentHelperWithoutTestingT},
		{name: "skips non-pointer testing.T", content: contentNonPointerTestingT},
		{name: "skips function with no params", content: contentNoParams},
		{name: "skips non-testing pointer param", content: contentNonTestingPointer},
		{name: "skips array pointer param", content: contentArrayPointer},
		{name: "skips interface method without body", content: contentInterfaceMethod},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test_test.go")
			err := os.WriteFile(testFile, []byte(tc.content), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			err = cryptoutilCmdCicdFormatGotest.FormatDir(logger, tmpDir)
			require.NoError(t, err)

			modifiedContent, err := os.ReadFile(testFile)
			require.NoError(t, err)

			if tc.wantHelper {
				require.Contains(t, string(modifiedContent), tc.helperStr)
			} else {
				require.NotContains(t, string(modifiedContent), ".Helper")
			}
		})
	}
}

func TestFormatDir_NoModification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fileName string
		content  string
	}{
		{name: "non-test Go file", fileName: "main.go", content: "package main\n\nfunc main() {}\n"},
		{name: "already has helper", fileName: "helper_test.go", content: contentAlreadyHasHelper},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, tc.fileName)
			err := os.WriteFile(testFile, []byte(tc.content), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			err = cryptoutilCmdCicdFormatGotest.FormatDir(logger, tmpDir)
			require.NoError(t, err)

			modifiedContent, err := os.ReadFile(testFile)
			require.NoError(t, err)
			require.Equal(t, tc.content, string(modifiedContent))
		})
	}
}

func TestFormatDir_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		wantErr string
	}{
		{
			name: "invalid Go file",
			setup: func(t *testing.T) string {
				t.Helper()

				tmpDir := t.TempDir()
				err := os.WriteFile(filepath.Join(tmpDir, "invalid_test.go"), []byte(contentInvalidGo), cryptoutilSharedMagic.CacheFilePermissions)
				require.NoError(t, err)

				return tmpDir
			},
			wantErr: "format-go-test failed",
		},
		{
			name: "nonexistent directory",
			setup: func(t *testing.T) string {
				t.Helper()

				return "/nonexistent/directory/path"
			},
			wantErr: "format-go-test failed",
		},
		{
			name: "read-only file",
			setup: func(t *testing.T) string {
				t.Helper()

				tmpDir := t.TempDir()
				testFile := filepath.Join(tmpDir, "helper_test.go")
				err := os.WriteFile(testFile, []byte(contentHelperNeedingFix), cryptoutilSharedMagic.CacheFilePermissions)
				require.NoError(t, err)

				err = os.Chmod(testFile, 0o400)
				require.NoError(t, err)

				t.Cleanup(func() { _ = os.Chmod(testFile, cryptoutilSharedMagic.CacheFilePermissions) })

				return tmpDir
			},
			wantErr: "format-go-test failed",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := tc.setup(t)
			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			err := cryptoutilCmdCicdFormatGotest.FormatDir(logger, dir)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestFormat(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := cryptoutilCmdCicdFormatGotest.Format(logger)
	require.NoError(t, err)
}
