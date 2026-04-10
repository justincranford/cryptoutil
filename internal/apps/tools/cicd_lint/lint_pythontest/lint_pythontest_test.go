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

func TestLint_HappyPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setupFn func(t *testing.T) map[string][]string
	}{
		{
			name:    "no files",
			setupFn: func(_ *testing.T) map[string][]string { return map[string][]string{} },
		},
		{
			name: "no python files",
			setupFn: func(_ *testing.T) map[string][]string {
				return map[string][]string{"go": {"main.go"}, "yml": {"config.yml"}}
			},
		},
		{
			name: "no python test files",
			setupFn: func(t *testing.T) map[string][]string {
				t.Helper()
				utilFile := filepath.Join(t.TempDir(), "utils.py")
				content := "class HelperUtil(unittest.TestCase):\n    def test_something(self):\n        self.assertEqual(1, 1)\n"
				require.NoError(t, os.WriteFile(utilFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

				return map[string][]string{"py": {utilFile}}
			},
		},
		{
			name: "valid pytest file",
			setupFn: func(t *testing.T) map[string][]string {
				t.Helper()
				testFile := filepath.Join(t.TempDir(), "test_calculator.py")
				content := "import pytest\n\n@pytest.mark.parametrize(\"x,expected\", [(1, 1), (2, 4)])\ndef test_square(x, expected):\n    assert x * x == expected\n"
				require.NoError(t, os.WriteFile(testFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

				return map[string][]string{"py": {testFile}}
			},
		},
		{
			name: "mixed extensions with valid test files",
			setupFn: func(t *testing.T) map[string][]string {
				t.Helper()
				tmpDir := t.TempDir()
				testFile1 := filepath.Join(tmpDir, "test_foo.py")
				testFile2 := filepath.Join(tmpDir, "bar_test.py")
				nonTestFile := filepath.Join(tmpDir, "helper.py")

				for _, f := range []string{testFile1, testFile2, nonTestFile} {
					require.NoError(t, os.WriteFile(f, []byte("# placeholder"), cryptoutilSharedMagic.CacheFilePermissions))
				}

				return map[string][]string{"py": {testFile1, testFile2, nonTestFile}}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			filesByExtension := tc.setupFn(t)

			err := cryptoutilLintPythonTest.Lint(logger, filesByExtension)

			require.NoError(t, err)
		})
	}
}

func TestLint_UnittestViolations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "unittest TestCase in test_ prefix file",
			filename: "test_legacy.py",
			content:  "import unittest\n\nclass LegacyTest(unittest.TestCase):\n    def test_something(self):\n        pass\n",
		},
		{
			name:     "unittest TestCase in _test suffix file",
			filename: "calculator_test.py",
			content:  "from unittest import TestCase\n\nclass CalculatorTest(TestCase):\n    def test_add(self):\n        self.assertEqual(1 + 1, 2)\n",
		},
		{
			name:     "self assert method usage",
			filename: "test_asserts.py",
			content:  "def test_value():\n    self.assertEqual(1, 1)\n    self.assertTrue(True)\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testFile := filepath.Join(t.TempDir(), tc.filename)
			require.NoError(t, os.WriteFile(testFile, []byte(tc.content), cryptoutilSharedMagic.CacheFilePermissions))

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			filesByExtension := map[string][]string{"py": {testFile}}

			err := cryptoutilLintPythonTest.Lint(logger, filesByExtension)

			require.Error(t, err)
			require.ErrorContains(t, err, "unittest antipattern violations")
		})
	}
}

func TestCheckUnittestAntipattern_Variants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setupFn  func(t *testing.T) string
		minCount int
		contains string
	}{
		{
			name:     "file not found",
			setupFn:  func(_ *testing.T) string { return "/nonexistent/path/test_file.py" },
			minCount: 1,
			contains: "Error reading file",
		},
		{
			name: "multiple violations",
			setupFn: func(t *testing.T) string {
				t.Helper()
				testFile := filepath.Join(t.TempDir(), "test_multi.py")
				content := "import unittest\n\nclass MultiTest(unittest.TestCase):\n    def test_a(self):\n        self.assertEqual(1, 1)\n    def test_b(self):\n        self.assertTrue(True)\n"
				require.NoError(t, os.WriteFile(testFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

				return testFile
			},
			minCount: 3,
			contains: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			filePath := tc.setupFn(t)

			issues := cryptoutilLintPythonTest.CheckUnittestAntipattern(filePath)

			require.GreaterOrEqual(t, len(issues), tc.minCount)

			if tc.contains != "" {
				require.Contains(t, issues[0], tc.contains)
			}
		})
	}
}
