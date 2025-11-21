package go_fix_thelper

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
)

func TestFix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		setupFiles      func(t *testing.T, dir string) error
		wantProcessed   int
		wantModified    int
		wantIssuesFixed int
		verifyFn        func(t *testing.T, dir string)
	}{
		{
			name:            "empty directory",
			setupFiles:      func(t *testing.T, dir string) error { t.Helper(); return nil },
			wantProcessed:   0,
			wantModified:    0,
			wantIssuesFixed: 0,
		},
		{
			name: "no test files",
			setupFiles: func(t *testing.T, dir string) error {
				t.Helper()

				goFile := filepath.Join(dir, "main.go")
				content := `package main

func main() {}
`

				return os.WriteFile(goFile, []byte(content), 0o600)
			},
			wantProcessed:   0,
			wantModified:    0,
			wantIssuesFixed: 0,
		},
		{
			name: "no helper functions",
			setupFiles: func(t *testing.T, dir string) error {
				t.Helper()

				testFile := filepath.Join(dir, "test_test.go")
				content := `package test

import "testing"

func TestExample(t *testing.T) {
	t.Log("test")
}
`

				return os.WriteFile(testFile, []byte(content), 0o600)
			},
			wantProcessed:   1,
			wantModified:    0,
			wantIssuesFixed: 0,
		},
		{
			name: "helper function missing t.Helper()",
			setupFiles: func(t *testing.T, dir string) error {
				t.Helper()

				testFile := filepath.Join(dir, "helpers_test.go")
				content := `package test

import "testing"

func setupTest(t *testing.T) {
	t.Log("setup")
}
`

				return os.WriteFile(testFile, []byte(content), 0o600)
			},
			wantProcessed:   1,
			wantModified:    1,
			wantIssuesFixed: 1,
			verifyFn: func(t *testing.T, dir string) {
				t.Helper()

				testFile := filepath.Join(dir, "helpers_test.go")
				fixed, err := os.ReadFile(testFile)
				require.NoError(t, err)
				require.Contains(t, string(fixed), "t.Helper()")
			},
		},
		{
			name: "helper function with t.Helper()",
			setupFiles: func(t *testing.T, dir string) error {
				t.Helper()

				testFile := filepath.Join(dir, "helpers_test.go")
				content := `package test

import "testing"

func setupTest(t *testing.T) {
	t.Helper()
	t.Log("setup")
}
`

				return os.WriteFile(testFile, []byte(content), 0o600)
			},
			wantProcessed:   1,
			wantModified:    0,
			wantIssuesFixed: 0,
		},
		{
			name: "multiple helper functions",
			setupFiles: func(t *testing.T, dir string) error {
				t.Helper()

				testFile := filepath.Join(dir, "helpers_test.go")
				content := `package test

import "testing"

func setupTest(t *testing.T) {
	t.Log("setup")
}

func checkResult(t *testing.T, expected int) {
	t.Log("checking")
}

func assertValid(t *testing.T) {
	t.Log("asserting")
}
`

				return os.WriteFile(testFile, []byte(content), 0o600)
			},
			wantProcessed:   1,
			wantModified:    1,
			wantIssuesFixed: 3,
			verifyFn: func(t *testing.T, dir string) {
				t.Helper()

				testFile := filepath.Join(dir, "helpers_test.go")
				fixed, err := os.ReadFile(testFile)
				require.NoError(t, err)
				require.Contains(t, string(fixed), "t.Helper()")
			},
		},
		{
			name: "helper function patterns",
			setupFiles: func(t *testing.T, dir string) error {
				t.Helper()

				testFile := filepath.Join(dir, "patterns_test.go")
				content := `package test

import "testing"

func setupEnvironment(t *testing.T) {}
func checkData(t *testing.T) {}
func assertCondition(t *testing.T) {}
func verifyState(t *testing.T) {}
func helperFunction(t *testing.T) {}
func createMock(t *testing.T) {}
func buildFixture(t *testing.T) {}
func mockService(t *testing.T) {}
`

				return os.WriteFile(testFile, []byte(content), 0o600)
			},
			wantProcessed:   1,
			wantModified:    1,
			wantIssuesFixed: 8,
		},
		{
			name: "nested directories",
			setupFiles: func(t *testing.T, dir string) error {
				t.Helper()

				subDir := filepath.Join(dir, "sub", "nested")
				if err := os.MkdirAll(subDir, 0o755); err != nil {
					return err
				}

				content := `package test

import "testing"

func setupTest(t *testing.T) {
	t.Log("setup")
}
`
				file1 := filepath.Join(dir, "test1_test.go")
				file2 := filepath.Join(dir, "sub", "test2_test.go")
				file3 := filepath.Join(subDir, "test3_test.go")

				if err := os.WriteFile(file1, []byte(content), 0o600); err != nil {
					return err
				}

				if err := os.WriteFile(file2, []byte(content), 0o600); err != nil {
					return err
				}

				return os.WriteFile(file3, []byte(content), 0o600)
			},
			wantProcessed:   3,
			wantModified:    3,
			wantIssuesFixed: 3,
		},
		{
			name: "helper without testing param",
			setupFiles: func(t *testing.T, dir string) error {
				t.Helper()

				testFile := filepath.Join(dir, "helpers_test.go")
				content := `package test

func setupGlobal() {
	// No testing.T parameter
}
`

				return os.WriteFile(testFile, []byte(content), 0o600)
			},
			wantProcessed:   1,
			wantModified:    0,
			wantIssuesFixed: 0,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			logger := cryptoutilCmdCicdCommon.NewLogger("test-thelper-" + tc.name)

			if tc.setupFiles != nil {
				require.NoError(t, tc.setupFiles(t, tmpDir))
			}

			processed, modified, issuesFixed, err := Fix(logger, tmpDir)
			require.NoError(t, err)
			require.Equal(t, tc.wantProcessed, processed)
			require.Equal(t, tc.wantModified, modified)
			require.Equal(t, tc.wantIssuesFixed, issuesFixed)

			if tc.verifyFn != nil {
				tc.verifyFn(t, tmpDir)
			}
		})
	}
}

func TestFix_InvalidDirectory(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-thelper")

	processed, modified, issuesFixed, err := Fix(logger, "/nonexistent/path")
	require.Error(t, err)
	require.Equal(t, 0, processed)
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}
