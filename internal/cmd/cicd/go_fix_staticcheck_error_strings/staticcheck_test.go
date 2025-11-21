package go_fix_staticcheck_error_strings

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
		name          string
		setupFiles    func(t *testing.T, tmpDir string)
		wantProcessed int
		wantModified  int
		wantIssued    int
		verifyFn      func(t *testing.T, tmpDir string)
	}{
		{
			name:          "empty_directory",
			setupFiles:    func(t *testing.T, tmpDir string) { t.Helper() },
			wantProcessed: 0,
			wantModified:  0,
			wantIssued:    0,
		},
		{
			name: "no_go_files",
			setupFiles: func(t *testing.T, tmpDir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test"), 0o600))
				require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte("key: value"), 0o600))
			},
			wantProcessed: 0,
			wantModified:  0,
			wantIssued:    0,
		},
		{
			name: "no_error_strings",
			setupFiles: func(t *testing.T, tmpDir string) {
				t.Helper()

				goFile := filepath.Join(tmpDir, "clean.go")
				content := "package test\n\nfunc Add(a, b int) int {\n\treturn a + b\n}\n"
				require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))
			},
			wantProcessed: 1,
			wantModified:  0,
			wantIssued:    0,
		},
		{
			name: "error_string_with_uppercase",
			setupFiles: func(t *testing.T, tmpDir string) {
				t.Helper()

				goFile := filepath.Join(tmpDir, "errors.go")
				content := "package test\n\nimport \"errors\"\n\nvar ErrInvalid = errors.New(\"Invalid input provided\")\n"
				require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))
			},
			wantProcessed: 1,
			wantModified:  1,
			wantIssued:    1,
			verifyFn: func(t *testing.T, tmpDir string) {
				t.Helper()

				fixed, err := os.ReadFile(filepath.Join(tmpDir, "errors.go"))
				require.NoError(t, err)
				require.Contains(t, string(fixed), "errors.New(\"invalid input provided\")")
				require.NotContains(t, string(fixed), "errors.New(\"Invalid input provided\")")
			},
		},
		{
			name: "error_string_with_acronym",
			setupFiles: func(t *testing.T, tmpDir string) {
				t.Helper()

				goFile := filepath.Join(tmpDir, "http_errors.go")
				content := "package test\n\nimport \"errors\"\n\nvar ErrHTTPFailed = errors.New(\"HTTP request failed\")\nvar ErrJSONInvalid = errors.New(\"JSON parsing error\")\nvar ErrURLMalformed = errors.New(\"URL is malformed\")\n"
				require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))
			},
			wantProcessed: 1,
			wantModified:  0,
			wantIssued:    0,
			verifyFn: func(t *testing.T, tmpDir string) {
				t.Helper()

				fixed, err := os.ReadFile(filepath.Join(tmpDir, "http_errors.go"))
				require.NoError(t, err)
				require.Contains(t, string(fixed), "errors.New(\"HTTP request failed\")")
				require.Contains(t, string(fixed), "errors.New(\"JSON parsing error\")")
				require.Contains(t, string(fixed), "errors.New(\"URL is malformed\")")
			},
		},
		{
			name: "multiple_error_strings",
			setupFiles: func(t *testing.T, tmpDir string) {
				t.Helper()

				goFile := filepath.Join(tmpDir, "multi_errors.go")
				content := "package test\n\nimport \"errors\"\n\nvar ErrOne = errors.New(\"First error occurred\")\nvar ErrTwo = errors.New(\"Second error occurred\")\nvar ErrThree = errors.New(\"Third error occurred\")\n"
				require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))
			},
			wantProcessed: 1,
			wantModified:  1,
			wantIssued:    3,
			verifyFn: func(t *testing.T, tmpDir string) {
				t.Helper()

				fixed, err := os.ReadFile(filepath.Join(tmpDir, "multi_errors.go"))
				require.NoError(t, err)
				require.Contains(t, string(fixed), "errors.New(\"first error occurred\")")
				require.Contains(t, string(fixed), "errors.New(\"second error occurred\")")
				require.Contains(t, string(fixed), "errors.New(\"third error occurred\")")
			},
		},
		{
			name: "mixed_acronyms_and_uppercase",
			setupFiles: func(t *testing.T, tmpDir string) {
				t.Helper()

				goFile := filepath.Join(tmpDir, "mixed.go")
				content := "package test\n\nimport \"errors\"\n\nvar ErrHTTP = errors.New(\"HTTP connection failed\")\nvar ErrGeneric = errors.New(\"Generic error occurred\")\nvar ErrJSON = errors.New(\"JSON decode error\")\nvar ErrBad = errors.New(\"Bad request received\")\n"
				require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))
			},
			wantProcessed: 1,
			wantModified:  1,
			wantIssued:    2,
			verifyFn: func(t *testing.T, tmpDir string) {
				t.Helper()

				fixed, err := os.ReadFile(filepath.Join(tmpDir, "mixed.go"))
				require.NoError(t, err)
				require.Contains(t, string(fixed), "errors.New(\"HTTP connection failed\")")
				require.Contains(t, string(fixed), "errors.New(\"generic error occurred\")")
				require.Contains(t, string(fixed), "errors.New(\"JSON decode error\")")
				require.Contains(t, string(fixed), "errors.New(\"bad request received\")")
			},
		},
		{
			name: "fmt_errorf",
			setupFiles: func(t *testing.T, tmpDir string) {
				t.Helper()

				goFile := filepath.Join(tmpDir, "fmt_errors.go")
				content := "package test\n\nimport \"fmt\"\n\nvar ErrFmt = fmt.Errorf(\"Failed to process request\")\n"
				require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))
			},
			wantProcessed: 1,
			wantModified:  1,
			wantIssued:    1,
			verifyFn: func(t *testing.T, tmpDir string) {
				t.Helper()

				fixed, err := os.ReadFile(filepath.Join(tmpDir, "fmt_errors.go"))
				require.NoError(t, err)
				require.Contains(t, string(fixed), "fmt.Errorf(\"failed to process request\")")
			},
		},
		{
			name: "test_files_skipped",
			setupFiles: func(t *testing.T, tmpDir string) {
				t.Helper()

				testFile := filepath.Join(tmpDir, "errors_test.go")
				content := "package test\n\nimport \"errors\"\n\nvar ErrTest = errors.New(\"Test error occurred\")\n"
				require.NoError(t, os.WriteFile(testFile, []byte(content), 0o600))
			},
			wantProcessed: 0,
			wantModified:  0,
			wantIssued:    0,
		},
		{
			name: "nested_directories",
			setupFiles: func(t *testing.T, tmpDir string) {
				t.Helper()

				subDir := filepath.Join(tmpDir, "sub", "nested")
				require.NoError(t, os.MkdirAll(subDir, 0o755))

				file1 := filepath.Join(tmpDir, "root.go")
				file2 := filepath.Join(tmpDir, "sub", "mid.go")
				file3 := filepath.Join(subDir, "deep.go")

				content := "package test\nimport \"errors\"\nvar Err = errors.New(\"Error occurred\")\n"
				require.NoError(t, os.WriteFile(file1, []byte(content), 0o600))
				require.NoError(t, os.WriteFile(file2, []byte(content), 0o600))
				require.NoError(t, os.WriteFile(file3, []byte(content), 0o600))
			},
			wantProcessed: 3,
			wantModified:  3,
			wantIssued:    3,
		},
		{
			name: "const_error_strings",
			setupFiles: func(t *testing.T, tmpDir string) {
				t.Helper()

				goFile := filepath.Join(tmpDir, "const_errors.go")
				content := "package test\n\nimport \"errors\"\n\nconst (\n\terrMsg = \"Error message one\"\n)\n\nvar ErrConst = errors.New(\"Constant error occurred\")\n"
				require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))
			},
			wantProcessed: 1,
			wantModified:  1,
			wantIssued:    1,
			verifyFn: func(t *testing.T, tmpDir string) {
				t.Helper()

				fixed, err := os.ReadFile(filepath.Join(tmpDir, "const_errors.go"))
				require.NoError(t, err)
				require.Contains(t, string(fixed), "errors.New(\"constant error occurred\")")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			logger := cryptoutilCmdCicdCommon.NewLogger("test-fix-" + tc.name)

			tc.setupFiles(t, tmpDir)

			processed, modified, issuesFixed, err := Fix(logger, tmpDir)
			require.NoError(t, err)
			require.Equal(t, tc.wantProcessed, processed)
			require.Equal(t, tc.wantModified, modified)
			require.Equal(t, tc.wantIssued, issuesFixed)

			if tc.verifyFn != nil {
				tc.verifyFn(t, tmpDir)
			}
		})
	}
}

func TestFix_InvalidDirectory(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-fix-invalid")

	processed, modified, issuesFixed, err := Fix(logger, "/nonexistent/path/to/nowhere")
	require.Error(t, err)
	require.Equal(t, 0, processed)
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}
