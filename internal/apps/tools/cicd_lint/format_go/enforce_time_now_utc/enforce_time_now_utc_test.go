// Copyright (c) 2025 Justin Cranford

package enforce_time_now_utc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestProcessGoFileForTimeNowUTC_Replacements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		content          string
		wantReplacements int
		wantContains     []string
		wantNotContains  []string
		wantUnchanged    bool
	}{
		{
			name: "basic replacement",
			content: `package main

import "time"

func main() {
now := time.Now()
println(now)
}
`,
			wantReplacements: 1,
			wantContains:     []string{"time.Now().UTC()"},
			wantNotContains:  []string{"time.Now()\n"},
		},
		{
			name: "already correct",
			content: `package main

import "time"

func main() {
now := time.Now().UTC()
println(now)
}
`,
			wantReplacements: 0,
			wantUnchanged:    true,
		},
		{
			name: "chained method calls",
			content: `package main

import "time"

func main() {
later := time.Now().Add(1 * time.Hour)
println(later)
}
`,
			wantReplacements: 1,
			wantContains:     []string{"time.Now().UTC().Add(1 * time.Hour)"},
		},
		{
			name: "variable assignment",
			content: `package main

import "time"

func main() {
t := time.Now()
later := t.Add(1 * time.Hour)
println(later)
}
`,
			wantReplacements: 1,
			wantContains:     []string{"t := time.Now().UTC()"},
		},
		{
			name: "UTC on variable not time.Now",
			content: `package main

import "time"

func main() {
var t time.Time
_ = t.UTC()
}
`,
			wantReplacements: 0,
		},
		{
			name: "inner call func not selector",
			content: `package foo

import "time"

func gettime() time.Time { return time.Now().UTC() }

func f() { _ = gettime().UTC() }
`,
			wantReplacements: 0,
		},
		{
			name: "inner sel not Now",
			content: `package foo

import "time"

func f() { _ = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UTC() }
`,
			wantReplacements: 0,
		},
		{
			name: "inner ident not time",
			content: `package foo

import "time"

type fakeTimer struct{}

func (f fakeTimer) Now() time.Time { return time.Now().UTC() }

func g() {
var ft fakeTimer
_ = ft.Now().UTC()
_ = ft.Now()
}
`,
			wantReplacements: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.go")

			err := os.WriteFile(testFile, []byte(tc.content), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			replacements, err := ProcessGoFileForTimeNowUTC(testFile)
			require.NoError(t, err)
			require.Equal(t, tc.wantReplacements, replacements)

			if tc.wantUnchanged {
				modifiedContent, err := os.ReadFile(testFile)
				require.NoError(t, err)
				require.Equal(t, tc.content, string(modifiedContent))
			}

			if len(tc.wantContains) > 0 || len(tc.wantNotContains) > 0 {
				modifiedContent, err := os.ReadFile(testFile)
				require.NoError(t, err)

				modifiedStr := string(modifiedContent)

				for _, want := range tc.wantContains {
					require.Contains(t, modifiedStr, want)
				}

				for _, notWant := range tc.wantNotContains {
					require.NotContains(t, modifiedStr, notWant)
				}
			}
		})
	}
}

func TestProcessGoFileForTimeNowUTC_SpecialPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		setupFn          func(t *testing.T) string
		wantReplacements int
		wantErr          string
	}{
		{
			name:             "read error on nonexistent file",
			setupFn:          func(_ *testing.T) string { return "/nonexistent/path/to/test.go" },
			wantReplacements: 0,
			wantErr:          "failed to read file",
		},
		{
			name:             "self-exclusion test file",
			setupFn:          func(_ *testing.T) string { return "enforce_time_now_utc_test.go" },
			wantReplacements: 0,
		},
		{
			name:             "self-exclusion impl file",
			setupFn:          func(_ *testing.T) string { return "enforce_time_now_utc.go" },
			wantReplacements: 0,
		},
		{
			name: "self-modification path exclusion",
			setupFn: func(t *testing.T) string {
				t.Helper()

				tmpDir := t.TempDir()
				fakeFormatGoDir := filepath.Join(tmpDir, "internal", "cmd", "cicd", "format_go")
				require.NoError(t, os.MkdirAll(fakeFormatGoDir, 0o700))

				goFile := filepath.Join(fakeFormatGoDir, "dummy.go")
				require.NoError(t, os.WriteFile(goFile, []byte("package format_go\n\nimport \"time\"\n\nvar t = time.Now()\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return goFile
			},
			wantReplacements: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			filePath := tc.setupFn(t)

			replacements, err := ProcessGoFileForTimeNowUTC(filePath)
			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.wantReplacements, replacements)
		})
	}
}

func TestEnforce_Scenarios(t *testing.T) {
	t.Parallel()

	noModContent := `package main

import "time"

func main() {
	now := time.Now().UTC()
	later := time.Now().UTC().Add(1 * time.Hour)
}
`

	tests := []struct {
		name     string
		files    map[string]string
		wantErr  string
		validate func(t *testing.T, dir string)
	}{
		{
			name: "integration with multiple replacements",
			files: map[string]string{
				"test.go": `package main

import "time"

func main() {
	now := time.Now()
	later := time.Now().Add(1 * time.Hour)
	alreadyCorrect := time.Now().UTC()
}
`,
			},
			wantErr: "modified 1 files",
			validate: func(t *testing.T, dir string) {
				t.Helper()

				content, err := os.ReadFile(filepath.Join(dir, "test.go"))
				require.NoError(t, err)

				modifiedStr := string(content)
				require.Equal(t, 3, strings.Count(modifiedStr, "time.Now().UTC()"))
				require.NotContains(t, modifiedStr, "time.Now()\n")
				require.NotContains(t, modifiedStr, "time.Now().Add")
			},
		},
		{
			name:  "no modifications needed",
			files: map[string]string{"test.go": noModContent},
			validate: func(t *testing.T, dir string) {
				t.Helper()

				content, err := os.ReadFile(filepath.Join(dir, "test.go"))
				require.NoError(t, err)
				require.Equal(t, noModContent, string(content))
			},
		},
		{
			name:  "invalid Go file parsed without error",
			files: map[string]string{"invalid.go": "package main\n\nthis is not valid Go code!\n"},
		},
		{
			name: "empty file map",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			filesByExtension := map[string][]string{}

			for filename, content := range tc.files {
				testFile := filepath.Join(tmpDir, filename)

				err := os.WriteFile(testFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
				require.NoError(t, err)

				filesByExtension["go"] = append(filesByExtension["go"], testFile)
			}

			logger := cryptoutilCmdCicdCommon.NewLogger("test-enforce-time-now-utc")
			err := Enforce(logger, filesByExtension)

			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tc.validate != nil {
				tc.validate(t, tmpDir)
			}
		})
	}
}
