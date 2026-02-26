// Copyright (c) 2025 Justin Cranford

package enforce_time_now_utc

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
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

	err := os.WriteFile(testFile, []byte(originalContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Run enforcement.
	replacements, err := ProcessGoFileForTimeNowUTC(testFile)
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

	err := os.WriteFile(testFile, []byte(originalContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Run enforcement.
	replacements, err := ProcessGoFileForTimeNowUTC(testFile)
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

	err := os.WriteFile(testFile, []byte(originalContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Run enforcement.
	replacements, err := ProcessGoFileForTimeNowUTC(testFile)
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

	err := os.WriteFile(testFile, []byte(originalContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Run enforcement.
	replacements, err := ProcessGoFileForTimeNowUTC(testFile)
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
	replacements, err := ProcessGoFileForTimeNowUTC("enforce_time_now_utc_test.go")
	require.NoError(t, err)
	require.Equal(t, 0, replacements, "format_go package files should be excluded from enforcement")

	// Process the enforcement file itself (should be excluded).
	replacements, err = ProcessGoFileForTimeNowUTC("enforce_time_now_utc.go")
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

	err := os.WriteFile(testFile, []byte(originalContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Create file map.
	filesByExtension := map[string][]string{
		"go": {testFile},
	}

	// Run enforcement via public API.
	logger := cryptoutilCmdCicdCommon.NewLogger("test-enforce-time-now-utc")
	err = Enforce(logger, filesByExtension)

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

	err := os.WriteFile(testFile, []byte(correctContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Create file map.
	filesByExtension := map[string][]string{
		"go": {testFile},
	}

	// Run enforcement via public API.
	logger := cryptoutilCmdCicdCommon.NewLogger("test-enforce-time-now-utc")
	err = Enforce(logger, filesByExtension)

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

	err := os.WriteFile(testFile, []byte(invalidContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Create file map.
	filesByExtension := map[string][]string{
		"go": {testFile},
	}

	// Run enforcement via public API.
	logger := cryptoutilCmdCicdCommon.NewLogger("test-enforce-time-now-utc")
	err = Enforce(logger, filesByExtension)

	// Should not error even if file parsing fails (errors are logged and skipped).
	require.NoError(t, err, "Should not error when file parsing fails")
}

// TestProcessGoFileForTimeNowUTC_ReadError tests the read file error path.
func TestProcessGoFileForTimeNowUTC_ReadError(t *testing.T) {
	t.Parallel()

	replacements, err := ProcessGoFileForTimeNowUTC("/nonexistent/path/to/test.go")
	require.Error(t, err, "processGoFileForTimeNowUTC should fail for nonexistent file")
	require.Equal(t, 0, replacements)
	require.Contains(t, err.Error(), "failed to read file")
}

// testGoContentUTCOnVariable is Go content with t.UTC() where t is a time.Time variable.
// This triggers the first-pass innerCall type assertion failure (Ident receiver, not CallExpr).
const testGoContentUTCOnVariable = `package main

import "time"

func main() {
var t time.Time
_ = t.UTC()
}
`

// TestProcessGoFileForTimeNowUTC_UTCOnVariable tests the first-pass type assertion failure
// for x.UTC() where x is an identifier (not a call expression).
func TestProcessGoFileForTimeNowUTC_UTCOnVariable(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	err := os.WriteFile(testFile, []byte(testGoContentUTCOnVariable), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	replacements, err := ProcessGoFileForTimeNowUTC(testFile)
	require.NoError(t, err, "Should not error on file with t.UTC() on variable")
	require.Equal(t, 0, replacements, "No replacements needed (already correct or not time.Now())")
}

func TestEnforce_EmptyFileMap(t *testing.T) {
t.Parallel()

logger := cryptoutilCmdCicdCommon.NewLogger("test-enforce-empty")
// Empty map means no Go files -> returns nil immediately.
err := Enforce(logger, map[string][]string{})
require.NoError(t, err, "Empty file map should return nil without error")
}

func TestProcessGoFileForTimeNowUTC_SelfModificationPath(t *testing.T) {
t.Parallel()

// Create a file at a path containing "internal/cmd/cicd/format_go".
// On Linux the C:/temp and R:/temp conditions are always true (not Windows paths),
// so the self-modification check fires and returns 0, nil.
tmpDir := t.TempDir()
fakeFormatGoDir := filepath.Join(tmpDir, "internal", "cmd", "cicd", "format_go")
require.NoError(t, os.MkdirAll(fakeFormatGoDir, 0o700))

goFile := filepath.Join(fakeFormatGoDir, "dummy.go")
require.NoError(t, os.WriteFile(goFile, []byte("package format_go\n\nimport \"time\"\n\nvar t = time.Now()\n"), cryptoutilSharedMagic.CacheFilePermissions))

replacements, err := ProcessGoFileForTimeNowUTC(goFile)
require.NoError(t, err)
require.Equal(t, 0, replacements, "Self-modification path should be silently skipped")
}

func TestProcessGoFileForTimeNowUTC_InnerCallFuncNotSelector(t *testing.T) {
t.Parallel()

// gettime().UTC() - innerCall.Fun is *ast.Ident, not *ast.SelectorExpr -> returns true
tmpDir := t.TempDir()
goFile := filepath.Join(tmpDir, "a.go")
content := `package foo

import "time"

func gettime() time.Time { return time.Now().UTC() }

func f() { _ = gettime().UTC() }
`
require.NoError(t, os.WriteFile(goFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

replacements, err := ProcessGoFileForTimeNowUTC(goFile)
require.NoError(t, err)
require.Equal(t, 0, replacements, "gettime().UTC() should not be modified")
}

func TestProcessGoFileForTimeNowUTC_InnerSelNotNow(t *testing.T) {
t.Parallel()

// time.Date(...).UTC() - innerSel.Sel.Name = "Date" != "Now" -> returns true
tmpDir := t.TempDir()
goFile := filepath.Join(tmpDir, "b.go")
content := `package foo

import "time"

func f() { _ = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UTC() }
`
require.NoError(t, os.WriteFile(goFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

replacements, err := ProcessGoFileForTimeNowUTC(goFile)
require.NoError(t, err)
require.Equal(t, 0, replacements, "time.Date().UTC() should not be modified")
}

func TestProcessGoFileForTimeNowUTC_InnerIdentNotTime(t *testing.T) {
t.Parallel()

// myT.Now().UTC() - ident.Name = "myT" != "time" -> returns true in both passes
tmpDir := t.TempDir()
goFile := filepath.Join(tmpDir, "c.go")
content := `package foo

import "time"

type fakeTimer struct{}

func (f fakeTimer) Now() time.Time { return time.Now().UTC() }

func g() {
var ft fakeTimer
_ = ft.Now().UTC()
_ = ft.Now()
}
`
require.NoError(t, os.WriteFile(goFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

replacements, err := ProcessGoFileForTimeNowUTC(goFile)
require.NoError(t, err)
require.Equal(t, 0, replacements, "non-time.Now() calls should not be modified")
}
