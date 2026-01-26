// Copyright (c) 2025 Justin Cranford

package format_go

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"

	"github.com/stretchr/testify/require"
)

// TestEnforceTimeNowUTC_BasicReplacement tests basic time.Now()  time.Now().UTC() replacement.
func TestEnforceTimeNowUTC_BasicReplacement(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.go")

	originalContent := `package main

import "time"

func main() {
now := time.Now()
println(now)
}
`

	err := os.WriteFile(testFile, []byte(originalContent), 0o600)
	require.NoError(t, err)

	// Run enforcement.
	replacements, err := processGoFileForTimeNowUTC(testFile)
	require.NoError(t, err)
	require.Equal(t, 1, replacements, "Should replace one time.Now() call")

	// Read modified content.
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)

	modifiedStr := string(modifiedContent)

	require.Contains(t, modifiedStr, "time.Now().UTC()", "Should contain time.Now().UTC()")
	require.NotContains(t, modifiedStr, "time.Now()\n", "Should not contain bare time.Now()")
}

// TestEnforceTimeNowUTC_AlreadyCorrect tests that time.Now().UTC() is not modified.
func TestEnforceTimeNowUTC_AlreadyCorrect(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.go")

	originalContent := `package main

import "time"

func main() {
now := time.Now().UTC()
println(now)
}
`

	err := os.WriteFile(testFile, []byte(originalContent), 0o600)
	require.NoError(t, err)

	// Run enforcement.
	replacements, err := processGoFileForTimeNowUTC(testFile)
	require.NoError(t, err)
	require.Equal(t, 0, replacements, "Should not modify already correct code")

	// Read content to verify unchanged.
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)

	require.Equal(t, originalContent, string(modifiedContent), "Content should be unchanged")
}

// TestEnforceTimeNowUTC_ChainedMethodCalls tests time.Now().Add(duration)  time.Now().UTC().Add(duration).
func TestEnforceTimeNowUTC_ChainedMethodCalls(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.go")

	originalContent := `package main

import "time"

func main() {
later := time.Now().Add(1 * time.Hour)
println(later)
}
`

	err := os.WriteFile(testFile, []byte(originalContent), 0o600)
	require.NoError(t, err)

	// Run enforcement.
	replacements, err := processGoFileForTimeNowUTC(testFile)
	require.NoError(t, err)
	require.Equal(t, 1, replacements, "Should replace time.Now() in chained call")

	// Read modified content.
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)

	modifiedStr := string(modifiedContent)

	require.Contains(t, modifiedStr, "time.Now().UTC().Add(1 * time.Hour)", "Should have UTC inserted before Add")
}

// TestEnforceTimeNowUTC_VariableAssignment tests t := time.Now()  t := time.Now().UTC().
func TestEnforceTimeNowUTC_VariableAssignment(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.go")

	originalContent := `package main

import "time"

func main() {
t := time.Now()
later := t.Add(1 * time.Hour)
println(later)
}
`

	err := os.WriteFile(testFile, []byte(originalContent), 0o600)
	require.NoError(t, err)

	// Run enforcement.
	replacements, err := processGoFileForTimeNowUTC(testFile)
	require.NoError(t, err)
	require.Equal(t, 1, replacements, "Should replace time.Now() in assignment")

	// Read modified content.
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)

	modifiedStr := string(modifiedContent)

	require.Contains(t, modifiedStr, "t := time.Now().UTC()", "Should have UTC in assignment")
}

// TestEnforceTimeNowUTC_SelfExclusion verifies format_go package files are excluded.
func TestEnforceTimeNowUTC_SelfExclusion(t *testing.T) {
	t.Parallel()

	// Process this test file itself (should be excluded).
	replacements, err := processGoFileForTimeNowUTC("enforce_time_now_utc_test.go")
	require.NoError(t, err)
	require.Equal(t, 0, replacements, "format_go package files should be excluded from enforcement")

	// Process the enforcement file itself (should be excluded).
	replacements, err = processGoFileForTimeNowUTC("enforce_time_now_utc.go")
	require.NoError(t, err)
	require.Equal(t, 0, replacements, "format_go package files should be excluded from enforcement")
}

// TestEnforceTimeNowUTC_Integration tests full enforcement flow.
func TestEnforceTimeNowUTC_Integration(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create test file with multiple time.Now() calls.
	testFile := filepath.Join(tmpDir, "test.go")

	originalContent := `package main

import "time"

func main() {
	now := time.Now()
	later := time.Now().Add(1 * time.Hour)
	alreadyCorrect := time.Now().UTC()
}
`

	err := os.WriteFile(testFile, []byte(originalContent), 0o600)
	require.NoError(t, err)

	// Create file map.
	filesByExtension := map[string][]string{
		"go": {testFile},
	}

	// Run enforcement via public API.
	logger := cryptoutilCmdCicdCommon.NewLogger("test-enforce-time-now-utc")
	err = enforceTimeNowUTC(logger, filesByExtension)

	// Expect error because files were modified.
	require.Error(t, err, "Should return error when files modified")
	require.Contains(t, err.Error(), "modified 1 files", "Error should indicate 1 file modified")

	// Read modified content.
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)

	modifiedStr := string(modifiedContent)

	// Verify all time.Now() calls have .UTC() except the one that already had it.
	require.Equal(t, 3, strings.Count(modifiedStr, "time.Now().UTC()"), "Should have 3 time.Now().UTC() calls")
	require.NotContains(t, modifiedStr, "time.Now()\n", "Should not have bare time.Now()")
	require.NotContains(t, modifiedStr, "time.Now().Add", "Should not have time.Now().Add without UTC")
}

// TestEnforceTimeNowUTC_NoModificationsNeeded tests when all files already have .UTC().
func TestEnforceTimeNowUTC_NoModificationsNeeded(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create test file that already has all .UTC() calls.
	testFile := filepath.Join(tmpDir, "test.go")

	correctContent := `package main

import "time"

func main() {
	now := time.Now().UTC()
	later := time.Now().UTC().Add(1 * time.Hour)
}
`

	err := os.WriteFile(testFile, []byte(correctContent), 0o600)
	require.NoError(t, err)

	// Create file map.
	filesByExtension := map[string][]string{
		"go": {testFile},
	}

	// Run enforcement via public API.
	logger := cryptoutilCmdCicdCommon.NewLogger("test-enforce-time-now-utc")
	err = enforceTimeNowUTC(logger, filesByExtension)

	// Expect no error because no files were modified.
	require.NoError(t, err, "Should not return error when no modifications needed")

	// Read content to verify it wasn't changed.
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)

	// Content should be identical.
	require.Equal(t, correctContent, string(modifiedContent), "Content should not be modified")
}

// TestEnforceTimeNowUTC_InvalidGoFile tests error handling for malformed Go files.
func TestEnforceTimeNowUTC_InvalidGoFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create invalid Go file.
	testFile := filepath.Join(tmpDir, "invalid.go")

	invalidContent := `package main

this is not valid Go code!
`

	err := os.WriteFile(testFile, []byte(invalidContent), 0o600)
	require.NoError(t, err)

	// Create file map.
	filesByExtension := map[string][]string{
		"go": {testFile},
	}

	// Run enforcement via public API.
	logger := cryptoutilCmdCicdCommon.NewLogger("test-enforce-time-now-utc")
	err = enforceTimeNowUTC(logger, filesByExtension)

	// Should not error even if file parsing fails (errors are logged and skipped).
	require.NoError(t, err, "Should not error when file parsing fails")
}
