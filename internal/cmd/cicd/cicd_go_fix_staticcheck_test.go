// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"cryptoutil/internal/cmd/cicd/common"
)

func TestGoFixStaticcheckErrorStrings_NoGoFiles(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("TestGoFixStaticcheckErrorStrings_NoGoFiles")
	files := []string{"README.md", "config.yml"}

	err := goFixStaticcheckErrorStrings(logger, files)
	require.NoError(t, err, "Should not error when no Go files present")
}

func TestGoFixStaticcheckErrorStrings_NoErrors(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// File with properly formatted error strings (all lowercase)
	content := `package test

import "fmt"

func example() error {
	return fmt.Errorf("this is a proper error message")
}
`
	err := os.WriteFile(testFile, []byte(content), 0o644)
	require.NoError(t, err)

	logger := common.NewLogger("TestGoFixStaticcheckErrorStrings_NoErrors")
	err = goFixStaticcheckErrorStrings(logger, []string{testFile})
	require.NoError(t, err, "Should not error when no fixes needed")

	// Verify file unchanged
	resultContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Equal(t, content, string(resultContent), "File should be unchanged")
}

func TestGoFixStaticcheckErrorStrings_BasicFixes(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// File with capitalized error strings
	originalContent := `package test

import (
	"errors"
	"fmt"
)

func example() error {
	if true {
		return fmt.Errorf("Missing openapi spec")
	}
	return errors.New("Invalid configuration")
}
`

	expectedContent := `package test

import (
	"errors"
	"fmt"
)

func example() error {
	if true {
		return fmt.Errorf("missing openapi spec")
	}
	return errors.New("invalid configuration")
}
`

	err := os.WriteFile(testFile, []byte(originalContent), 0o644)
	require.NoError(t, err)

	logger := common.NewLogger("TestGoFixStaticcheckErrorStrings_BasicFixes")
	err = goFixStaticcheckErrorStrings(logger, []string{testFile})
	require.Error(t, err, "Should return error when fixes are made")
	require.Contains(t, err.Error(), "fixed 2 error strings", "Should report number of fixes")

	// Verify file was fixed
	resultContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Equal(t, expectedContent, string(resultContent), "Error strings should be lowercased")
}

func TestGoFixStaticcheckErrorStrings_PreservesAcronyms(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// File with acronyms that should be preserved
	content := `package test

import "fmt"

func example() error {
	return fmt.Errorf("HTTP request failed")
}

func example2() error {
	return fmt.Errorf("JSON parsing error")
}

func example3() error {
	return fmt.Errorf("UUID generation failed")
}

func example4() error {
	return fmt.Errorf("UUIDs can't be nil")
}
`

	err := os.WriteFile(testFile, []byte(content), 0o644)
	require.NoError(t, err)

	logger := common.NewLogger("TestGoFixStaticcheckErrorStrings_PreservesAcronyms")
	err = goFixStaticcheckErrorStrings(logger, []string{testFile})
	require.NoError(t, err, "Should not modify error strings starting with acronyms")

	// Verify file unchanged
	resultContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Equal(t, content, string(resultContent), "Acronyms should be preserved")
}

func TestGoFixStaticcheckErrorStrings_MixedCase(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	originalContent := `package test

import "fmt"

func example() error {
	// Should fix this
	err1 := fmt.Errorf("Failed to connect")

	// Should NOT fix this (acronym)
	err2 := fmt.Errorf("HTTP connection failed")

	// Should fix this
	err3 := fmt.Errorf("Connection timeout occurred")

	// Should NOT fix this (already lowercase)
	err4 := fmt.Errorf("operation failed")

	return err1
}
`

	expectedContent := `package test

import "fmt"

func example() error {
	// Should fix this
	err1 := fmt.Errorf("failed to connect")

	// Should NOT fix this (acronym)
	err2 := fmt.Errorf("HTTP connection failed")

	// Should fix this
	err3 := fmt.Errorf("connection timeout occurred")

	// Should NOT fix this (already lowercase)
	err4 := fmt.Errorf("operation failed")

	return err1
}
`

	err := os.WriteFile(testFile, []byte(originalContent), 0o644)
	require.NoError(t, err)

	logger := common.NewLogger("TestGoFixStaticcheckErrorStrings_MixedCase")
	err = goFixStaticcheckErrorStrings(logger, []string{testFile})
	require.Error(t, err, "Should return error when fixes are made")
	require.Contains(t, err.Error(), "fixed 2 error strings", "Should fix only non-acronym capitalized strings")

	resultContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Equal(t, expectedContent, string(resultContent))
}

func TestGoFixStaticcheckErrorStrings_MultipleFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.go")
	file2 := filepath.Join(tmpDir, "file2.go")
	file3 := filepath.Join(tmpDir, "file3.go")

	// File 1: Has errors to fix
	content1 := `package test
import "fmt"
func f1() error { return fmt.Errorf("Error occurred") }
`
	expected1 := `package test
import "fmt"
func f1() error { return fmt.Errorf("error occurred") }
`

	// File 2: No errors to fix
	content2 := `package test
import "fmt"
func f2() error { return fmt.Errorf("already lowercase") }
`

	// File 3: Has errors to fix
	content3 := `package test
import "errors"
func f3() error { return errors.New("Something went wrong") }
`
	expected3 := `package test
import "errors"
func f3() error { return errors.New("something went wrong") }
`

	err := os.WriteFile(file1, []byte(content1), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte(content2), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(file3, []byte(content3), 0o644)
	require.NoError(t, err)

	logger := common.NewLogger("TestGoFixStaticcheckErrorStrings_MultipleFiles")
	err = goFixStaticcheckErrorStrings(logger, []string{file1, file2, file3})
	require.Error(t, err, "Should return error when fixes are made")
	require.Contains(t, err.Error(), "fixed 2 error strings", "Should fix errors in 2 files")

	// Verify file1 was fixed
	result1, err := os.ReadFile(file1)
	require.NoError(t, err)
	require.Equal(t, expected1, string(result1))

	// Verify file2 was unchanged
	result2, err := os.ReadFile(file2)
	require.NoError(t, err)
	require.Equal(t, content2, string(result2))

	// Verify file3 was fixed
	result3, err := os.ReadFile(file3)
	require.NoError(t, err)
	require.Equal(t, expected3, string(result3))
}

func TestGoFixStaticcheckErrorStrings_ExcludesGenerated(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create subdirectories
	apiClientDir := filepath.Join(tmpDir, "api", "client")
	err := os.MkdirAll(apiClientDir, 0o755)
	require.NoError(t, err)

	// Generated file that should be skipped
	genFile := filepath.Join(tmpDir, "test_gen.go")
	pbFile := filepath.Join(tmpDir, "test.pb.go")
	apiFile := filepath.Join(apiClientDir, "client.go")

	genContent := `package test
import "fmt"
func gen() error { return fmt.Errorf("generated error") }
`

	err = os.WriteFile(genFile, []byte(genContent), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(pbFile, []byte(genContent), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(apiFile, []byte(genContent), 0o644)
	require.NoError(t, err)

	logger := common.NewLogger("TestGoFixStaticcheckErrorStrings_ExcludesGenerated")
	err = goFixStaticcheckErrorStrings(logger, []string{genFile, pbFile, apiFile})
	require.NoError(t, err, "Should not process generated files")

	// Verify files were not modified
	for _, file := range []string{genFile, pbFile, apiFile} {
		content, err := os.ReadFile(file)
		require.NoError(t, err)
		require.Equal(t, genContent, string(content), "Generated files should not be modified")
	}
}

func TestLowercaseFirst(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty string", "", ""},
		{"Single char uppercase", "A", "a"},
		{"Single char lowercase", "a", "a"},
		{"Multiple words", "Failed to connect", "failed to connect"},
		{"Already lowercase", "already lowercase", "already lowercase"},
		{"Unicode", "Über cool", "über cool"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := lowercaseFirst(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestFilterGoFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Mix of files",
			input:    []string{"main.go", "README.md", "test.go", "config.yml"},
			expected: []string{"main.go", "test.go"},
		},
		{
			name:     "Excludes generated",
			input:    []string{"main.go", "generated_gen.go", "proto.pb.go"},
			expected: []string{"main.go"},
		},
		{
			name:     "Excludes vendor",
			input:    []string{"main.go", "vendor/pkg/lib.go", "internal/app.go"},
			expected: []string{"main.go", "internal/app.go"},
		},
		{
			name:     "Excludes API",
			input:    []string{"main.go", "api/client/client.go", "api/model/model.go", "api/server/server.go"},
			expected: []string{"main.go"},
		},
		{
			name:     "No Go files",
			input:    []string{"README.md", "config.yml"},
			expected: nil,
		},
		{
			name:     "Windows paths",
			input:    []string{"main.go", "vendor\\pkg\\lib.go", "api\\client\\client.go"},
			expected: []string{"main.go"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := filterGoFiles(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestGoFixAll_NoFiles(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("TestGoFixAll_NoFiles")
	files := []string{"README.md", "config.yml"}

	err := goFixAll(logger, files)
	require.NoError(t, err, "Should not error when no Go files present")
}

func TestGoFixAll_NoChangesNeeded(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// File with properly formatted code (no fixes needed)
	content := `package test

import "fmt"

func example() error {
	return fmt.Errorf("already lowercase error")
}
`
	err := os.WriteFile(testFile, []byte(content), 0o644)
	require.NoError(t, err)

	logger := common.NewLogger("TestGoFixAll_NoChangesNeeded")
	err = goFixAll(logger, []string{testFile})
	require.NoError(t, err, "Should not error when no fixes needed")

	// Verify file unchanged
	resultContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Equal(t, content, string(resultContent), "File should be unchanged")
}

func TestGoFixAll_AppliesFixes(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_test.go") // Use _test.go suffix for thelper

	originalContent := `package test

import "fmt"
import "testing"

func example() error {
	return fmt.Errorf("Error occurred") // Uppercase error string - will be fixed by staticcheck
}

func setupTest(t *testing.T) {
	// Helper function without t.Helper() - will be fixed by thelper
}
`

	err := os.WriteFile(testFile, []byte(originalContent), 0o644)
	require.NoError(t, err)

	logger := common.NewLogger("TestGoFixAll_AppliesFixes")
	err = goFixAll(logger, []string{testFile})
	require.Error(t, err, "Should return error when fixes are made")

	// Verify file was modified
	resultContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	// Should contain t.Helper() (added by thelper)
	require.Contains(t, string(resultContent), ".Helper()", "Should have added Helper() call")
	// Should have lowercase error string (fixed by staticcheck)
	require.Contains(t, string(resultContent), `"error occurred"`, "Should have lowercased error string")
}
