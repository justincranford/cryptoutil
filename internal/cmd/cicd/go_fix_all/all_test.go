package go_fix_all

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
		setupFiles      func(t *testing.T, tmpDir string)
		goVersion       string
		wantProcessed   int
		wantModified    int
		wantIssuesFixed int
		verifyFn        func(t *testing.T, tmpDir string)
	}{
		{
			name:            "empty_directory",
			setupFiles:      func(t *testing.T, tmpDir string) { t.Helper() },
			goVersion:       "1.25.4",
			wantProcessed:   0,
			wantModified:    0,
			wantIssuesFixed: 0,
		},
		{
			name: "all_fix_types",
			setupFiles: func(t *testing.T, tmpDir string) {
				t.Helper()

				staticcheckContent := "package test\n\nimport \"errors\"\n\nvar ErrFailed = errors.New(\"Failed to process\")\n"
				require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "errors.go"), []byte(staticcheckContent), 0o600))

				copyloopvarContent := "package test\n\nfunc Process(items []int) {\n\tfor _, item := range items {\n\t\titem := item\n\t\tprintln(item)\n\t}\n}\n"
				require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "loop.go"), []byte(copyloopvarContent), 0o600))

				thelperContent := "package test\n\nimport \"testing\"\n\nfunc setupTest(t *testing.T) {\n\tt.Log(\"setup\")\n}\n"
				require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "helpers_test.go"), []byte(thelperContent), 0o600))
			},
			goVersion:       "1.25.4",
			wantProcessed:   5,
			wantModified:    3,
			wantIssuesFixed: 3,
			verifyFn: func(t *testing.T, tmpDir string) {
				t.Helper()

				staticcheckFixed, err := os.ReadFile(filepath.Join(tmpDir, "errors.go"))
				require.NoError(t, err)
				require.Contains(t, string(staticcheckFixed), "errors.New(\"failed to process\")")

				copyloopvarFixed, err := os.ReadFile(filepath.Join(tmpDir, "loop.go"))
				require.NoError(t, err)
				require.NotContains(t, string(copyloopvarFixed), "item := item")

				thelperFixed, err := os.ReadFile(filepath.Join(tmpDir, "helpers_test.go"))
				require.NoError(t, err)
				require.Contains(t, string(thelperFixed), "t.Helper()")
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			logger := cryptoutilCmdCicdCommon.NewLogger("test-all-" + tc.name)

			tc.setupFiles(t, tmpDir)

			processed, modified, issuesFixed, err := Fix(logger, tmpDir, tc.goVersion)
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

	logger := cryptoutilCmdCicdCommon.NewLogger("test-all-invalid")

	processed, modified, issuesFixed, err := Fix(logger, "/nonexistent/path", "1.25.4")
	require.Error(t, err)
	require.Equal(t, 0, processed)
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}
