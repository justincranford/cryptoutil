// Copyright (c) 2025 Justin Cranford

package lint_pythontest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilLintPythonTest "cryptoutil/internal/apps/tools/cicd_lint/lint_pythontest"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestLint_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := cryptoutilLintPythonTest.Lint(logger, map[string][]string{})

	require.NoError(t, err, "Lint should succeed with no files")
}

func TestLint_NoPythonFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go":  {"main.go"},
		"yml": {"config.yml"},
	}

	err := cryptoutilLintPythonTest.Lint(logger, filesByExtension)

	require.NoError(t, err, "Lint should succeed with no Python files")
}

func TestLint_NoPythonTestFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	utilFile := filepath.Join(tmpDir, "utils.py")
	content := `class HelperUtil(unittest.TestCase):
    def test_something(self):
        self.assertEqual(1, 1)
`
	err := os.WriteFile(utilFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"py": {utilFile},
	}

	err = cryptoutilLintPythonTest.Lint(logger, filesByExtension)

	require.NoError(t, err, "Lint should skip non-test Python files (not test_*.py or *_test.py)")
}

func TestLint_ValidPytestFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_calculator.py")
	content := `import pytest

@pytest.mark.parametrize("x,expected", [(1, 1), (2, 4)])
def test_square(x, expected):
    assert x * x == expected
`
	err := os.WriteFile(testFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"py": {testFile},
	}

	err = cryptoutilLintPythonTest.Lint(logger, filesByExtension)

	require.NoError(t, err, "Lint should succeed for pytest-style test files")
}

func TestLint_UnittestTestCase_PrefixFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_legacy.py")
	content := `import unittest

class LegacyTest(unittest.TestCase):
    def test_something(self):
        pass
`
	err := os.WriteFile(testFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"py": {testFile},
	}

	err = cryptoutilLintPythonTest.Lint(logger, filesByExtension)

	require.Error(t, err, "Lint should fail for unittest.TestCase usage")
	require.ErrorContains(t, err, "unittest antipattern violations")
}

func TestLint_UnittestTestCase_SuffixFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "calculator_test.py")
	content := `from unittest import TestCase

class CalculatorTest(TestCase):
    def test_add(self):
        self.assertEqual(1 + 1, 2)
`
	err := os.WriteFile(testFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"py": {testFile},
	}

	err = cryptoutilLintPythonTest.Lint(logger, filesByExtension)

	require.Error(t, err, "Lint should fail for unittest.TestCase usage in *_test.py file")
	require.ErrorContains(t, err, "unittest antipattern violations")
}

func TestLint_SelfAssertMethod(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_asserts.py")
	content := `def test_value():
    self.assertEqual(1, 1)
    self.assertTrue(True)
`
	err := os.WriteFile(testFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"py": {testFile},
	}

	err = cryptoutilLintPythonTest.Lint(logger, filesByExtension)

	require.Error(t, err, "Lint should fail for self.assert* usage in test files")
	require.ErrorContains(t, err, "unittest antipattern violations")
}

func TestCheckUnittestAntipattern_FileNotFound(t *testing.T) {
	t.Parallel()

	issues := cryptoutilLintPythonTest.CheckUnittestAntipattern("/nonexistent/path/test_file.py")

	require.Len(t, issues, 1)
	require.Contains(t, issues[0], "Error reading file")
}

func TestCheckUnittestAntipattern_MultipleViolations(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_multi.py")
	content := `import unittest

class MultiTest(unittest.TestCase):
    def test_a(self):
        self.assertEqual(1, 1)
    def test_b(self):
        self.assertTrue(True)
`
	err := os.WriteFile(testFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	issues := cryptoutilLintPythonTest.CheckUnittestAntipattern(testFile)

	require.GreaterOrEqual(t, len(issues), 3, "Should detect TestCase and multiple self.assert* calls")
}

func TestFilterPythonTestFiles_MixedExtensions(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	testFile1 := filepath.Join(tmpDir, "test_foo.py")
	testFile2 := filepath.Join(tmpDir, "bar_test.py")
	nonTestFile := filepath.Join(tmpDir, "helper.py")

	for _, f := range []string{testFile1, testFile2, nonTestFile} {
		err := os.WriteFile(f, []byte("# placeholder"), cryptoutilSharedMagic.CacheFilePermissions)
		require.NoError(t, err)
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"py": {testFile1, testFile2, nonTestFile},
	}

	err := cryptoutilLintPythonTest.Lint(logger, filesByExtension)

	require.NoError(t, err, "Valid test files with no violations should pass")
}
