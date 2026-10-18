// Copyright (c) 2025-2026 Justin Cranford.
package migration_comment_headers_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	lintFitnessMigrationCommentHeaders "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/migration_comment_headers"
	lintFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
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

// createAllPSDirStubs creates a minimal internal/apps/{InternalAppsDir} stub for every
// PS in the registry. This satisfies the hard-error-on-absent-dir requirement without
// pre-populating migrations content, so individual tests can control what they add.
func createAllPSDirStubs(t *testing.T, tmpDir string) {
	t.Helper()

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		dir := filepath.Join(tmpDir, "internal", "apps", filepath.FromSlash(ps.InternalAppsDir))
		require.NoError(t, os.MkdirAll(dir, cryptoutilSharedMagic.CICDTempDirPermissions))
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

const smKMSMigrationDisplayName = "Key Management"

func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)

	err := lintFitnessMigrationCommentHeaders.CheckInDir(newTestLogger(), root)
	require.NoError(t, err)
}

// Sequential: changes the process working directory to exercise Check.
func TestCheck_WorkspaceRoot(t *testing.T) {
	root := findProjectRoot(t)
	original, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(root))
	defer func() {
		require.NoError(t, os.Chdir(original))
	}()

	require.NoError(t, lintFitnessMigrationCommentHeaders.Check(newTestLogger()))
}

func TestCheckInDir_CorrectHeaders(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createAllPSDirStubs(t, tmpDir)

	// Write correct migrations for sm-kms (representative PS with migrations).
	const smIMAppsDir = "sm-kms/"

	writeMigrationFile(t, tmpDir, smIMAppsDir, "2001_init.up.sql", upHeader(smKMSMigrationDisplayName))
	writeMigrationFile(t, tmpDir, smIMAppsDir, "2001_init.down.sql", downHeader(smKMSMigrationDisplayName))

	err := lintFitnessMigrationCommentHeaders.CheckInDir(newTestLogger(), tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_WrongUpHeader(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createAllPSDirStubs(t, tmpDir)

	const smIMAppsDir = "sm-kms/"

	writeMigrationFile(t, tmpDir, smIMAppsDir, "2001_init.up.sql",
		"--\n-- OLD NAME database schema\n-- Detail\n--\n\nCREATE TABLE foo (id INTEGER);\n",
	)
	writeMigrationFile(t, tmpDir, smIMAppsDir, "2001_init.down.sql", downHeader(smKMSMigrationDisplayName))

	err := lintFitnessMigrationCommentHeaders.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), smKMSMigrationDisplayName+" database schema")
	assert.Contains(t, err.Error(), "OLD NAME database schema")
}

func TestCheckInDir_WrongDownHeader(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createAllPSDirStubs(t, tmpDir)

	const smIMAppsDir = "sm-kms/"

	writeMigrationFile(t, tmpDir, smIMAppsDir, "2001_init.up.sql", upHeader(smKMSMigrationDisplayName))
	writeMigrationFile(t, tmpDir, smIMAppsDir, "2001_init.down.sql",
		"--\n-- OLD NAME database schema rollback\n-- Detail\n--\n\nDROP TABLE foo;\n",
	)

	err := lintFitnessMigrationCommentHeaders.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), smKMSMigrationDisplayName+" database schema rollback")
	assert.Contains(t, err.Error(), "OLD NAME database schema rollback")
}

func TestCheckInDir_NoCommentHeader(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createAllPSDirStubs(t, tmpDir)

	const smIMAppsDir = "sm-kms/"

	// File with no comment header at all.

	writeMigrationFile(t, tmpDir, smIMAppsDir, "2001_init.up.sql",
		"CREATE TABLE foo (id INTEGER);\n",
	)
	writeMigrationFile(t, tmpDir, smIMAppsDir, "2001_init.down.sql", downHeader(smKMSMigrationDisplayName))

	err := lintFitnessMigrationCommentHeaders.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no comment header")
}

func TestCheckInDir_FrameworkMigrationSkipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createAllPSDirStubs(t, tmpDir)

	const smIMAppsDir = "sm-kms/"

	// Framework migration (1001) - must be skipped even with wrong header.

	writeMigrationFile(t, tmpDir, smIMAppsDir, "1001_framework.up.sql",
		"--\n-- Some framework migration\n", // wrong header but should be skipped
	)
	// Domain migration (2001) with correct header.
	writeMigrationFile(t, tmpDir, smIMAppsDir, "2001_init.up.sql", upHeader(smKMSMigrationDisplayName))
	writeMigrationFile(t, tmpDir, smIMAppsDir, "2001_init.down.sql", downHeader(smKMSMigrationDisplayName))

	err := lintFitnessMigrationCommentHeaders.CheckInDir(newTestLogger(), tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_NoMigrationsDir_NoError(t *testing.T) {
	t.Parallel()

	// A PS with no migrations directory (but the apps dir exists) produces no violation.
	tmpDir := t.TempDir()
	createAllPSDirStubs(t, tmpDir)

	// sm-kms apps dir already created by createAllPSDirStubs; no migrations subdir added.

	err := lintFitnessMigrationCommentHeaders.CheckInDir(newTestLogger(), tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_ArchivedMigrationDirSkipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createAllPSDirStubs(t, tmpDir)
	writeMigrationFile(t, tmpDir, "sm-kms/_archive", "2001_invalid.up.sql", "not a canonical header\n")

	err := lintFitnessMigrationCommentHeaders.CheckInDir(newTestLogger(), tmpDir)
	require.NoError(t, err)
}

func TestCheckMigrationDir_IgnoresDirectoriesAndInvalidNames(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createAllPSDirStubs(t, tmpDir)

	const internalAppsDir = "sm-kms/"

	migrationsDir := filepath.Join(tmpDir, "internal", "apps", filepath.FromSlash(internalAppsDir), "repository", "migrations")
	require.NoError(t, os.MkdirAll(filepath.Join(migrationsDir, "nested"), cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "README.sql"), []byte("not a migration"), cryptoutilSharedMagic.CICDOutputFilePermissions))
	writeMigrationFile(t, tmpDir, internalAppsDir, "2001_init.up.sql", upHeader(smKMSMigrationDisplayName))

	err := lintFitnessMigrationCommentHeaders.CheckInDir(newTestLogger(), tmpDir)
	require.NoError(t, err)
}

func TestCheckPS_WalkErrors(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	createAllPSDirStubs(t, rootDir)

	ps := lintFitnessRegistry.AllProductServices()[0]

	tests := []struct {
		name   string
		walkFn func(string, filepath.WalkFunc) error
	}{
		{
			name: "walk function error",
			walkFn: func(string, filepath.WalkFunc) error {
				return os.ErrPermission
			},
		},
		{
			name: "callback error",
			walkFn: func(path string, callback filepath.WalkFunc) error {
				return callback(path, nil, os.ErrPermission)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			violations := lintFitnessMigrationCommentHeaders.ExportedCheckPSWithWalk(rootDir, ps, tc.walkFn)
			require.Len(t, violations, 1)
			require.Contains(t, violations[0], "error walking directory")
		})
	}
}

func TestCheckMigrationDir_ReadError(t *testing.T) {
	t.Parallel()

	ps := lintFitnessRegistry.AllProductServices()[0]
	violations := lintFitnessMigrationCommentHeaders.ExportedCheckMigrationDirWithReadDir(
		t.TempDir(),
		ps,
		"migrations",
		func(string) ([]os.DirEntry, error) {
			return nil, os.ErrPermission
		},
	)
	require.Len(t, violations, 1)
	require.Contains(t, violations[0], "cannot read migrations dir")
}

func TestCheckMigrationFile_OpenAndScannerErrors(t *testing.T) {
	t.Parallel()

	ps := lintFitnessRegistry.AllProductServices()[0]
	rootDir := t.TempDir()
	missingPath := filepath.Join(rootDir, "missing.up.sql")
	openError := lintFitnessMigrationCommentHeaders.ExportedCheckMigrationFileWithOpen(
		rootDir,
		ps,
		missingPath,
		"up",
		func(string) (*os.File, error) {
			return nil, os.ErrPermission
		},
	)
	require.Len(t, openError, 1)
	require.Contains(t, openError[0], "cannot open")

	oversizedPath := filepath.Join(rootDir, "oversized.up.sql")
	require.NoError(t, os.WriteFile(oversizedPath, []byte(strings.Repeat("x", 100_000)), cryptoutilSharedMagic.CICDOutputFilePermissions))
	scannerError := lintFitnessMigrationCommentHeaders.ExportedCheckMigrationFileWithOpen(
		rootDir,
		ps,
		oversizedPath,
		"up",
		os.Open,
	)
	require.Len(t, scannerError, 1)
	require.Contains(t, scannerError[0], "read error")
}

func TestCheckPS_AbsentPSDir(t *testing.T) {
	t.Parallel()

	// A PS whose internal/apps/{InternalAppsDir} does not exist at all is a hard error.
	tmpDir := t.TempDir()
	// Do NOT call createAllPSDirStubs â€” leave all PS dirs absent.

	err := lintFitnessMigrationCommentHeaders.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not exist")
}
