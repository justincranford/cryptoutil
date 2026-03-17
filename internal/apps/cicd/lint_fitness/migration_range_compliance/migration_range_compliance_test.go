// Copyright (c) 2025 Justin Cranford

package migration_range_compliance

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

// makeMigrationDir creates a migrations directory with SQL files.
func makeMigrationDir(t *testing.T, root, relDir string, fileNumbers []int) {
	t.Helper()

	dir := filepath.Join(root, filepath.FromSlash(relDir))
	require.NoError(t, os.MkdirAll(dir, cryptoutilSharedMagic.DirPermissions))

	for _, n := range fileNumbers {
		name := fmt.Sprintf("%04d_init.up.sql", n)
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("-- migration"), cryptoutilSharedMagic.CacheFilePermissions))
	}
}

const (
	templateMigRelDir = "internal/apps/template/service/server/repository/migrations"
	joseMigRelDir     = "internal/apps/jose/ja/repository/migrations"
	identityMigRelDir = "internal/apps/identity/idp/repository/migrations"
)

// ---- CheckInDir: template range ----

func TestCheckInDir_TemplateMigrations_ValidRange_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	makeMigrationDir(t, tmp, templateMigRelDir, []int{1001, 1002, 1003})
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_TemplateMigrations_BelowMin_Fails(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	makeMigrationDir(t, tmp, templateMigRelDir, []int{1001, 0})
	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "migration range compliance")
}

func TestCheckInDir_TemplateMigrations_AboveMax_Fails(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	makeMigrationDir(t, tmp, templateMigRelDir, []int{1001, 2000})
	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "migration range compliance")
}

// ---- CheckInDir: domain range ----

func TestCheckInDir_DomainMigrations_ValidRange_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	makeMigrationDir(t, tmp, joseMigRelDir, []int{2001, 2002})
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_DomainMigrations_BelowMin_Fails(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	makeMigrationDir(t, tmp, joseMigRelDir, []int{2001, 1})
	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "migration range compliance")
}

// ---- CheckInDir: identity skipped ----

func TestCheckInDir_IdentityMigrations_Skipped(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	// Identity uses 0001-0011 legacy numbering — excluded from range compliance.
	makeMigrationDir(t, tmp, identityMigRelDir, []int{1, 2, 3, 11})
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_EmptyDir_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

// ---- findDomainMigrationDirs ----

func TestFindDomainMigrationDirs_NoAppsDir_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	appsDir := filepath.Join(tmp, "internal", "apps")
	templateDir := filepath.Join(tmp, "internal", "apps", cryptoutilSharedMagic.SkeletonTemplateServiceName, "service", "server", "repository", "migrations")
	dirs, err := findDomainMigrationDirs(appsDir, templateDir)
	require.NoError(t, err)
	require.Empty(t, dirs)
}

func TestFindDomainMigrationDirs_WithIdentity_ExcludesIt(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	makeMigrationDir(t, tmp, identityMigRelDir, []int{1, 2})
	appsDir := filepath.Join(tmp, "internal", "apps")
	templateDir := filepath.Join(tmp, "internal", "apps", cryptoutilSharedMagic.SkeletonTemplateServiceName, "service", "server", "repository", "migrations")
	dirs, err := findDomainMigrationDirs(appsDir, templateDir)
	require.NoError(t, err)

	for _, d := range dirs {
		require.NotContains(t, d, cryptoutilSharedMagic.IdentityProductName)
	}
}

func TestFindDomainMigrationDirs_WithJose_IncludesIt(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	makeMigrationDir(t, tmp, joseMigRelDir, []int{2001})
	appsDir := filepath.Join(tmp, "internal", "apps")
	templateDir := filepath.Join(tmp, "internal", "apps", cryptoutilSharedMagic.SkeletonTemplateServiceName, "service", "server", "repository", "migrations")
	dirs, err := findDomainMigrationDirs(appsDir, templateDir)
	require.NoError(t, err)
	require.NotEmpty(t, dirs)
}

// ---- checkDir directly ----

func TestCheckDir_TemplateDirWithBadFile_ReturnsViolations(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "migrations")
	require.NoError(t, os.MkdirAll(dir, cryptoutilSharedMagic.DirPermissions))
	// File below template minimum range.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "0001_init.up.sql"), []byte("-- bad"), cryptoutilSharedMagic.CacheFilePermissions))
	violations, err := checkDir(dir, templateMigrationMin, templateMigrationMax, true)
	require.NoError(t, err)
	require.NotEmpty(t, violations)
}

func TestCheckDir_ValidTemplateFile_ReturnsNoViolations(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "migrations")
	require.NoError(t, os.MkdirAll(dir, cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "1001_init.up.sql"), []byte("-- ok"), cryptoutilSharedMagic.CacheFilePermissions))
	violations, err := checkDir(dir, templateMigrationMin, templateMigrationMax, true)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestCheckDir_NonSQLFile_Ignored(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "migrations")
	require.NoError(t, os.MkdirAll(dir, cryptoutilSharedMagic.DirPermissions))
	// README.md must not trigger a range violation.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("docs"), cryptoutilSharedMagic.CacheFilePermissions))
	violations, err := checkDir(dir, templateMigrationMin, templateMigrationMax, true)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestCheckDir_MissingDir_ReturnsNilNoError(t *testing.T) {
	t.Parallel()

	violations, err := checkDir("/nonexistent/migrations", templateMigrationMin, templateMigrationMax, true)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestCheckDir_WithSubdirectory_IsSkipped(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "migrations")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "subdir"), cryptoutilSharedMagic.DirPermissions))
	// Only the subdir is in migrations dir; subdir entries should be skipped (entry.IsDir() continue).
	violations, err := checkDir(dir, templateMigrationMin, templateMigrationMax, true)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestCheckDir_WithNonMatchingSqlFile_IsSkipped(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "migrations")
	require.NoError(t, os.MkdirAll(dir, cryptoutilSharedMagic.DirPermissions))
	// File like "init.sql" has no numeric prefix - matches == nil, continue.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "init.sql"), []byte("-- migration"), cryptoutilSharedMagic.CacheFilePermissions))
	violations, err := checkDir(dir, templateMigrationMin, templateMigrationMax, true)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestFindDomainMigrationDirs_WithArchivedSubdir_Skipped(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	// Create a _archived subdirectory under appsDir - it should be skipped via strings.HasPrefix check.
	archived := filepath.Join(tmp, "_archived", "migrations")
	require.NoError(t, os.MkdirAll(archived, cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(archived, "0001_invalid.up.sql"), []byte("-- bad"), cryptoutilSharedMagic.CacheFilePermissions))

	appsDir := tmp
	templateDir := filepath.Join(tmp, cryptoutilSharedMagic.SkeletonTemplateServiceName, "migrations")
	dirs, err := findDomainMigrationDirs(appsDir, templateDir)
	require.NoError(t, err)
	require.Empty(t, dirs)
}

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

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_Integration(t *testing.T) {
	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping - cannot find project root")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(root))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-migration-range-compliance")

	err = Check(logger)
	require.NoError(t, err)
}

func TestCheckInDir_NoInternalAppsDir_Succeeds(t *testing.T) {
	t.Parallel()

	// When appsDir doesn't exist, findDomainMigrationDirs returns nil/nil.
	// When the template dir doesn't exist, checkDir returns nil/nil.
	// CheckInDir should succeed with no violations.
	tmp := t.TempDir()
	err := CheckInDir(cryptoutilCmdCicdCommon.NewLogger("test"), tmp)
	require.NoError(t, err)
}
