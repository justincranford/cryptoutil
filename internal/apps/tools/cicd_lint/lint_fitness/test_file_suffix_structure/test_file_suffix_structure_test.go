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
	require.NoError(t, os.WriteFile(destPath, data, cryptoutilSharedMagic.FilePermissionsDefault))

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
// CheckFiles — suffix rules
// -----------------------------------------------------------------------

func TestCheckFiles_BenchFileHasBenchmark(t *testing.T) {
	t.Parallel()

	rootDir := buildSuffixRoot(t)
	rules, err := LoadSuffixRules(rootDir, os.ReadFile)
	require.NoError(t, err)

	pkgDir := filepath.Join(t.TempDir(), "mypkg")
	f := writeTestFile(t, pkgDir, "foo_bench_test.go",
		"package mypkg\nimport \"testing\"\nfunc BenchmarkFoo(b *testing.B) {}\n")
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err = CheckFiles(logger, []string{f}, rules, os.ReadFile)

	require.NoError(t, err)
}

func TestCheckFiles_BenchFileMissingBenchmark(t *testing.T) {
	t.Parallel()

	rootDir := buildSuffixRoot(t)
	rules, err := LoadSuffixRules(rootDir, os.ReadFile)
	require.NoError(t, err)

	pkgDir := filepath.Join(t.TempDir(), "mypkg")
	f := writeTestFile(t, pkgDir, "foo_bench_test.go",
		"package mypkg\nimport \"testing\"\nfunc TestFoo(t *testing.T) {}\n")
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err = CheckFiles(logger, []string{f}, rules, os.ReadFile)

	require.Error(t, err)
	require.Contains(t, err.Error(), "violation(s)")
}

func TestCheckFiles_BenchFileForbiddenFuzz(t *testing.T) {
	t.Parallel()

	rootDir := buildSuffixRoot(t)
	rules, err := LoadSuffixRules(rootDir, os.ReadFile)
	require.NoError(t, err)

	pkgDir := filepath.Join(t.TempDir(), "mypkg")
	f := writeTestFile(t, pkgDir, "foo_bench_test.go",
		"package mypkg\nimport \"testing\"\nfunc BenchmarkFoo(b *testing.B) {}\nfunc FuzzFoo(f *testing.F) {}\n")
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err = CheckFiles(logger, []string{f}, rules, os.ReadFile)

	require.Error(t, err)
	require.Contains(t, err.Error(), "violation(s)")
}

func TestCheckFiles_FuzzFileHasFuzz(t *testing.T) {
	t.Parallel()

	rootDir := buildSuffixRoot(t)
	rules, err := LoadSuffixRules(rootDir, os.ReadFile)
	require.NoError(t, err)

	pkgDir := filepath.Join(t.TempDir(), "mypkg")
	f := writeTestFile(t, pkgDir, "foo_fuzz_test.go",
		"package mypkg\nimport \"testing\"\nfunc FuzzFoo(f *testing.F) {}\n")
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err = CheckFiles(logger, []string{f}, rules, os.ReadFile)

	require.NoError(t, err)
}

func TestCheckFiles_FuzzFileMissingFuzz(t *testing.T) {
	t.Parallel()

	rootDir := buildSuffixRoot(t)
	rules, err := LoadSuffixRules(rootDir, os.ReadFile)
	require.NoError(t, err)

	pkgDir := filepath.Join(t.TempDir(), "mypkg")
	f := writeTestFile(t, pkgDir, "foo_fuzz_test.go",
		"package mypkg\nimport \"testing\"\nfunc TestFoo(t *testing.T) {}\n")
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err = CheckFiles(logger, []string{f}, rules, os.ReadFile)

	require.Error(t, err)
	require.Contains(t, err.Error(), "violation(s)")
}

func TestCheckFiles_PropertyFileHasBuildTag(t *testing.T) {
	t.Parallel()

	rootDir := buildSuffixRoot(t)
	rules, err := LoadSuffixRules(rootDir, os.ReadFile)
	require.NoError(t, err)

	pkgDir := filepath.Join(t.TempDir(), "mypkg")
	f := writeTestFile(t, pkgDir, "foo_property_test.go",
		"//go:build !fuzz\npackage mypkg\nimport \"testing\"\nfunc TestFooProp(t *testing.T) {}\n")
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err = CheckFiles(logger, []string{f}, rules, os.ReadFile)

	require.NoError(t, err)
}

func TestCheckFiles_PropertyFileMissingBuildTag(t *testing.T) {
	t.Parallel()

	rootDir := buildSuffixRoot(t)
	rules, err := LoadSuffixRules(rootDir, os.ReadFile)
	require.NoError(t, err)

	pkgDir := filepath.Join(t.TempDir(), "mypkg")
	f := writeTestFile(t, pkgDir, "foo_property_test.go",
		"package mypkg\nimport \"testing\"\nfunc TestFooProp(t *testing.T) {}\n")
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err = CheckFiles(logger, []string{f}, rules, os.ReadFile)

	require.Error(t, err)
	require.Contains(t, err.Error(), "violation(s)")
}

func TestCheckFiles_IntegrationFileNoFuzz(t *testing.T) {
	t.Parallel()

	rootDir := buildSuffixRoot(t)
	rules, err := LoadSuffixRules(rootDir, os.ReadFile)
	require.NoError(t, err)

	pkgDir := filepath.Join(t.TempDir(), "mypkg")
	f := writeTestFile(t, pkgDir, "foo_integration_test.go",
		"package mypkg\nimport \"testing\"\nfunc TestFoo(t *testing.T) {}\n")
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err = CheckFiles(logger, []string{f}, rules, os.ReadFile)

	require.NoError(t, err)
}

func TestCheckFiles_IntegrationFileForbiddenFuzz(t *testing.T) {
	t.Parallel()

	rootDir := buildSuffixRoot(t)
	rules, err := LoadSuffixRules(rootDir, os.ReadFile)
	require.NoError(t, err)

	pkgDir := filepath.Join(t.TempDir(), "mypkg")
	f := writeTestFile(t, pkgDir, "foo_integration_test.go",
		"package mypkg\nimport \"testing\"\nfunc FuzzFoo(f *testing.F) {}\n")
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err = CheckFiles(logger, []string{f}, rules, os.ReadFile)

	require.Error(t, err)
	require.Contains(t, err.Error(), "violation(s)")
}

// -----------------------------------------------------------------------
// CheckFiles — content rules
// -----------------------------------------------------------------------

func TestCheckFiles_ContentRule_FuzzInWrongFile(t *testing.T) {
	t.Parallel()

	rootDir := buildSuffixRoot(t)
	rules, err := LoadSuffixRules(rootDir, os.ReadFile)
	require.NoError(t, err)

	pkgDir := filepath.Join(t.TempDir(), "mypkg")
	f := writeTestFile(t, pkgDir, "foo_test.go",
		"package mypkg\nimport \"testing\"\nfunc FuzzFoo(f *testing.F) {}\n")
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err = CheckFiles(logger, []string{f}, rules, os.ReadFile)

	require.Error(t, err)
	require.Contains(t, err.Error(), "violation(s)")
}

func TestCheckFiles_ContentRule_BenchmarkInWrongFile(t *testing.T) {
	t.Parallel()

	rootDir := buildSuffixRoot(t)
	rules, err := LoadSuffixRules(rootDir, os.ReadFile)
	require.NoError(t, err)

	pkgDir := filepath.Join(t.TempDir(), "mypkg")
	f := writeTestFile(t, pkgDir, "foo_test.go",
		"package mypkg\nimport \"testing\"\nfunc BenchmarkFoo(b *testing.B) {}\n")
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err = CheckFiles(logger, []string{f}, rules, os.ReadFile)

	require.Error(t, err)
	require.Contains(t, err.Error(), "violation(s)")
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

func TestCheckFiles_InvalidContentPatternInRules(t *testing.T) {
	t.Parallel()

	rules := &SuffixRules{
		ContentRules: []ContentRule{
			{ContentPattern: "[invalid-regex", RequiredSuffix: "_test.go"},
		},
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	pkgDir := filepath.Join(t.TempDir(), "mypkg")
	f := writeTestFile(t, pkgDir, "foo_test.go", "package mypkg\n")

	err := CheckFiles(logger, []string{f}, rules, os.ReadFile)

	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid content_pattern")
}

func TestCheckFiles_InvalidRequiredPatternInRule(t *testing.T) {
	t.Parallel()

	rules := &SuffixRules{
		SuffixRules: []SuffixRule{
			{Suffix: "_test.go", RequiredContentPatterns: []string{"[bad-regex"}},
		},
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	pkgDir := filepath.Join(t.TempDir(), "mypkg")
	f := writeTestFile(t, pkgDir, "foo_test.go", "package mypkg\n")

	err := CheckFiles(logger, []string{f}, rules, os.ReadFile)

	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid required_content_pattern")
}

func TestCheckFiles_InvalidForbiddenPatternInRule(t *testing.T) {
	t.Parallel()

	rules := &SuffixRules{
		SuffixRules: []SuffixRule{
			{Suffix: "_test.go", ForbiddenContentPatterns: []string{"[bad-regex"}},
		},
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	pkgDir := filepath.Join(t.TempDir(), "mypkg")
	f := writeTestFile(t, pkgDir, "foo_test.go", "package mypkg\n")

	err := CheckFiles(logger, []string{f}, rules, os.ReadFile)

	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid forbidden_content_pattern")
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

func TestFindTestFileSuffixProjectRoot_GetwdError(t *testing.T) {
	t.Parallel()

	stubGetwdFn := func() (string, error) {
		return "", fmt.Errorf("simulated getwd error")
	}

	result, err := findTestFileSuffixProjectRoot(stubGetwdFn)

	require.Error(t, err)
	require.Empty(t, result)
	require.Contains(t, err.Error(), "simulated getwd error")
}

func TestFindTestFileSuffixProjectRoot_GoModNotFound(t *testing.T) {
	t.Parallel()

	stubGetwdFn := func() (string, error) {
		root := filepath.VolumeName(t.TempDir())
		if root == "" {
			root = "/"
		} else {
			root += "\\"
		}

		return root, nil
	}

	result, err := findTestFileSuffixProjectRoot(stubGetwdFn)

	require.Error(t, err)
	require.Empty(t, result)
	require.Contains(t, err.Error(), "go.mod not found")
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
