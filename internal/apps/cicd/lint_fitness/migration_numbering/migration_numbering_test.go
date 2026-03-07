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
	require.NoError(t, os.MkdirAll(migrationsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "2001_init.up.sql"), []byte("CREATE TABLE t;"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "2001_init.down.sql"), []byte("DROP TABLE t;"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_ValidTemplateMigrations(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.SkeletonTemplateServiceName, "service", "server", "repository", "migrations")
	require.NoError(t, os.MkdirAll(templateDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "1001_sessions.up.sql"), []byte("CREATE TABLE s;"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "1001_sessions.down.sql"), []byte("DROP TABLE s;"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_TemplateAtMaxBoundary(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.SkeletonTemplateServiceName, "service", "server", "repository", "migrations")
	require.NoError(t, os.MkdirAll(templateDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	// Version 1999 is the maximum template version — must be accepted.
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "1999_last_template.up.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "1999_last_template.down.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_DomainVersionBelowMinimum(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "internal", "apps", "myproduct", "myservice", "repository", "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
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
	templateDir := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.SkeletonTemplateServiceName, "service", "server", "repository", "migrations")
	require.NoError(t, os.MkdirAll(templateDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
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
	require.NoError(t, os.MkdirAll(migrationsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
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
	require.NoError(t, os.MkdirAll(migrationsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
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
	require.NoError(t, os.MkdirAll(migrationsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
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
	require.NoError(t, os.MkdirAll(migrationsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_ArchivedDirsSkipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	archivedDir := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.PKIProductName, "_ca-archived", "repository", "migrations")
	require.NoError(t, os.MkdirAll(archivedDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(archivedDir, "0001_bad.up.sql"), []byte("BAD"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_MultipleDomainDirs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	dir1 := filepath.Join(tmpDir, "internal", "apps", "product1", "svc1", "repository", "migrations")
	require.NoError(t, os.MkdirAll(dir1, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "2001_init.up.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "2001_init.down.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))

	dir2 := filepath.Join(tmpDir, "internal", "apps", "product2", "svc2", "repository", "migrations")
	require.NoError(t, os.MkdirAll(dir2, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(dir2, "2002_tables.up.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(dir2, "2002_tables.down.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_TemplateBelowMinimum(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.SkeletonTemplateServiceName, "service", "server", "repository", "migrations")
	require.NoError(t, os.MkdirAll(templateDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
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
	require.NoError(t, os.MkdirAll(migrationsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
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
	require.NoError(t, os.MkdirAll(migrationsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(migrationsDir, "subdir"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
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

	dirs, err := findDomainMigrationDirs(tmpDir, filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.SkeletonTemplateServiceName, "service", "server", "repository", "migrations"))
	require.NoError(t, err)
	require.Empty(t, dirs)
}

// Sequential: uses os.Chdir (global process state).
func TestCheck_FromProjectRoot(t *testing.T) {
	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping - cannot find project root")
	}

	origDir, wdErr := os.Getwd()
	require.NoError(t, wdErr)

	require.NoError(t, os.Chdir(root))

	t.Cleanup(func() {
		require.NoError(t, os.Chdir(origDir))
	})

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	require.NoError(t, err)
}

func TestCheckMigrationDir_ReadDirError(t *testing.T) {
	t.Parallel()

	// Create a directory with a permissions issue.
	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	// Write a valid file, then make the dir unreadable.
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "2001_init.up.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.Chmod(migrationsDir, 0o000))

	t.Cleanup(func() {
		_ = os.Chmod(migrationsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
	})

	errs := checkMigrationDir(migrationsDir, domainMigrationMin, 0, false)
	require.NotEmpty(t, errs)
	require.Contains(t, errs[0], "failed to read directory")
}

func TestCheckMigrationDir_DomainAboveMax(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "9999_too_high.up.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "9999_too_high.down.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))

	// Template with maxVersion=1999 should catch >1999.
	errs := checkMigrationDir(migrationsDir, templateMigrationMin, templateMigrationMax, true)
	require.NotEmpty(t, errs)
	require.Contains(t, errs[0], "exceeds maximum")
}

func TestCheckInDir_TemplateExceedsMax(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.SkeletonTemplateServiceName, "service", "server", "repository", "migrations")
	require.NoError(t, os.MkdirAll(templateDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "2500_above_range.up.sql"), []byte("BAD"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "2500_above_range.down.sql"), []byte("BAD"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "exceeds maximum 1999")
}

func TestFindDomainMigrationDirs_LegacyExcluded(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a legacy dir that matches the exclusion list.
	legacyDir := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.IdentityProductName, "repository", "migrations")
	require.NoError(t, os.MkdirAll(legacyDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(legacyDir, "0001_legacy.up.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))

	// Create a non-legacy domain dir.
	domainDir := filepath.Join(tmpDir, "internal", "apps", "myservice", "repository", "migrations")
	require.NoError(t, os.MkdirAll(domainDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(domainDir, "2001_init.up.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(domainDir, "2001_init.down.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))

	dirs, err := findDomainMigrationDirs(tmpDir, filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.SkeletonTemplateServiceName, "service", "server", "repository", "migrations"))
	require.NoError(t, err)
	require.Len(t, dirs, 1)
	require.Contains(t, dirs[0], "myservice")
}

// saveRestoreSeams saves and restores test seams.
func saveRestoreSeams(t *testing.T) {
	t.Helper()

	origPathAbs := pathAbsFunc
	origAtoi := atoiFunc

	t.Cleanup(func() {
		pathAbsFunc = origPathAbs
		atoiFunc = origAtoi
	})
}

func TestFindDomainMigrationDirs_WalkPermissionError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	appsDir := filepath.Join(tmpDir, "internal", "apps")
	badDir := filepath.Join(appsDir, "badservice")
	require.NoError(t, os.MkdirAll(badDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.Chmod(badDir, 0o000))

	t.Cleanup(func() {
		_ = os.Chmod(badDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
	})

	templateDir := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.SkeletonTemplateServiceName, "service", "server", "repository", "migrations")

	_, err := findDomainMigrationDirs(tmpDir, templateDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to walk apps directory")
}

// Seam tests below: NOT parallel because they modify package-level seam variables.
// They must run after all parallel tests complete to avoid data races.

// Sequential: modifies package-level seam variables.
func TestSeam_FindDomainMigrationDirs_AbsTemplateError(t *testing.T) {
	saveRestoreSeams(t)

	pathAbsFunc = func(_ string) (string, error) {
		return "", fmt.Errorf("injected abs error")
	}

	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.SkeletonTemplateServiceName, "service", "server", "repository", "migrations")

	_, err := findDomainMigrationDirs(tmpDir, templateDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get absolute path for template dir")
}

// Sequential: modifies package-level seam variables.
func TestSeam_FindDomainMigrationDirs_WalkAbsError(t *testing.T) {
	saveRestoreSeams(t)

	callCount := 0

	pathAbsFunc = func(path string) (string, error) {
		callCount++
		// First call is for templateDir, second for legacy paths - let those succeed.
		// Fail on a later call during Walk.
		if callCount > 3 {
			return "", fmt.Errorf("injected walk abs error")
		}

		return filepath.Abs(path)
	}

	tmpDir := t.TempDir()
	appsDir := filepath.Join(tmpDir, "internal", "apps")
	migDir := filepath.Join(appsDir, "myservice", "repository", "migrations")
	require.NoError(t, os.MkdirAll(migDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(migDir, "2001_init.up.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))

	templateDir := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.SkeletonTemplateServiceName, "service", "server", "repository", "migrations")
	require.NoError(t, os.MkdirAll(templateDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	_, err := findDomainMigrationDirs(tmpDir, templateDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to walk apps directory")
}

// Sequential: modifies package-level seam variables.
func TestSeam_CheckInDir_FindDomainDirsError(t *testing.T) {
	saveRestoreSeams(t)

	pathAbsFunc = func(_ string) (string, error) {
		return "", fmt.Errorf("injected abs error for domain dirs")
	}

	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.SkeletonTemplateServiceName, "service", "server", "repository", "migrations")
	require.NoError(t, os.MkdirAll(templateDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to find domain migration directories")
}

// Sequential: modifies package-level seam variables.
func TestSeam_CheckMigrationDir_AtoiError(t *testing.T) {
	saveRestoreSeams(t)

	atoiFunc = func(_ string) (int, error) {
		return 0, fmt.Errorf("injected atoi error")
	}

	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "2001_init.up.sql"), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))

	errs := checkMigrationDir(migrationsDir, domainMigrationMin, 0, false)
	require.NotEmpty(t, errs)
	require.Contains(t, errs[0], "failed to parse version number")
}
