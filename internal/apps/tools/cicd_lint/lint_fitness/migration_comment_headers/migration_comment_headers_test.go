// Copyright (c) 2025 Justin Cranford

package migration_comment_headers_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessMigrationCommentHeaders "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/migration_comment_headers"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

func findProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err, "failed to get working directory")

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("skipping integration test: cannot find project root (no go.mod)")
		}

		dir = parent
	}
}

// writeMigrationFile creates a SQL migration file under a migrations directory.
func writeMigrationFile(t *testing.T, tmpDir, psAppsDir, filename, content string) {
	t.Helper()

	migDir := filepath.Join(tmpDir, "internal", "apps", filepath.FromSlash(psAppsDir), "repository", "migrations")
	require.NoError(t, os.MkdirAll(migDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(migDir, filename), []byte(content), cryptoutilSharedMagic.CICDOutputFilePermissions))
}

// upHeader builds a correct up migration header for a display name.
func upHeader(displayName string) string {
	return "--\n-- " + displayName + " database schema\n-- Detail line\n--\n\nCREATE TABLE foo (id INTEGER);\n"
}

// downHeader builds a correct down migration header for a display name.
func downHeader(displayName string) string {
	return "--\n-- " + displayName + " database schema rollback\n-- Detail line\n--\n\nDROP TABLE foo;\n"
}

func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)

	err := lintFitnessMigrationCommentHeaders.CheckInDir(newTestLogger(), root)
	require.NoError(t, err)
}

func TestCheckInDir_CorrectHeaders(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Write correct migrations for sm-im (representative PS with migrations).
	const (
		smIMAppsDir     = "sm-im/"
		smIMDisplayName = "Instant Messenger"
	)

	writeMigrationFile(t, tmpDir, smIMAppsDir, "2001_init.up.sql", upHeader(smIMDisplayName))
	writeMigrationFile(t, tmpDir, smIMAppsDir, "2001_init.down.sql", downHeader(smIMDisplayName))

	err := lintFitnessMigrationCommentHeaders.CheckInDir(newTestLogger(), tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_WrongUpHeader(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	const (
		smIMAppsDir     = "sm-im/"
		smIMDisplayName = "Instant Messenger"
	)

	writeMigrationFile(t, tmpDir, smIMAppsDir, "2001_init.up.sql",
		"--\n-- OLD NAME database schema\n-- Detail\n--\n\nCREATE TABLE foo (id INTEGER);\n",
	)
	writeMigrationFile(t, tmpDir, smIMAppsDir, "2001_init.down.sql", downHeader(smIMDisplayName))

	err := lintFitnessMigrationCommentHeaders.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), smIMDisplayName+" database schema")
	assert.Contains(t, err.Error(), "OLD NAME database schema")
}

func TestCheckInDir_WrongDownHeader(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	const (
		smIMAppsDir     = "sm-im/"
		smIMDisplayName = "Instant Messenger"
	)

	writeMigrationFile(t, tmpDir, smIMAppsDir, "2001_init.up.sql", upHeader(smIMDisplayName))
	writeMigrationFile(t, tmpDir, smIMAppsDir, "2001_init.down.sql",
		"--\n-- OLD NAME database schema rollback\n-- Detail\n--\n\nDROP TABLE foo;\n",
	)

	err := lintFitnessMigrationCommentHeaders.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), smIMDisplayName+" database schema rollback")
	assert.Contains(t, err.Error(), "OLD NAME database schema rollback")
}

func TestCheckInDir_NoCommentHeader(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	const (
		smIMAppsDir     = "sm-im/"
		smIMDisplayName = "Instant Messenger"
	)

	// File with no comment header at all.

	writeMigrationFile(t, tmpDir, smIMAppsDir, "2001_init.up.sql",
		"CREATE TABLE foo (id INTEGER);\n",
	)
	writeMigrationFile(t, tmpDir, smIMAppsDir, "2001_init.down.sql", downHeader(smIMDisplayName))

	err := lintFitnessMigrationCommentHeaders.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no comment header")
}

func TestCheckInDir_FrameworkMigrationSkipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	const (
		smIMAppsDir     = "sm-im/"
		smIMDisplayName = "Instant Messenger"
	)

	// Framework migration (1001) - must be skipped even with wrong header.

	writeMigrationFile(t, tmpDir, smIMAppsDir, "1001_framework.up.sql",
		"--\n-- Some framework migration\n", // wrong header but should be skipped
	)
	// Domain migration (2001) with correct header.
	writeMigrationFile(t, tmpDir, smIMAppsDir, "2001_init.up.sql", upHeader(smIMDisplayName))
	writeMigrationFile(t, tmpDir, smIMAppsDir, "2001_init.down.sql", downHeader(smIMDisplayName))

	err := lintFitnessMigrationCommentHeaders.CheckInDir(newTestLogger(), tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_NoMigrationsDir_Skipped(t *testing.T) {
	t.Parallel()

	// PS with no migrations directory should be silently skipped.
	tmpDir := t.TempDir()

	// Create sm/im/ apps dir but no migrations subdir.
	appsDir := filepath.Join(tmpDir, "internal", "apps", "sm", "im")
	require.NoError(t, os.MkdirAll(appsDir, cryptoutilSharedMagic.CICDTempDirPermissions))

	err := lintFitnessMigrationCommentHeaders.CheckInDir(newTestLogger(), tmpDir)
	require.NoError(t, err)
}
