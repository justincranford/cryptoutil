// Copyright (c) 2025 Justin Cranford

package go_enforce_test_patterns_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/cmd/cicd/common"
	"cryptoutil/internal/cmd/cicd/go_enforce_test_patterns"
)

const (
	testValidTestFileContent = "package example_test\n\nimport (\n\t\"testing\"\n\t\"github.com/stretchr/testify/require\"\n\tgoogleUuid \"github.com/google/uuid\"\n)\n\nfunc TestExample(t *testing.T) {\n\tid := googleUuid.NewV7()\n\trequire.NotNil(t, id)\n}\n"
)

func TestCheckTestFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		content         string
		wantIssueCount  int
		wantContains    []string
		nonExistentFile bool
	}{
		{
			name:           "valid_test_file",
			content:        testValidTestFileContent,
			wantIssueCount: 0,
		},
		{
			name:           "uuid_new",
			content:        "package example_test\n\nimport (\n\t\"testing\"\n\t\"github.com/google/uuid\"\n)\n\nfunc TestExample(t *testing.T) {\n\tid := uuid.New()\n}\n",
			wantIssueCount: 1,
			wantContains:   []string{"uuid.New()", "uuid.NewV7()"},
		},
		{
			name:           "hardcoded_uuid",
			content:        "package example_test\n\nimport \"testing\"\n\nfunc TestExample(t *testing.T) {\n\tid := \"550e8400-e29b-41d4-a716-446655440000\"\n}\n",
			wantIssueCount: 1,
			wantContains:   []string{"hardcoded UUID", "uuid.NewV7()"},
		},
		{
			name:           "t_errorf",
			content:        "package example_test\n\nimport \"testing\"\n\nfunc TestExample(t *testing.T) {\n\tt.Errorf(\"expected %d, got %d\", 1, 2)\n}\n",
			wantIssueCount: 1,
			wantContains:   []string{"t.Errorf()", "require.Errorf()"},
		},
		{
			name:           "t_fatalf",
			content:        "package example_test\n\nimport \"testing\"\n\nfunc TestExample(t *testing.T) {\n\tt.Fatalf(\"test failed: %v\", err)\n}\n",
			wantIssueCount: 1,
			wantContains:   []string{"t.Fatalf()", "require.Fatalf()"},
		},
		{
			name:           "testify_usage_without_import",
			content:        "package example_test\n\nimport \"testing\"\n\nfunc TestExample(t *testing.T) {\n\trequire.Equal(t, 1, 1)\n}\n",
			wantIssueCount: 1,
			wantContains:   []string{"testify assertions", "doesn't import testify"},
		},
		{
			name:           "multiple_issues",
			content:        "package example_test\n\nimport (\n\t\"testing\"\n\t\"github.com/google/uuid\"\n)\n\nfunc TestExample(t *testing.T) {\n\tid1 := uuid.New()\n\tid2 := \"550e8400-e29b-41d4-a716-446655440000\"\n\tt.Errorf(\"error\")\n\tt.Fatalf(\"fatal\")\n\trequire.Equal(t, 1, 1)\n}\n",
			wantIssueCount: 5,
			wantContains:   []string{"uuid.New()", "hardcoded UUID", "t.Errorf()", "t.Fatalf()", "doesn't import testify"},
		},
		{
			name:            "non_existent_file",
			nonExistentFile: true,
			wantIssueCount:  1,
			wantContains:    []string{"Error reading file"},
		},
	}

	for _, tc := range tests {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var testFile string
			if tc.nonExistentFile {
				testFile = "/nonexistent/path/test.go"
			} else {
				tempDir := t.TempDir()
				testFile = filepath.Join(tempDir, tc.name+"_test.go")
				err := os.WriteFile(testFile, []byte(tc.content), 0o600)
				require.NoError(t, err)
			}

			issues := go_enforce_test_patterns.CheckTestFile(testFile)

			require.Len(t, issues, tc.wantIssueCount)

			for _, want := range tc.wantContains {
				found := false

				for _, issue := range issues {
					if strings.Contains(issue, want) {
						found = true

						break
					}
				}

				require.True(t, found, "Expected to find %q in issues", want)
			}
		})
	}
}

func TestEnforce(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setupFiles   func(t *testing.T, tmpDir string) []string
		wantError    bool
		wantContains string
	}{
		{
			name: "no_test_files",
			setupFiles: func(t *testing.T, tmpDir string) []string {
				t.Helper()

				return []string{"main.go", "util.go", "config.go"}
			},
			wantError: false,
		},
		{
			name: "valid_test_files",
			setupFiles: func(t *testing.T, tmpDir string) []string {
				t.Helper()

				testFile := filepath.Join(tmpDir, "valid_test.go")
				err := os.WriteFile(testFile, []byte(testValidTestFileContent), 0o600)
				require.NoError(t, err)

				return []string{testFile}
			},
			wantError: false,
		},
		{
			name: "invalid_test_file",
			setupFiles: func(t *testing.T, tmpDir string) []string {
				t.Helper()

				testFile := filepath.Join(tmpDir, "invalid_test.go")
				content := "package example_test\n\nimport (\n\t\"testing\"\n\tgoogleUuid \"github.com/google/uuid\"\n)\n\nfunc TestExample(t *testing.T) {\n\tid := googleUuid.New()\n\tt.Errorf(\"error\")\n}\n"
				err := os.WriteFile(testFile, []byte(content), 0o600)
				require.NoError(t, err)

				return []string{testFile}
			},
			wantError:    true,
			wantContains: "test pattern violations",
		},
		{
			name: "excluded_files",
			setupFiles: func(t *testing.T, tmpDir string) []string {
				t.Helper()

				cicdTestFile := filepath.Join(tmpDir, "cicd_test.go")
				cicdContent := "package cicd_test\n\nimport \"testing\"\n\nfunc TestExample(t *testing.T) {\n\tt.Errorf(\"deliberate violation\")\n}\n"
				err := os.WriteFile(cicdTestFile, []byte(cicdContent), 0o600)
				require.NoError(t, err)

				return []string{cicdTestFile}
			},
			wantError: false,
		},
		{
			name: "multiple_files_with_issues",
			setupFiles: func(t *testing.T, tmpDir string) []string {
				t.Helper()

				file1 := filepath.Join(tmpDir, "test1_test.go")
				content1 := "package example_test\n\nimport (\n\t\"testing\"\n\t\"github.com/google/uuid\"\n)\n\nfunc TestExample1(t *testing.T) {\n\tid := uuid.New()\n}\n"
				err := os.WriteFile(file1, []byte(content1), 0o600)
				require.NoError(t, err)

				file2 := filepath.Join(tmpDir, "test2_test.go")
				content2 := "package test2_test\n\nimport \"testing\"\n\nfunc TestExample2(t *testing.T) {\n\tt.Errorf(\"error\")\n}\n"
				err = os.WriteFile(file2, []byte(content2), 0o600)
				require.NoError(t, err)

				return []string{file1, file2}
			},
			wantError:    true,
			wantContains: "2 files",
		},
	}

	for _, tc := range tests {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			logger := common.NewLogger("test-enforce-" + tc.name)

			allFiles := tc.setupFiles(t, tmpDir)

			err := go_enforce_test_patterns.Enforce(logger, allFiles)

			if tc.wantError {
				require.Error(t, err)

				if tc.wantContains != "" {
					require.Contains(t, err.Error(), tc.wantContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
