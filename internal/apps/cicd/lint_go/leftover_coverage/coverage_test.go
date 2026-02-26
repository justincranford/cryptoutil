// Copyright (c) 2025 Justin Cranford

package leftover_coverage

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

// TestCheckMainGoFile_ReadError verifies that lintGoCmdMainPattern.CheckMainGoFile returns error when file cannot be read.

// TestCheckLeftoverCoverageInDir_WithCoverageFiles verifies that coverage files are detected
// and deleted, returning an error to trigger CI awareness.
func TestCheckLeftoverCoverageInDir_WithCoverageFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Create a leftover coverage file.
	covFile := filepath.Join(tmpDir, "coverage.out")
	err := os.WriteFile(covFile, []byte("mode: atomic\n"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	err = CheckInDir(logger, tmpDir)
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
	err := os.WriteFile(goFile, []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	err = CheckInDir(logger, tmpDir)
	require.NoError(t, err, "Should pass when no coverage files exist")
}

// TestMatchesCoveragePattern_ComplexWildcard verifies that filenames like
// "MYCOVERAGE.HTML" (uppercase) are matched case-insensitively by the complex wildcard path.
// filepath.Match is case-sensitive on Linux, so uppercase names require the fallback path.
func TestMatchesCoveragePattern_ComplexWildcard(t *testing.T) {
	t.Parallel()

	// "mycoverage.html" matches *coverage*.html via filepath.Match (case-sensitive).
	require.True(t, MatchesCoveragePattern("mycoverage.html"),
		"mycoverage.html should match *coverage*.html")

	// "MYCOVERAGE.HTML" (uppercase) bypasses filepath.Match on Linux (case-sensitive)
	// and is matched by the case-insensitive complex wildcard fallback.
	require.True(t, MatchesCoveragePattern("MYCOVERAGE.HTML"),
		"MYCOVERAGE.HTML should match *coverage*.html via case-insensitive fallback")

	// "notafile.go" should not match.
	require.False(t, MatchesCoveragePattern("notafile.go"),
		"notafile.go should not match any coverage pattern")
}

// TestMatchesCoveragePattern_CoverageOut verifies "coverage.out" is detected.
func TestMatchesCoveragePattern_CoverageOut(t *testing.T) {
	t.Parallel()

	require.True(t, MatchesCoveragePattern("coverage.out"),
		"coverage.out should match *.out pattern")
	require.True(t, MatchesCoveragePattern("profile.prof"),
		"profile.prof should match *.prof pattern")
	require.False(t, MatchesCoveragePattern("main.go"),
		"main.go should not match any coverage pattern")
}

// TestCheckLeftoverCoverageInDir_UnreadableSubdir verifies that CheckInDir
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

	err = CheckInDir(logger, tmpDir)
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
	err := os.MkdirAll(subDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
	require.NoError(t, err)

	covFile := filepath.Join(subDir, "coverage.out")
	err = os.WriteFile(covFile, []byte("mode: atomic\n"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Make the subdirectory read-only so Remove will fail.
	err = os.Chmod(subDir, 0o555)
	require.NoError(t, err)

	// Restore directory permissions on cleanup.
	t.Cleanup(func() {
		_ = os.Chmod(subDir, 0o700)
	})

	// With no successful deletions, deletedFiles is empty and function returns nil.
	err = CheckInDir(logger, tmpDir)
	require.NoError(t, err, "Should not error when no file was successfully removed")
}
