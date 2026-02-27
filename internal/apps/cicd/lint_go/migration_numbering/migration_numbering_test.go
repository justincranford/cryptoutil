// Copyright (c) 2025 Justin Cranford

package migration_numbering

import (
"fmt"
"os"
"path/filepath"
"testing"

cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

"github.com/stretchr/testify/require"
)

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}

		dir = parent
	}
}

func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping integration test - cannot find project root (no go.mod)")
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, root)
	require.NoError(t, err)
}

func TestCheckInDir_ValidDomainMigrations(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
migrationsDir := filepath.Join(tmpDir, "internal", "apps", "myproduct", "myservice", "repository", "migrations")
require.NoError(t, os.MkdirAll(migrationsDir, 0o755))
require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "2001_init.up.sql"), []byte("CREATE TABLE t;"), cryptoutilSharedMagic.CacheFilePermissions))
require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "2001_init.down.sql"), []byte("DROP TABLE t;"), cryptoutilSharedMagic.CacheFilePermissions))

logger := cryptoutilCmdCicdCommon.NewLogger("test")
err := CheckInDir(logger, tmpDir)
require.NoError(t, err)
}

func TestCheckInDir_ValidTemplateMigrations(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
templateDir := filepath.Join(tmpDir, "internal", "apps", "template", "service", "server", "repository", "migrations")
require.NoError(t, os.MkdirAll(templateDir, 0o755))
require.NoError(t, os.WriteFile(filepath.Join(templateDir, "1001_sessions.up.sql"), []byte("CREATE TABLE s;"), cryptoutilSharedMagic.CacheFilePermissions))
require.NoError(t, os.WriteFile(filepath.Join(templateDir, "1001_sessions.down.sql"), []byte("DROP TABLE s;"), cryptoutilSharedMagic.CacheFilePermissions))

logger := cryptoutilCmdCicdCommon.NewLogger("test")
err := CheckInDir(logger, tmpDir)
require.NoError(t, err)
}

func TestCheckInDir_DomainVersionBelowMinimum(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
migrationsDir := filepath.Join(tmpDir, "internal", "apps", "myproduct", "myservice", "repository", "migrations")
require.NoError(t, os.MkdirAll(migrationsDir, 0o755))
require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "0001_init.up.sql"), []byte("CREATE TABLE t;"), cryptoutilSharedMagic.CacheFilePermissions))
require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "0001_init.down.sql"), []byte("DROP TABLE t;"), cryptoutilSharedMagic.CacheFilePermissions))

logger := cryptoutilCmdCicdCommon.NewLogger("test")
err := CheckInDir(logger, tmpDir)
require.Error(t, err)
require.Contains(t, err.Error(), "below minimum 2001")
}

func TestCheckInDir_TemplateVersionAboveMaximum(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
templateDir := filepath.Join(tmpDir, "internal", "apps", "template", "service", "server", "repository", "migrations")
require.NoError(t, os.MkdirAll(templateDir, 0o755))
require.NoError(t, os.WriteFile(filepath.Join(templateDir, "2001_too_high.up.sql"), []byte("CREATE TABLE t;"), cryptoutilSharedMagic.CacheFilePermissions))
require.NoError(t, os.WriteFile(filepath.Join(templateDir, "2001_too_high.down.sql"), []byte("DROP TABLE t;"), cryptoutilSharedMagic.CacheFilePermissions))

logger := cryptoutilCmdCicdCommon.NewLogger("test")
err := CheckInDir(logger, tmpDir)
require.Error(t, err)
require.Contains(t, err.Error(), "exceeds maximum 1999")
}

func TestCheckInDir_MissingDownFile(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
migrationsDir := filepath.Join(tmpDir, "internal", "apps", "myproduct", "myservice", "repository", "migrations")
require.NoError(t, os.MkdirAll(migrationsDir, 0o755))
require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "2001_init.up.sql"), []byte("CREATE TABLE t;"), cryptoutilSharedMagic.CacheFilePermissions))

logger := cryptoutilCmdCicdCommon.NewLogger("test")
err := CheckInDir(logger, tmpDir)
require.Error(t, err)
require.Contains(t, err.Error(), "missing .down.sql")
}

func TestCheckInDir_MissingUpFile(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
migrationsDir := filepath.Join(tmpDir, "internal", "apps", "myproduct", "myservice", "repository", "migrations")
require.NoError(t, os.MkdirAll(migrationsDir, 0o755))
require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "2001_init.down.sql"), []byte("DROP TABLE t;"), cryptoutilSharedMagic.CacheFilePermissions))

logger := cryptoutilCmdCicdCommon.NewLogger("test")
err := CheckInDir(logger, tmpDir)
require.Error(t, err)
require.Contains(t, err.Error(), "missing .up.sql")
}

func TestCheckInDir_InvalidFilename(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
migrationsDir := filepath.Join(tmpDir, "internal", "apps", "myproduct", "myservice", "repository", "migrations")
require.NoError(t, os.MkdirAll(migrationsDir, 0o755))
require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "bad_file.sql"), []byte("INVALID"), cryptoutilSharedMagic.CacheFilePermissions))

logger := cryptoutilCmdCicdCommon.NewLogger("test")
err := CheckInDir(logger, tmpDir)
require.Error(t, err)
require.Contains(t, err.Error(), "does not match migration naming pattern")
}

func TestCheckInDir_NoAppsDir(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()

logger := cryptoutilCmdCicdCommon.NewLogger("test")
err := CheckInDir(logger, tmpDir)
require.NoError(t, err)
}

func TestCheckInDir_EmptyMigrationsDir(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
migrationsDir := filepath.Join(tmpDir, "internal", "apps", "myproduct", "myservice", "repository", "migrations")
require.NoError(t, os.MkdirAll(migrationsDir, 0o755))

logger := cryptoutilCmdCicdCommon.NewLogger("test")
err := CheckInDir(logger, tmpDir)
require.NoError(t, err)
}

func TestCheckInDir_ArchivedDirsSkipped(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
archivedDir := filepath.Join(tmpDir, "internal", "apps", "pki", "_ca-archived", "repository", "migrations")
require.NoError(t, os.MkdirAll(archivedDir, 0o755))
require.NoError(t, os.WriteFile(filepath.Join(archivedDir, "0001_bad.up.sql"), []byte("BAD"), cryptoutilSharedMagic.CacheFilePermissions))

logger := cryptoutilCmdCicdCommon.NewLogger("test")
err := CheckInDir(logger, tmpDir)
require.NoError(t, err)
}

func TestCheckInDir_MultipleDomainDirs(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()

dir1 := filepath.Join(tmpDir, "internal", "apps", "product1", "svc1", "repository", "migrations")
require.NoError(t, os.MkdirAll(dir1, 0o755))
require.NoError(t, os.WriteFile(filepath.Join(dir1, "2001_init.up.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))
require.NoError(t, os.WriteFile(filepath.Join(dir1, "2001_init.down.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))

dir2 := filepath.Join(tmpDir, "internal", "apps", "product2", "svc2", "repository", "migrations")
require.NoError(t, os.MkdirAll(dir2, 0o755))
require.NoError(t, os.WriteFile(filepath.Join(dir2, "2002_tables.up.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))
require.NoError(t, os.WriteFile(filepath.Join(dir2, "2002_tables.down.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))

logger := cryptoutilCmdCicdCommon.NewLogger("test")
err := CheckInDir(logger, tmpDir)
require.NoError(t, err)
}

func TestCheckInDir_TemplateBelowMinimum(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
templateDir := filepath.Join(tmpDir, "internal", "apps", "template", "service", "server", "repository", "migrations")
require.NoError(t, os.MkdirAll(templateDir, 0o755))
require.NoError(t, os.WriteFile(filepath.Join(templateDir, "0500_too_low.up.sql"), []byte("BAD"), cryptoutilSharedMagic.CacheFilePermissions))
require.NoError(t, os.WriteFile(filepath.Join(templateDir, "0500_too_low.down.sql"), []byte("BAD"), cryptoutilSharedMagic.CacheFilePermissions))

logger := cryptoutilCmdCicdCommon.NewLogger("test")
err := CheckInDir(logger, tmpDir)
require.Error(t, err)
require.Contains(t, err.Error(), "below minimum 1001")
}

func TestCheckInDir_MultipleVersions(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
migrationsDir := filepath.Join(tmpDir, "internal", "apps", "myproduct", "myservice", "repository", "migrations")
require.NoError(t, os.MkdirAll(migrationsDir, 0o755))
require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "2001_init.up.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))
require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "2001_init.down.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))
require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "2002_add_column.up.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))
require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "2002_add_column.down.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))

logger := cryptoutilCmdCicdCommon.NewLogger("test")
err := CheckInDir(logger, tmpDir)
require.NoError(t, err)
}

func TestCheckMigrationDir_SubdirectoriesIgnored(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
migrationsDir := filepath.Join(tmpDir, "migrations")
require.NoError(t, os.MkdirAll(migrationsDir, 0o755))
require.NoError(t, os.MkdirAll(filepath.Join(migrationsDir, "subdir"), 0o755))
require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "2001_init.up.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))
require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "2001_init.down.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))

errs := checkMigrationDir(migrationsDir, domainMigrationMin, 0, false)
require.Empty(t, errs)
}

func TestCheckMigrationDir_NonexistentDir(t *testing.T) {
t.Parallel()

errs := checkMigrationDir("/nonexistent/migrations", domainMigrationMin, 0, false)
require.Empty(t, errs)
}

func TestFindDomainMigrationDirs_NoAppsDir(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()

dirs, err := findDomainMigrationDirs(tmpDir, filepath.Join(tmpDir, "internal", "apps", "template", "service", "server", "repository", "migrations"))
require.NoError(t, err)
require.Empty(t, dirs)
}
