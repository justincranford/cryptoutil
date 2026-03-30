// Copyright (c) 2025 Justin Cranford

package migration_range_compliance

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

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
	templateMigRelDir = "internal/apps/framework/service/server/repository/migrations"
	joseMigRelDir     = "internal/apps/jose-ja/repository/migrations"
	smImMigRelDir     = "internal/apps/sm-im/repository/migrations"
	identityMigRelDir = "internal/apps/identity-idp/repository/migrations"
	unknownSvcMigDir  = "internal/apps/unknown-service/repository/migrations"
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

// ---- CheckInDir: domain range per-PS-ID ----

func TestCheckInDir_DomainMigrations_ValidRange_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	// jose-ja registry range is 4001-4999.
	makeMigrationDir(t, tmp, joseMigRelDir, []int{4001, 4002})
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_DomainMigrations_BelowMin_Fails(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	// jose-ja min is 4001; version 1 is below range.
	makeMigrationDir(t, tmp, joseMigRelDir, []int{4001, 1})
	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "migration range compliance")
}

func TestCheckInDir_DomainMigrations_AboveMax_Fails(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	// jose-ja max is 4999; version 5001 is above range.
	makeMigrationDir(t, tmp, joseMigRelDir, []int{4001, 5001})
	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "migration range compliance")
}

func TestCheckInDir_DomainMigrations_SmIm_ValidRange_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	// sm-im registry range is 3001-3999.
	makeMigrationDir(t, tmp, smImMigRelDir, []int{3001, 3002})
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
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
	templateDir := filepath.Join(tmp, "internal", "apps", cryptoutilSharedMagic.FrameworkProductName, "service", "server", "repository", "migrations")
	dirs, err := findDomainMigrationDirs(appsDir, templateDir)
	require.NoError(t, err)
	require.Empty(t, dirs)
}

func TestFindDomainMigrationDirs_WithIdentity_ExcludesIt(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	makeMigrationDir(t, tmp, identityMigRelDir, []int{1, 2})
	appsDir := filepath.Join(tmp, "internal", "apps")
	templateDir := filepath.Join(tmp, "internal", "apps", cryptoutilSharedMagic.FrameworkProductName, "service", "server", "repository", "migrations")
	dirs, err := findDomainMigrationDirs(appsDir, templateDir)
	require.NoError(t, err)
	require.Empty(t, dirs, "identity migrations should be excluded from domain range compliance")
}

func TestFindDomainMigrationDirs_WithJose_IncludesIt(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	makeMigrationDir(t, tmp, joseMigRelDir, []int{4001})
	appsDir := filepath.Join(tmp, "internal", "apps")
	templateDir := filepath.Join(tmp, "internal", "apps", cryptoutilSharedMagic.FrameworkProductName, "service", "server", "repository", "migrations")
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
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte(cryptoutilSharedMagic.CICDExcludeDirDocs), cryptoutilSharedMagic.CacheFilePermissions))
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
	templateDir := filepath.Join(tmp, cryptoutilSharedMagic.FrameworkProductName, "migrations")
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

// ---- checkRegistryRangeCollisions ----

func TestCheckRegistryRangeCollisions_NoOverlaps(t *testing.T) {
	t.Parallel()

	violations := checkRegistryRangeCollisions()
	require.Empty(t, violations, "registry should have no overlapping migration ranges; got: %v", violations)
}

// ---- buildPSIDRangeMap ----

func TestBuildPSIDRangeMap_ContainsAllPSIDs(t *testing.T) {
	t.Parallel()

	rangeMap := buildPSIDRangeMap()
	allPS := lintFitnessRegistry.AllProductServices()

	for _, ps := range allPS {
		_, ok := rangeMap[ps.PSID]
		require.True(t, ok, "expected PS-ID %q in range map", ps.PSID)
	}
}

func TestBuildPSIDRangeMap_JoseJaHasCorrectRange(t *testing.T) {
	t.Parallel()

	rangeMap := buildPSIDRangeMap()
	joseRange, ok := rangeMap[cryptoutilSharedMagic.OTLPServiceJoseJA]
	require.True(t, ok)
	require.Equal(t, 4001, joseRange.Start)
	require.Equal(t, 4999, joseRange.End)
}

// ---- checkRangeCollisions ----

func TestCheckRangeCollisions_NoOverlap_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	ranges := []lintFitnessRegistry.MigrationRangeInfo{
		{PSID: "svc-a", Start: 1001, End: 1999},
		{PSID: "svc-b", Start: 2001, End: 2999},
		{PSID: "svc-c", Start: 3001, End: 3999},
	}
	violations := checkRangeCollisions(ranges)
	require.Empty(t, violations)
}

func TestCheckRangeCollisions_Overlap_ReturnsViolation(t *testing.T) {
	t.Parallel()

	ranges := []lintFitnessRegistry.MigrationRangeInfo{
		{PSID: "svc-a", Start: 1001, End: 2500},
		{PSID: "svc-b", Start: 2001, End: 2999},
	}
	violations := checkRangeCollisions(ranges)
	require.NotEmpty(t, violations)
	require.Contains(t, violations[0], "svc-a")
	require.Contains(t, violations[0], "svc-b")
}

func TestCheckRangeCollisions_EmptyRanges_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	violations := checkRangeCollisions(nil)
	require.Empty(t, violations)
}

// ---- CheckInDir: PS-ID not in registry falls back ----

func TestCheckInDir_UnknownPSID_UsesLooseLowerBound(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	// Use a directory name unknown to registry; files at 2001+ pass loose lower bound.
	makeMigrationDir(t, tmp, unknownSvcMigDir, []int{2001, 9999})
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_UnknownPSID_BelowLooseLower_Fails(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	// Use a directory name unknown to registry; file at 999 is below loose lower bound (1999+1=2000).
	makeMigrationDir(t, tmp, unknownSvcMigDir, []int{999})
	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "migration range compliance")
}

// Note: checkDir error path (non-IsNotExist ReadDir failure) and
// findDomainMigrationDirsWithPSID absErr path are structural ceilings:
// on Windows, os.ReadDir on a file returns os.IsNotExist (treated as "no dir"),
// and filepath.Abs never fails for valid paths. These OS-level failure paths
// cannot be triggered without OS-level intervention.
