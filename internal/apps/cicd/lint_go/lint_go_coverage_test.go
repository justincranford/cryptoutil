// Copyright (c) 2025 Justin Cranford

package lint_go

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

// TestCheckMainGoFile_ReadError verifies that checkMainGoFile returns error when file cannot be read.
func TestCheckMainGoFile_ReadError(t *testing.T) {
	t.Parallel()

	err := checkMainGoFile("/nonexistent/path/main.go")
	require.Error(t, err, "Should return error for nonexistent file")
	require.Contains(t, err.Error(), "failed to read file")
}

// TestCheckMainGoFile_PatternMismatch verifies that checkMainGoFile returns error when
// the file does not match the required main() pattern.
func TestCheckMainGoFile_PatternMismatch(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mainFile := filepath.Join(tmpDir, "main.go")
	content := []byte(`package main

func main() {
	// does not use os.Exit with the cryptoutil pattern
}
`)

	err := os.WriteFile(mainFile, content, 0o600)
	require.NoError(t, err)

	err = checkMainGoFile(mainFile)
	require.Error(t, err, "Should return error for non-matching pattern")
	require.Contains(t, err.Error(), "does not match required pattern")
}

// TestCheckCmdMainPatternInDir_NoCmdDir verifies that checkCmdMainPatternInDir returns nil
// when there is no cmd/ directory.
func TestCheckCmdMainPatternInDir_NoCmdDir(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// tmpDir has no cmd/ subdirectory.
	err := checkCmdMainPatternInDir(logger, tmpDir)
	require.NoError(t, err, "Should succeed when no cmd/ directory found")
}

// TestCheckCmdMainPatternInDir_WithViolation verifies that checkCmdMainPatternInDir
// returns error when main.go does not match the required pattern.
func TestCheckCmdMainPatternInDir_WithViolation(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()
	cmdDir := filepath.Join(tmpDir, "cmd", "myapp")

	err := os.MkdirAll(cmdDir, 0o755)
	require.NoError(t, err)

	// Write a main.go that does NOT match the required pattern.
	mainFile := filepath.Join(cmdDir, "main.go")
	content := []byte(`package main

func main() {
	// incorrect pattern - no os.Exit
}
`)

	err = os.WriteFile(mainFile, content, 0o600)
	require.NoError(t, err)

	err = checkCmdMainPatternInDir(logger, tmpDir)
	require.Error(t, err, "Should return error for non-matching main.go")
	require.Contains(t, err.Error(), "cmd/ main() pattern violations")
}

// TestCheckCryptoRandInDir_WithViolation verifies that checkCryptoRandInDir returns
// an error when a Go file uses math/rand.
func TestCheckCryptoRandInDir_WithViolation(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()
	violationFile := filepath.Join(tmpDir, "badrand.go")

	content := []byte(`package foo

import "math/rand"

func getNum() int {
	return rand.Intn(100)
}
`)

	err := os.WriteFile(violationFile, content, 0o600)
	require.NoError(t, err)

	err = checkCryptoRandInDir(logger, tmpDir)
	require.Error(t, err, "Should return error when math/rand is used")
	require.Contains(t, err.Error(), "math/rand violations")
}

// TestCheckCryptoRandInDir_Clean verifies that checkCryptoRandInDir returns nil
// when no math/rand violations are found.
func TestCheckCryptoRandInDir_Clean(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()
	cleanFile := filepath.Join(tmpDir, "clean.go")

	content := []byte("package foo\n")

	err := os.WriteFile(cleanFile, content, 0o600)
	require.NoError(t, err)

	err = checkCryptoRandInDir(logger, tmpDir)
	require.NoError(t, err, "Should pass when no math/rand is used")
}

// TestCheckInsecureSkipVerifyInDir_WithViolation verifies that checkInsecureSkipVerifyInDir
// returns error when InsecureSkipVerify: true is found in production code.
func TestCheckInsecureSkipVerifyInDir_WithViolation(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()
	violationFile := filepath.Join(tmpDir, "badtls.go")

	content := []byte(`package foo

import "crypto/tls"

func getClient() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
	}
}
`)

	err := os.WriteFile(violationFile, content, 0o600)
	require.NoError(t, err)

	err = checkInsecureSkipVerifyInDir(logger, tmpDir)
	require.Error(t, err, "Should return error when InsecureSkipVerify: true is used")
	require.Contains(t, err.Error(), "InsecureSkipVerify violations")
}

// TestCheckInsecureSkipVerifyInDir_Clean verifies that checkInsecureSkipVerifyInDir
// returns nil when no InsecureSkipVerify violations are found.
func TestCheckInsecureSkipVerifyInDir_Clean(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()
	cleanFile := filepath.Join(tmpDir, "goodtls.go")

	content := []byte("package foo\n")

	err := os.WriteFile(cleanFile, content, 0o600)
	require.NoError(t, err)

	err = checkInsecureSkipVerifyInDir(logger, tmpDir)
	require.NoError(t, err, "Should pass when no InsecureSkipVerify usage found")
}

// TestFindMathRandViolationsInDir_SkipsTestFiles verifies that test files are skipped.
func TestFindMathRandViolationsInDir_SkipsTestFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "pkg_test.go")
	content := []byte(`package foo

import "math/rand"

func TestHelper() int {
	return rand.Intn(100)
}
`)

	err := os.WriteFile(testFile, content, 0o600)
	require.NoError(t, err)

	violations, err := findMathRandViolationsInDir(tmpDir)
	require.NoError(t, err)
	require.Empty(t, violations, "Test files should be skipped for math/rand check")
}

// TestFindMathRandViolationsInDir_SkipsNolintFiles verifies that files with nolint are skipped.
func TestFindMathRandViolationsInDir_SkipsNolintFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	nolintFile := filepath.Join(tmpDir, "pkg.go")
	content := []byte(`package foo

import "math/rand" //nolint:gosec

func get() int {
	return rand.Intn(100)
}
`)

	err := os.WriteFile(nolintFile, content, 0o600)
	require.NoError(t, err)

	violations, err := findMathRandViolationsInDir(tmpDir)
	require.NoError(t, err)
	require.Empty(t, violations, "Files with nolint should be skipped")
}

// TestFindInsecureSkipVerifyViolationsInDir_SkipsTestFiles verifies test files are skipped.
func TestFindInsecureSkipVerifyViolationsInDir_SkipsTestFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "pkg_test.go")
	content := []byte(`package foo

import "crypto/tls"

func TestHelper() *tls.Config {
	return &tls.Config{InsecureSkipVerify: true}
}
`)

	err := os.WriteFile(testFile, content, 0o600)
	require.NoError(t, err)

	violations, err := findInsecureSkipVerifyViolationsInDir(tmpDir)
	require.NoError(t, err)
	require.Empty(t, violations, "Test files should be skipped for InsecureSkipVerify check")
}

// TestCheckCryptoRandInDir_SkipsTestHelperDirs verifies that test helper directories
// are excluded from the math/rand check.
func TestCheckCryptoRandInDir_SkipsTestHelperDirs(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Place violations only inside test helper directories (should be skipped).
	testingDir := filepath.Join(tmpDir, "testing")
	err := os.MkdirAll(testingDir, 0o755)
	require.NoError(t, err)

	violationFile := filepath.Join(testingDir, "helper.go")
	content := []byte(`package testing

import "math/rand"

func get() int { return rand.Intn(100) }
`)

	err = os.WriteFile(violationFile, content, 0o600)
	require.NoError(t, err)

	err = checkCryptoRandInDir(logger, tmpDir)
	require.NoError(t, err, "testing/ directory should be excluded from math/rand check")
}

// TestCheckInsecureSkipVerifyInDir_SkipsTestHelperDirs verifies that test helper
// directories are excluded from the InsecureSkipVerify check.
func TestCheckInsecureSkipVerifyInDir_SkipsTestHelperDirs(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	testingDir := filepath.Join(tmpDir, "testing")
	err := os.MkdirAll(testingDir, 0o755)
	require.NoError(t, err)

	violationFile := filepath.Join(testingDir, "helper.go")
	content := []byte(`package testing

import "crypto/tls"

func Helper() *tls.Config { return &tls.Config{InsecureSkipVerify: true} }
`)

	err = os.WriteFile(violationFile, content, 0o600)
	require.NoError(t, err)

	err = checkInsecureSkipVerifyInDir(logger, tmpDir)
	require.NoError(t, err, "testing/ directory should be excluded")
}

// TestCheckLeftoverCoverageInDir_WithCoverageFiles verifies that coverage files are detected
// and deleted, returning an error to trigger CI awareness.
func TestCheckLeftoverCoverageInDir_WithCoverageFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Create a leftover coverage file.
	covFile := filepath.Join(tmpDir, "coverage.out")
	err := os.WriteFile(covFile, []byte("mode: atomic\n"), 0o600)
	require.NoError(t, err)

	err = checkLeftoverCoverageInDir(logger, tmpDir)
	require.Error(t, err, "Should return error when coverage files found")
	require.Contains(t, err.Error(), "found and deleted")

	// Verify the file was deleted.
	_, statErr := os.Stat(covFile)
	require.True(t, os.IsNotExist(statErr), "Coverage file should have been deleted")
}

// TestCheckLeftoverCoverageInDir_Clean verifies no error when no coverage files exist.
func TestCheckLeftoverCoverageInDir_Clean(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Create a normal Go file (should not be detected as coverage file).
	goFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(goFile, []byte("package main\n"), 0o600)
	require.NoError(t, err)

	err = checkLeftoverCoverageInDir(logger, tmpDir)
	require.NoError(t, err, "Should pass when no coverage files exist")
}

// TestMatchesCoveragePattern_ComplexWildcard verifies that filenames like
// "MYCOVERAGE.HTML" (uppercase) are matched case-insensitively by the complex wildcard path.
// filepath.Match is case-sensitive on Linux, so uppercase names require the fallback path.
func TestMatchesCoveragePattern_ComplexWildcard(t *testing.T) {
	t.Parallel()

	// "mycoverage.html" matches *coverage*.html via filepath.Match (case-sensitive).
	require.True(t, matchesCoveragePattern("mycoverage.html"),
		"mycoverage.html should match *coverage*.html")

	// "MYCOVERAGE.HTML" (uppercase) bypasses filepath.Match on Linux (case-sensitive)
	// and is matched by the case-insensitive complex wildcard fallback.
	require.True(t, matchesCoveragePattern("MYCOVERAGE.HTML"),
		"MYCOVERAGE.HTML should match *coverage*.html via case-insensitive fallback")

	// "notafile.go" should not match.
	require.False(t, matchesCoveragePattern("notafile.go"),
		"notafile.go should not match any coverage pattern")
}

// TestMatchesCoveragePattern_CoverageOut verifies "coverage.out" is detected.
func TestMatchesCoveragePattern_CoverageOut(t *testing.T) {
	t.Parallel()

	require.True(t, matchesCoveragePattern("coverage.out"),
		"coverage.out should match *.out pattern")
	require.True(t, matchesCoveragePattern("profile.prof"),
		"profile.prof should match *.prof pattern")
	require.False(t, matchesCoveragePattern("main.go"),
		"main.go should not match any coverage pattern")
}

// TestCheckLeftoverCoverageInDir_UnreadableSubdir verifies that checkLeftoverCoverageInDir
// returns error when a subdirectory cannot be accessed during the walk.
func TestCheckLeftoverCoverageInDir_UnreadableSubdir(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Create a subdirectory with no read permissions to trigger walkErr.
	unaccessibleDir := filepath.Join(tmpDir, "noperm")
	err := os.MkdirAll(unaccessibleDir, 0o000)
	require.NoError(t, err)

	// Restore directory permissions on cleanup so TempDir can clean up.
	t.Cleanup(func() {
		_ = os.Chmod(unaccessibleDir, 0o700)
	})

	err = checkLeftoverCoverageInDir(logger, tmpDir)
	require.Error(t, err, "Should return error when subdirectory cannot be read")
	require.Contains(t, err.Error(), "failed to walk")
}

// TestCheckLeftoverCoverageInDir_RemoveError verifies that when a coverage file
// cannot be deleted, the function logs a warning but does not return an error
// (since no files were successfully deleted, deletedFiles remains empty).
func TestCheckLeftoverCoverageInDir_RemoveError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Create a subdirectory containing a coverage file.
	subDir := filepath.Join(tmpDir, "reports")
	err := os.MkdirAll(subDir, 0o755)
	require.NoError(t, err)

	covFile := filepath.Join(subDir, "coverage.out")
	err = os.WriteFile(covFile, []byte("mode: atomic\n"), 0o600)
	require.NoError(t, err)

	// Make the subdirectory read-only so Remove will fail.
	err = os.Chmod(subDir, 0o555)
	require.NoError(t, err)

	// Restore directory permissions on cleanup.
	t.Cleanup(func() {
		_ = os.Chmod(subDir, 0o700)
	})

	// With no successful deletions, deletedFiles is empty and function returns nil.
	err = checkLeftoverCoverageInDir(logger, tmpDir)
	require.NoError(t, err, "Should not error when no file was successfully removed")
}

// TestCheckCmdMainPatternInDir_WalkError verifies that checkCmdMainPatternInDir
// returns error when a cmd/ subdirectory is not accessible during the walk.
func TestCheckCmdMainPatternInDir_WalkError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Create cmd/myapp but make myapp unreadable.
	cmdDir := filepath.Join(tmpDir, "cmd", "myapp")
	err := os.MkdirAll(cmdDir, 0o755)
	require.NoError(t, err)

	// Place a main.go inside the cmd dir.
	mainFile := filepath.Join(cmdDir, "main.go")
	err = os.WriteFile(mainFile, []byte("package main\nfunc main(){}"), 0o600)
	require.NoError(t, err)

	// Make myapp inaccessible.
	err = os.Chmod(cmdDir, 0o000)
	require.NoError(t, err)

	t.Cleanup(func() { _ = os.Chmod(cmdDir, 0o700) })

	err = checkCmdMainPatternInDir(logger, tmpDir)
	require.Error(t, err, "Should error when cmd subdirectory is not accessible")
	require.Contains(t, err.Error(), "failed to walk cmd directory")
}

// TestFindMathRandViolationsInDir_WalkDirError verifies that findMathRandViolationsInDir
// returns error when a subdirectory is inaccessible during the walk.
func TestFindMathRandViolationsInDir_WalkDirError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create an inaccessible subdirectory.
	badDir := filepath.Join(tmpDir, "baddir")
	err := os.MkdirAll(badDir, 0o000)
	require.NoError(t, err)

	t.Cleanup(func() { _ = os.Chmod(badDir, 0o700) })

	violations, err := findMathRandViolationsInDir(tmpDir)
	require.Error(t, err, "Should error when a subdirectory cannot be accessed")
	require.Nil(t, violations)
}

// TestFindInsecureSkipVerifyViolationsInDir_WalkDirError verifies error on inaccessible dir.
func TestFindInsecureSkipVerifyViolationsInDir_WalkDirError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	badDir := filepath.Join(tmpDir, "baddir")
	err := os.MkdirAll(badDir, 0o000)
	require.NoError(t, err)

	t.Cleanup(func() { _ = os.Chmod(badDir, 0o700) })

	violations, err := findInsecureSkipVerifyViolationsInDir(tmpDir)
	require.Error(t, err, "Should error when a subdirectory cannot be accessed")
	require.Nil(t, violations)
}

// Suppress unused import warning for fmt if no direct fmt call is made.
var _ = fmt.Sprintf
