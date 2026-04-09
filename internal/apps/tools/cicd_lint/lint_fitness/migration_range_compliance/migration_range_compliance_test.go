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

// createTemplateMigrationsDirStub creates an empty template migrations directory
// so that tests focused on domain migration ranges do not fail on the absent template dir check.
func createTemplateMigrationsDirStub(t *testing.T, rootDir string) {
	t.Helper()

	dir := filepath.Join(rootDir, filepath.FromSlash(templateMigRelDir))
	require.NoError(t, os.MkdirAll(dir, cryptoutilSharedMagic.DirPermissions))
}

const (
	templateMigRelDir = "internal/apps/framework/service/server/repository/migrations"
	joseMigRelDir     = "internal/apps/jose-ja/repository/migrations"
	smImMigRelDir     = "internal/apps/sm-im/repository/migrations"
	identityMigRelDir = "internal/apps/identity-idp/repository/migrations"
	unknownSvcMigDir  = "internal/apps/unknown-service/repository/migrations"
)

func TestCheckInDir_TemplateMigrations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileNumbers []int
		wantErr     bool
		errContains string
	}{
		{name: "valid range passes", fileNumbers: []int{1001, 1002, 1003}},
		{name: "below min fails", fileNumbers: []int{1001, 0}, wantErr: true, errContains: "migration range compliance"},
		{name: "above max fails", fileNumbers: []int{1001, 2000}, wantErr: true, errContains: "migration range compliance"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tmp := t.TempDir()
			makeMigrationDir(t, tmp, templateMigRelDir, tc.fileNumbers)

			err := CheckInDir(newTestLogger(), tmp)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckInDir_DomainMigrations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		migRelDir   string
		fileNumbers []int
		wantErr     bool
		errContains string
	}{
		{name: "jose-ja valid range passes", migRelDir: joseMigRelDir, fileNumbers: []int{4001, 4002}},
		{name: "jose-ja below min fails", migRelDir: joseMigRelDir, fileNumbers: []int{4001, 1}, wantErr: true, errContains: "migration range compliance"},
		{name: "jose-ja above max fails", migRelDir: joseMigRelDir, fileNumbers: []int{4001, 5001}, wantErr: true, errContains: "migration range compliance"},
		{name: "sm-im valid range passes", migRelDir: smImMigRelDir, fileNumbers: []int{3001, 3002}},
		{name: "identity skipped", migRelDir: identityMigRelDir, fileNumbers: []int{1, 2, 3, 11}},
		{name: "empty dir passes", migRelDir: "", fileNumbers: nil},
		{name: "unknown psid valid loose bound", migRelDir: unknownSvcMigDir, fileNumbers: []int{2001, 9999}},
		{name: "unknown psid below loose lower fails", migRelDir: unknownSvcMigDir, fileNumbers: []int{999}, wantErr: true, errContains: "migration range compliance"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tmp := t.TempDir()
			createTemplateMigrationsDirStub(t, tmp)

			if tc.migRelDir != "" {
				makeMigrationDir(t, tmp, tc.migRelDir, tc.fileNumbers)
			}

			err := CheckInDir(newTestLogger(), tmp)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFindDomainMigrationDirs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(t *testing.T, tmp string)
		wantEmpty bool
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "no apps dir returns error",
			setup:     func(_ *testing.T, _ string) {},
			wantEmpty: true,
			wantErr:   true,
			errMsg:    "internal/apps directory not found",
		},
		{
			name: "identity excluded",
			setup: func(t *testing.T, tmp string) {
				t.Helper()
				makeMigrationDir(t, tmp, identityMigRelDir, []int{1, 2})
			},
			wantEmpty: true,
		},
		{
			name: "jose included",
			setup: func(t *testing.T, tmp string) {
				t.Helper()
				makeMigrationDir(t, tmp, joseMigRelDir, []int{4001})
			},
			wantEmpty: false,
		},
		{
			name: "archived subdir skipped",
			setup: func(t *testing.T, tmp string) {
				t.Helper()
				// Create appsDir so the existence check passes, then add _archived under it.
				appsDir := filepath.Join(tmp, "internal", "apps")
				require.NoError(t, os.MkdirAll(appsDir, cryptoutilSharedMagic.DirPermissions))
				archived := filepath.Join(appsDir, "_archived", "migrations")
				require.NoError(t, os.MkdirAll(archived, cryptoutilSharedMagic.DirPermissions))
				require.NoError(t, os.WriteFile(filepath.Join(archived, "0001_invalid.up.sql"), []byte("-- bad"), cryptoutilSharedMagic.CacheFilePermissions))
			},
			wantEmpty: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tmp := t.TempDir()
			tc.setup(t, tmp)

			appsDir := filepath.Join(tmp, "internal", "apps")
			templateDir := filepath.Join(tmp, "internal", "apps", cryptoutilSharedMagic.FrameworkProductName, "service", "server", "repository", "migrations")
			dirs, err := findDomainMigrationDirs(appsDir, templateDir)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
				require.Empty(t, dirs)
			} else {
				require.NoError(t, err)

				if tc.wantEmpty {
					require.Empty(t, dirs)
				} else {
					require.NotEmpty(t, dirs)
				}
			}
		})
	}
}

func TestCheckDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFiles     func(t *testing.T, dir string)
		min            int
		max            int
		isTemplate     bool
		wantViolations bool
	}{
		{
			name: "template bad file returns violations",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "0001_init.up.sql"), []byte("-- bad"), cryptoutilSharedMagic.CacheFilePermissions))
			},
			min: templateMigrationMin, max: templateMigrationMax, isTemplate: true,
			wantViolations: true,
		},
		{
			name: "valid template file no violations",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "1001_init.up.sql"), []byte("-- ok"), cryptoutilSharedMagic.CacheFilePermissions))
			},
			min: templateMigrationMin, max: templateMigrationMax, isTemplate: true,
		},
		{
			name: "non-SQL file ignored",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte(cryptoutilSharedMagic.CICDExcludeDirDocs), cryptoutilSharedMagic.CacheFilePermissions))
			},
			min: templateMigrationMin, max: templateMigrationMax, isTemplate: true,
		},
		{
			name: "subdirectory skipped",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "subdir"), cryptoutilSharedMagic.DirPermissions))
			},
			min: templateMigrationMin, max: templateMigrationMax, isTemplate: true,
		},
		{
			name: "non-matching SQL file skipped",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "init.sql"), []byte("-- migration"), cryptoutilSharedMagic.CacheFilePermissions))
			},
			min: templateMigrationMin, max: templateMigrationMax, isTemplate: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tmp := t.TempDir()
			dir := filepath.Join(tmp, "migrations")
			require.NoError(t, os.MkdirAll(dir, cryptoutilSharedMagic.DirPermissions))

			if tc.setupFiles != nil {
				tc.setupFiles(t, dir)
			}

			violations, err := checkDir(dir, tc.min, tc.max, tc.isTemplate)
			require.NoError(t, err)

			if tc.wantViolations {
				require.NotEmpty(t, violations)
			} else {
				require.Empty(t, violations)
			}
		})
	}
}

func TestCheckDir_MissingDir_ReturnsError(t *testing.T) {
	t.Parallel()

	violations, err := checkDir("/nonexistent/migrations", templateMigrationMin, templateMigrationMax, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "directory not found")
	require.Empty(t, violations)
}

func TestCheckRangeCollisions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		ranges         []lintFitnessRegistry.MigrationRangeInfo
		wantViolations bool
		checkContains  []string
	}{
		{
			name: "no overlap returns empty",
			ranges: []lintFitnessRegistry.MigrationRangeInfo{
				{PSID: "svc-a", Start: 1001, End: 1999},
				{PSID: "svc-b", Start: 2001, End: 2999},
				{PSID: "svc-c", Start: 3001, End: 3999},
			},
		},
		{
			name: "overlap returns violation",
			ranges: []lintFitnessRegistry.MigrationRangeInfo{
				{PSID: "svc-a", Start: 1001, End: 2500},
				{PSID: "svc-b", Start: 2001, End: 2999},
			},
			wantViolations: true,
			checkContains:  []string{"svc-a", "svc-b"},
		},
		{name: "empty ranges returns empty", ranges: nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			violations := checkRangeCollisions(tc.ranges)

			if tc.wantViolations {
				require.NotEmpty(t, violations)

				for _, s := range tc.checkContains {
					require.Contains(t, violations[0], s)
				}
			} else {
				require.Empty(t, violations)
			}
		})
	}
}

func TestCheckRegistryRangeCollisions_NoOverlaps(t *testing.T) {
	t.Parallel()

	violations := checkRegistryRangeCollisions()
	require.Empty(t, violations, "registry should have no overlapping migration ranges; got: %v", violations)
}

func TestBuildPSIDRangeMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		check func(t *testing.T, rangeMap map[string]lintFitnessRegistry.MigrationRangeInfo)
	}{
		{
			name: "contains all PSIDs",
			check: func(t *testing.T, rangeMap map[string]lintFitnessRegistry.MigrationRangeInfo) {
				t.Helper()

				for _, ps := range lintFitnessRegistry.AllProductServices() {
					_, ok := rangeMap[ps.PSID]
					require.True(t, ok, "expected PS-ID %q in range map", ps.PSID)
				}
			},
		},
		{
			name: "jose-ja has correct range",
			check: func(t *testing.T, rangeMap map[string]lintFitnessRegistry.MigrationRangeInfo) {
				t.Helper()

				joseRange, ok := rangeMap[cryptoutilSharedMagic.OTLPServiceJoseJA]
				require.True(t, ok)
				require.Equal(t, 4001, joseRange.Start)
				require.Equal(t, 4999, joseRange.End)
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rangeMap := buildPSIDRangeMap()
			tc.check(t, rangeMap)
		})
	}
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

// Note: TestCheckInDir_NoInternalAppsDir cannot be implemented via CheckInDir because
// createTemplateMigrationsDirStub necessarily creates internal/apps/ (the template dir
// lives under it). The equivalent coverage is provided by
// TestFindDomainMigrationDirs "no apps dir returns error" which tests the function directly.

// Note: checkDir error path (non-IsNotExist ReadDir failure) and
// findDomainMigrationDirsWithPSID absErr path are structural ceilings:
// on Windows, os.ReadDir on a file returns os.IsNotExist (treated as "no dir"),
// and filepath.Abs never fails for valid paths. These OS-level failure paths
// cannot be triggered without OS-level intervention.
