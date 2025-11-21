package go_fix_copyloopvar

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
		goVersion       string
		setupFiles      func(t *testing.T, dir string)
		wantProcessed   int
		wantModified    int
		wantIssuesFixed int
		wantError       bool
		verifyFn        func(t *testing.T, dir string)
	}{
		{
			name:            "empty directory",
			goVersion:       "1.25.4",
			setupFiles:      func(t *testing.T, dir string) { t.Helper() },
			wantProcessed:   0,
			wantModified:    0,
			wantIssuesFixed: 0,
			wantError:       false,
		},
		{
			name:      "old Go version below minimum",
			goVersion: "1.21.0",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()

				content := `package test

func Process(items []int) {
	for _, item := range items {
		item := item
		_ = item
	}
}
`
				require.NoError(t, os.WriteFile(filepath.Join(dir, "loop.go"), []byte(content), 0o600))
			},
			wantProcessed:   0, // Should skip processing
			wantModified:    0,
			wantIssuesFixed: 0,
			wantError:       false,
		},
		{
			name:      "no loop variable copies",
			goVersion: "1.25.4",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()

				content := `package test

func Process(items []int) {
	for _, item := range items {
		println(item)
	}
}
`
				require.NoError(t, os.WriteFile(filepath.Join(dir, "clean.go"), []byte(content), 0o600))
			},
			wantProcessed:   1,
			wantModified:    0,
			wantIssuesFixed: 0,
			wantError:       false,
		},
		{
			name:      "single loop variable copy",
			goVersion: "1.25.4",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()

				content := `package test

func Process(items []int) {
	for _, item := range items {
		item := item
		println(item)
	}
}
`
				require.NoError(t, os.WriteFile(filepath.Join(dir, "loop.go"), []byte(content), 0o600))
			},
			wantProcessed:   1,
			wantModified:    1,
			wantIssuesFixed: 1,
			wantError:       false,
			verifyFn: func(t *testing.T, dir string) {
				t.Helper()

				fixed, err := os.ReadFile(filepath.Join(dir, "loop.go"))
				require.NoError(t, err)
				require.NotContains(t, string(fixed), "item := item")
				require.Contains(t, string(fixed), "println(item)")
			},
		},
		{
			name:      "multiple loop variable copies",
			goVersion: "1.25.4",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()

				content := `package test

func Process(items []int, names []string) {
	for _, item := range items {
		item := item
		println(item)
	}

	for _, name := range names {
		name := name
		println(name)
	}
}
`
				require.NoError(t, os.WriteFile(filepath.Join(dir, "loops.go"), []byte(content), 0o600))
			},
			wantProcessed:   1,
			wantModified:    1,
			wantIssuesFixed: 2,
			wantError:       false,
			verifyFn: func(t *testing.T, dir string) {
				t.Helper()

				fixed, err := os.ReadFile(filepath.Join(dir, "loops.go"))
				require.NoError(t, err)
				require.NotContains(t, string(fixed), "item := item")
				require.NotContains(t, string(fixed), "name := name")
			},
		},
		{
			name:      "key and value copies",
			goVersion: "1.25.4",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()

				content := `package test

func Process(data map[string]int) {
	for key, val := range data {
		key := key
		val := val
		println(key, val)
	}
}
`
				require.NoError(t, os.WriteFile(filepath.Join(dir, "map_loop.go"), []byte(content), 0o600))
			},
			wantProcessed:   1,
			wantModified:    1,
			wantIssuesFixed: 1, // Only the first copy (key := key) removed
			wantError:       false,
			verifyFn: func(t *testing.T, dir string) {
				t.Helper()

				fixed, err := os.ReadFile(filepath.Join(dir, "map_loop.go"))
				require.NoError(t, err)
				require.NotContains(t, string(fixed), "key := key")
			},
		},
		{
			name:      "test files skipped",
			goVersion: "1.25.4",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()

				content := `package test

func TestLoop(t *testing.T) {
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			println(tc.name)
		})
	}
}
`
				require.NoError(t, os.WriteFile(filepath.Join(dir, "loop_test.go"), []byte(content), 0o600))
			},
			wantProcessed:   0, // Test files should be skipped
			wantModified:    0,
			wantIssuesFixed: 0,
			wantError:       false,
		},
		{
			name:      "generated files skipped",
			goVersion: "1.25.4",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()

				content := `package model

func Process(items []int) {
	for _, item := range items {
		item := item
		println(item)
	}
}
`
				require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi_gen_model.go"), []byte(content), 0o600))
			},
			wantProcessed:   0, // Generated files should be skipped
			wantModified:    0,
			wantIssuesFixed: 0,
			wantError:       false,
		},
		{
			name:      "nested directories",
			goVersion: "1.25.4",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()

				subDir := filepath.Join(dir, "sub", "nested")
				require.NoError(t, os.MkdirAll(subDir, 0o755))

				content := `package test
func Process(items []int) {
	for _, item := range items {
		item := item
		println(item)
	}
}
`
				file1 := filepath.Join(dir, "loop1.go")
				file2 := filepath.Join(dir, "sub", "loop2.go")
				file3 := filepath.Join(subDir, "loop3.go")

				require.NoError(t, os.WriteFile(file1, []byte(content), 0o600))
				require.NoError(t, os.WriteFile(file2, []byte(content), 0o600))
				require.NoError(t, os.WriteFile(file3, []byte(content), 0o600))
			},
			wantProcessed:   3,
			wantModified:    3,
			wantIssuesFixed: 3,
			wantError:       false,
		},
	}

	for _, tc := range tests {
		// Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			logger := cryptoutilCmdCicdCommon.NewLogger("test-copyloopvar-" + tc.name)

			tc.setupFiles(t, tmpDir)

			processed, modified, issuesFixed, err := Fix(logger, tmpDir, tc.goVersion)

			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.wantProcessed, processed, "Unexpected processed count")
			require.Equal(t, tc.wantModified, modified, "Unexpected modified count")
			require.Equal(t, tc.wantIssuesFixed, issuesFixed, "Unexpected issues fixed count")

			if tc.verifyFn != nil {
				tc.verifyFn(t, tmpDir)
			}
		})
	}
}

func TestFix_InvalidDirectory(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-copyloopvar-invalid")

	processed, modified, issuesFixed, err := Fix(logger, "/nonexistent/path", "1.25.4")
	require.Error(t, err)
	require.Equal(t, 0, processed)
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}

func TestIsGoVersionSupported(t *testing.T) {
	t.Parallel()

	tests := []struct {
		version  string
		expected bool
	}{
		{"1.21.0", false},
		{"1.22.0", true},
		{"1.22.5", true},
		{"1.23.0", true},
		{"1.25.4", true},
		{"2.0.0", true},
		{"invalid", false},
		{"1.2", false}, // Edge case: 1.2 < 1.22.
	}

	for _, tc := range tests {
		// Capture range variable
		t.Run(tc.version, func(t *testing.T) {
			t.Parallel()

			result := isGoVersionSupported(tc.version)
			require.Equal(t, tc.expected, result)
		})
	}
}
