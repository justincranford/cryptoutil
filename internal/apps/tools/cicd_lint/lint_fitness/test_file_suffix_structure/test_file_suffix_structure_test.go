// Copyright (c) 2025 Justin Cranford

package test_file_suffix_structure

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------
// Test helpers
// -----------------------------------------------------------------------

// buildSuffixRoot creates a temp root dir with the test-file-suffix-rules YAML.
func buildSuffixRoot(t *testing.T) string {
	t.Helper()

	rootDir := t.TempDir()
	yamlDir := filepath.Dir(filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDTestFileSuffixRulesFile)))
	require.NoError(t, os.MkdirAll(yamlDir, 0o700))

	src := filepath.Join(".", "test-file-suffix-rules.yaml")
	data, err := os.ReadFile(src)
	require.NoError(t, err, "test-file-suffix-rules.yaml must exist in the package directory")

	destPath := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDTestFileSuffixRulesFile))
	require.NoError(t, os.WriteFile(destPath, data, cryptoutilSharedMagic.FilePermissionsDefault)) //nolint:gosec // G703: test setup writing to t.TempDir(); path is constructed from controlled inputs

	return rootDir
}

// writeTestFile writes content to a _test.go file in tmpDir/pkg/.
func writeTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	require.NoError(t, os.MkdirAll(dir, 0o700))
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))

	return path
}

// -----------------------------------------------------------------------
// LoadSuffixRules
// -----------------------------------------------------------------------

func TestLoadSuffixRules_HappyPath(t *testing.T) {
	t.Parallel()

	rootDir := buildSuffixRoot(t)

	rules, err := LoadSuffixRules(rootDir, os.ReadFile)

	require.NoError(t, err)
	require.NotEmpty(t, rules.SuffixRules)
	require.NotEmpty(t, rules.ContentRules)
}

func TestLoadSuffixRules_FileNotFound(t *testing.T) {
	t.Parallel()

	rules, err := LoadSuffixRules(t.TempDir(), os.ReadFile)

	require.Error(t, err)
	require.Nil(t, rules)
	require.Contains(t, err.Error(), "failed to read")
}

func TestLoadSuffixRules_InvalidYAML(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	yamlDir := filepath.Dir(filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDTestFileSuffixRulesFile)))
	require.NoError(t, os.MkdirAll(yamlDir, 0o700))

	destPath := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDTestFileSuffixRulesFile))
	require.NoError(t, os.WriteFile(destPath, []byte("!!! not: valid: yaml: ["), cryptoutilSharedMagic.FilePermissionsDefault))

	rules, err := LoadSuffixRules(rootDir, os.ReadFile)

	require.Error(t, err)
	require.Nil(t, rules)
	require.Contains(t, err.Error(), "failed to parse")
}

// -----------------------------------------------------------------------
// CheckFiles — suffix and content rules
// -----------------------------------------------------------------------

func TestCheckFiles_SuffixAndContentRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		filename    string
		content     string
		wantErr     bool
		errContains string
	}{
		{
			name:     "bench file has benchmark",
			filename: "foo_bench_test.go",
			content:  "package mypkg\nimport \"testing\"\nfunc BenchmarkFoo(b *testing.B) {}\n",
		},
		{
			name:        "bench file missing benchmark",
			filename:    "foo_bench_test.go",
			content:     "package mypkg\nimport \"testing\"\nfunc TestFoo(t *testing.T) {}\n",
			wantErr:     true,
			errContains: "violation(s)",
		},
		{
			name:        "bench file forbidden fuzz",
			filename:    "foo_bench_test.go",
			content:     "package mypkg\nimport \"testing\"\nfunc BenchmarkFoo(b *testing.B) {}\nfunc FuzzFoo(f *testing.F) {}\n",
			wantErr:     true,
			errContains: "violation(s)",
		},
		{
			name:     "fuzz file has fuzz",
			filename: "foo_fuzz_test.go",
			content:  "package mypkg\nimport \"testing\"\nfunc FuzzFoo(f *testing.F) {}\n",
		},
		{
			name:        "fuzz file missing fuzz",
			filename:    "foo_fuzz_test.go",
			content:     "package mypkg\nimport \"testing\"\nfunc TestFoo(t *testing.T) {}\n",
			wantErr:     true,
			errContains: "violation(s)",
		},
		{
			name:     "property file has build tag",
			filename: "foo_property_test.go",
			content:  "//go:build !fuzz\npackage mypkg\nimport \"testing\"\nfunc TestFooProp(t *testing.T) {}\n",
		},
		{
			name:        "property file missing build tag",
			filename:    "foo_property_test.go",
			content:     "package mypkg\nimport \"testing\"\nfunc TestFooProp(t *testing.T) {}\n",
			wantErr:     true,
			errContains: "violation(s)",
		},
		{
			name:     "integration file no fuzz",
			filename: "foo_integration_test.go",
			content:  "package mypkg\nimport \"testing\"\nfunc TestFoo(t *testing.T) {}\n",
		},
		{
			name:        "integration file forbidden fuzz",
			filename:    "foo_integration_test.go",
			content:     "package mypkg\nimport \"testing\"\nfunc FuzzFoo(f *testing.F) {}\n",
			wantErr:     true,
			errContains: "violation(s)",
		},
		{
			name:        "fuzz in wrong file suffix",
			filename:    "foo_test.go",
			content:     "package mypkg\nimport \"testing\"\nfunc FuzzFoo(f *testing.F) {}\n",
			wantErr:     true,
			errContains: "violation(s)",
		},
		{
			name:        "benchmark in wrong file suffix",
			filename:    "foo_test.go",
			content:     "package mypkg\nimport \"testing\"\nfunc BenchmarkFoo(b *testing.B) {}\n",
			wantErr:     true,
			errContains: "violation(s)",
		},
	}

	rootDir := buildSuffixRoot(t)
	rules, err := LoadSuffixRules(rootDir, os.ReadFile)
	require.NoError(t, err)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pkgDir := filepath.Join(t.TempDir(), "mypkg")
			f := writeTestFile(t, pkgDir, tc.filename, tc.content)
			logger := cryptoutilCmdCicdCommon.NewLogger("test")

			checkErr := CheckFiles(logger, []string{f}, rules, os.ReadFile)

			if tc.wantErr {
				require.Error(t, checkErr)
				require.Contains(t, checkErr.Error(), tc.errContains)
			} else {
				require.NoError(t, checkErr)
			}
		})
	}
}

func TestCheckFiles_EmptyFileList(t *testing.T) {
	t.Parallel()

	rootDir := buildSuffixRoot(t)
	rules, err := LoadSuffixRules(rootDir, os.ReadFile)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckFiles(logger, []string{}, rules, os.ReadFile)

	require.NoError(t, err)
}

func TestCheckFiles_ReadFileError(t *testing.T) {
	t.Parallel()

	rootDir := buildSuffixRoot(t)
	rules, err := LoadSuffixRules(rootDir, os.ReadFile)
	require.NoError(t, err)

	pkgDir := filepath.Join(t.TempDir(), "mypkg")
	f := writeTestFile(t, pkgDir, "foo_test.go", "package mypkg\n")

	stubReadFileFn := func(_ string) ([]byte, error) {
		return nil, fmt.Errorf("simulated read error")
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckFiles(logger, []string{f}, rules, stubReadFileFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "simulated read error")
}

func TestCheckFiles_InvalidRegexPatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		rules       *SuffixRules
		errContains string
	}{
		{
			name: "invalid content pattern in content rules",
			rules: &SuffixRules{
				ContentRules: []ContentRule{
					{ContentPattern: "[invalid-regex", RequiredSuffix: "_test.go"},
				},
			},
			errContains: "invalid content_pattern",
		},
		{
			name: "invalid required pattern in suffix rule",
			rules: &SuffixRules{
				SuffixRules: []SuffixRule{
					{Suffix: "_test.go", RequiredContentPatterns: []string{"[bad-regex"}},
				},
			},
			errContains: "invalid required_content_pattern",
		},
		{
			name: "invalid forbidden pattern in suffix rule",
			rules: &SuffixRules{
				SuffixRules: []SuffixRule{
					{Suffix: "_test.go", ForbiddenContentPatterns: []string{"[bad-regex"}},
				},
			},
			errContains: "invalid forbidden_content_pattern",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			pkgDir := filepath.Join(t.TempDir(), "mypkg")
			f := writeTestFile(t, pkgDir, "foo_test.go", "package mypkg\n")

			checkErr := CheckFiles(logger, []string{f}, tc.rules, os.ReadFile)

			require.Error(t, checkErr)
			require.Contains(t, checkErr.Error(), tc.errContains)
		})
	}
}

// -----------------------------------------------------------------------
// CheckInDir
// -----------------------------------------------------------------------

func TestCheckInDir_HappyPath(t *testing.T) {
	t.Parallel()

	rootDir := buildSuffixRoot(t)
	pkgDir := filepath.Join(rootDir, "mypkg")
	writeTestFile(t, pkgDir, "foo_bench_test.go",
		"package mypkg\nimport \"testing\"\nfunc BenchmarkFoo(b *testing.B) {}\n")
	writeTestFile(t, pkgDir, "foo_fuzz_test.go",
		"package mypkg\nimport \"testing\"\nfunc FuzzFoo(f *testing.F) {}\n")
	writeTestFile(t, pkgDir, "foo_test.go",
		"package mypkg\nimport \"testing\"\nfunc TestFoo(t *testing.T) {}\n")

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, rootDir)

	require.NoError(t, err)
}

func TestCheckInDir_ManifestError(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir() // no YAML
	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "test-file-suffix-structure")
}

func TestCheckInDir_WalkDirError(t *testing.T) {
	t.Parallel()

	stubWalkDirFn := func(_ string, _ fs.WalkDirFunc) error {
		return fmt.Errorf("simulated walk error")
	}

	rootDir := buildSuffixRoot(t)
	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := checkInDir(logger, rootDir, os.ReadFile, stubWalkDirFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "simulated walk error")
}

func TestCheckInDir_WalkCallbackError(t *testing.T) {
	t.Parallel()

	callbackErr := fmt.Errorf("simulated callback error")
	stubWalkDirFn := func(root string, fn fs.WalkDirFunc) error {
		// Invoke the walker callback with an error to trigger the walkFileErr != nil branch.
		return fn(root, nil, callbackErr)
	}

	rootDir := buildSuffixRoot(t)
	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := checkInDir(logger, rootDir, os.ReadFile, stubWalkDirFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "simulated callback error")
}

// -----------------------------------------------------------------------
// Check (integration + seam tests)
// -----------------------------------------------------------------------

func TestCheck_Integration(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-test-file-suffix-structure")
	err := Check(logger)

	require.NoError(t, err, "test-file-suffix-structure should pass on real project files")
}

func TestCheck_ProjectRootError(t *testing.T) {
	t.Parallel()

	stubGetwdFn := func() (string, error) {
		return "", fmt.Errorf("simulated root error")
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	rootDir, err := findTestFileSuffixProjectRoot(stubGetwdFn)
	if err != nil {
		require.Contains(t, err.Error(), "simulated root error")

		return
	}

	err = CheckInDir(logger, rootDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "simulated root error")
}

func TestFindTestFileSuffixProjectRoot_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		getwdFn     func() (string, error)
		errContains string
	}{
		{
			name: "getwd error",
			getwdFn: func() (string, error) {
				return "", fmt.Errorf("simulated getwd error")
			},
			errContains: "simulated getwd error",
		},
		{
			name: "go.mod not found",
			getwdFn: func() (string, error) {
				root := filepath.VolumeName(t.TempDir())
				if root == "" {
					root = "/"
				} else {
					root += "\\"
				}

				return root, nil
			},
			errContains: "go.mod not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := findTestFileSuffixProjectRoot(tc.getwdFn)

			require.Error(t, err)
			require.Empty(t, result)
			require.Contains(t, err.Error(), tc.errContains)
		})
	}
}

// -----------------------------------------------------------------------
// isExcludedFromContentRules
// -----------------------------------------------------------------------

func TestIsExcludedFromContentRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		excluded bool
	}{
		{"format_gotest test file", "internal/apps/tools/cicd_lint/format_gotest/foo_test.go", true},
		{"lint_fitness test file", "internal/apps/tools/cicd_lint/lint_fitness/foo_test.go", true},
		{"lint_gotest test file", "internal/apps/tools/cicd_lint/lint_gotest/foo_test.go", true},
		{"regular application test", "internal/shared/hkdf/hkdf_bench_test.go", false},
		{"windows backslash path format_gotest", "internal\\apps\\tools\\cicd_lint\\format_gotest\\bar_test.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isExcludedFromContentRules(tt.path)

			require.Equal(t, tt.excluded, result)
		})
	}
}
