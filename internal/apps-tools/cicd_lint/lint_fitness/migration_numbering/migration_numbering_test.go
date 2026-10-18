// Copyright (c) 2025-2026 Justin Cranford.
package migration_numbering

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
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

func mkdirAll(t *testing.T, dir string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(dir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
}

func writeSQL(t *testing.T, dir, filename string) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Join(dir, filename), []byte("OK"), cryptoutilSharedMagic.CacheFilePermissions))
}

func tplRelPath() string {
	return filepath.Join(cryptoutilSharedMagic.FrameworkInternalDir, "service", cryptoutilSharedMagic.CMD_SERVER, "repository", "migrations")
}

func domainRelPath(parts ...string) string {
	elems := make([]string, 0, 2+len(parts)+2)
	elems = append(elems, "internal", "apps")
	elems = append(elems, parts...)
	elems = append(elems, "repository", "migrations")

	return filepath.Join(elems...)
}

func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	require.NoError(t, CheckInDir(logger, root))
}

func TestCheckInDir(t *testing.T) {
	t.Parallel()

	tpl := tplRelPath()
	dom := domainRelPath("myproduct", "myservice")
	archivedDir := filepath.Join("internal", "apps", cryptoutilSharedMagic.SMProductName, "_ca-archived", "repository", "migrations")

	type fileEntry struct {
		relPath string
		names   []string
	}

	tests := []struct {
		name         string
		stubTemplate bool
		files        []fileEntry
		wantErr      string
	}{
		{name: "valid domain migrations", stubTemplate: true, files: []fileEntry{{dom, []string{test2001InitUpSQL, test2001InitDownSQL}}}},
		{name: "valid template migrations", files: []fileEntry{{tpl, []string{"1001_sessions.up.sql", "1001_sessions.down.sql"}}}},
		{name: "template at max boundary", files: []fileEntry{{tpl, []string{"1999_last.up.sql", "1999_last.down.sql"}}}},
		{name: "empty migrations dir", stubTemplate: true, files: []fileEntry{{dom, nil}}},
		{name: "archived dirs skipped", stubTemplate: true, files: []fileEntry{{archivedDir, []string{"0001_bad.up.sql"}}}},
		{name: "multiple domain dirs", stubTemplate: true, files: []fileEntry{
			{domainRelPath("product1", "svc1"), []string{test2001InitUpSQL, test2001InitDownSQL}},
			{domainRelPath("product2", "svc2"), []string{"2002_tables.up.sql", "2002_tables.down.sql"}},
		}},
		{name: "multiple versions", stubTemplate: true, files: []fileEntry{
			{dom, []string{test2001InitUpSQL, test2001InitDownSQL, "2002_add_column.up.sql", "2002_add_column.down.sql"}},
		}},
		{name: "domain version below minimum", files: []fileEntry{{dom, []string{"0001_init.up.sql", "0001_init.down.sql"}}}, wantErr: "below minimum 2001"},
		{name: "template version above maximum", files: []fileEntry{{tpl, []string{"2001_too_high.up.sql", "2001_too_high.down.sql"}}}, wantErr: testExceedsMaximum1999},
		{name: "missing down file", files: []fileEntry{{dom, []string{test2001InitUpSQL}}}, wantErr: "missing .down.sql"},
		{name: "missing up file", files: []fileEntry{{dom, []string{test2001InitDownSQL}}}, wantErr: "missing .up.sql"},
		{name: "invalid filename", files: []fileEntry{{dom, []string{"bad_file.sql"}}}, wantErr: "does not match migration naming pattern"},
		{name: "no apps dir", wantErr: testInternalAppsDirectoryNotFound},
		{name: "template below minimum", files: []fileEntry{{tpl, []string{"0500_too_low.up.sql", "0500_too_low.down.sql"}}}, wantErr: "below minimum 1001"},
		{name: "template exceeds max", files: []fileEntry{{tpl, []string{"2500_above.up.sql", "2500_above.down.sql"}}}, wantErr: testExceedsMaximum1999},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			if tc.stubTemplate {
				mkdirAll(t, filepath.Join(tmpDir, tpl))
			}

			// Create internal/apps/ dir so findDomainMigrationDirs does not fail before
			// the template or domain checks can return their violations — except for the
			// "no apps dir" test case which explicitly exercises that error path.
			if tc.wantErr != testInternalAppsDirectoryNotFound {
				mkdirAll(t, filepath.Join(tmpDir, "internal", "apps"))
			}

			for _, f := range tc.files {
				dir := filepath.Join(tmpDir, f.relPath)
				mkdirAll(t, dir)

				for _, name := range f.names {
					writeSQL(t, dir, name)
				}
			}

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			err := CheckInDir(logger, tmpDir)

			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckMigrationDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupDir   func(t *testing.T) string
		minVersion int
		maxVersion int
		isTemplate bool
		wantErr    string
	}{
		{
			name: "subdirectories ignored",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := filepath.Join(t.TempDir(), "migrations")
				mkdirAll(t, dir)
				mkdirAll(t, filepath.Join(dir, "subdir"))
				writeSQL(t, dir, test2001InitUpSQL)
				writeSQL(t, dir, test2001InitDownSQL)

				return dir
			},
			minVersion: domainMigrationMin,
		},
		{
			name: "nonexistent directory",
			setupDir: func(t *testing.T) string {
				t.Helper()

				return filepath.Join(t.TempDir(), "nonexistent")
			},
			minVersion: domainMigrationMin,
			wantErr:    "does not exist",
		},
		{
			name: "version exceeds maximum",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := filepath.Join(t.TempDir(), "migrations")
				mkdirAll(t, dir)
				writeSQL(t, dir, "9999_too_high.up.sql")
				writeSQL(t, dir, "9999_too_high.down.sql")

				return dir
			},
			minVersion: templateMigrationMin,
			maxVersion: templateMigrationMax,
			isTemplate: true,
			wantErr:    "exceeds maximum",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := tc.setupDir(t)
			errs := checkMigrationDir(dir, tc.minVersion, tc.maxVersion, tc.isTemplate)

			if tc.wantErr == "" {
				require.Empty(t, errs)
			} else {
				require.NotEmpty(t, errs)
				require.Contains(t, errs[0], tc.wantErr)
			}
		})
	}
}

func TestFindDomainMigrationDirs_NoAppsDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	dirs, err := findDomainMigrationDirs(tmpDir, filepath.Join(tmpDir, tplRelPath()))
	require.Error(t, err)
	require.Contains(t, err.Error(), testInternalAppsDirectoryNotFound)
	require.Empty(t, dirs)
}

func TestFindDomainMigrationDirs_LegacyExcluded(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	legacyDir := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.IdentityProductName, "repository", "migrations")
	mkdirAll(t, legacyDir)
	writeSQL(t, legacyDir, "0001_legacy.up.sql")

	dir := filepath.Join(tmpDir, domainRelPath("myservice"))
	mkdirAll(t, dir)
	writeSQL(t, dir, test2001InitUpSQL)
	writeSQL(t, dir, test2001InitDownSQL)

	dirs, err := findDomainMigrationDirs(tmpDir, filepath.Join(tmpDir, tplRelPath()))
	require.NoError(t, err)
	require.Len(t, dirs, 1)
	require.Contains(t, dirs[0], "myservice")
}

func TestMigrationNumberingOperations_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		run     func(*testing.T) error
		wantErr string
	}{
		{
			name: "template absolute path error",
			run: func(t *testing.T) error {
				operations := productionMigrationNumberingOperations()
				operations.pathAbsFn = func(string) (string, error) {
					return "", fmt.Errorf("injected template absolute path error")
				}

				tmpDir := t.TempDir()

				_, err := findDomainMigrationDirsWithOperations(tmpDir, tplRelPath(), operations)

				return err
			},
			wantErr: "failed to get absolute path for template dir",
		},
		{
			name: "legacy absolute path error",
			run: func(t *testing.T) error {
				operations := productionMigrationNumberingOperations()
				tmpDir := t.TempDir()
				legacyDir := filepath.Join(tmpDir, legacyMigrationPaths[0])
				operations.pathAbsFn = func(path string) (string, error) {
					if path == legacyDir {
						return "", fmt.Errorf("injected legacy absolute path error")
					}

					return filepath.Abs(path)
				}

				_, err := findDomainMigrationDirsWithOperations(tmpDir, filepath.Join(tmpDir, tplRelPath()), operations)

				return err
			},
			wantErr: "failed to get absolute path for legacy migration dir",
		},
		{
			name: "apps stat error",
			run: func(t *testing.T) error {
				operations := productionMigrationNumberingOperations()
				operations.statFn = func(string) (os.FileInfo, error) {
					return nil, fmt.Errorf("injected apps stat error")
				}
				tmpDir := t.TempDir()

				_, err := findDomainMigrationDirsWithOperations(tmpDir, filepath.Join(tmpDir, tplRelPath()), operations)

				return err
			},
			wantErr: "failed to stat internal/apps directory",
		},
		{
			name: "walk error",
			run: func(t *testing.T) error {
				operations := productionMigrationNumberingOperations()
				operations.walkFn = func(string, filepath.WalkFunc) error {
					return fmt.Errorf("injected walk error")
				}
				tmpDir := t.TempDir()
				mkdirAll(t, filepath.Join(tmpDir, "internal", "apps"))

				_, err := findDomainMigrationDirsWithOperations(tmpDir, filepath.Join(tmpDir, tplRelPath()), operations)

				return err
			},
			wantErr: migrationNumberingWalkAppsDirectoryErrorPrefix,
		},
		{
			name: "walk callback error",
			run: func(t *testing.T) error {
				operations := productionMigrationNumberingOperations()
				operations.walkFn = func(path string, walkFn filepath.WalkFunc) error {
					info, err := os.Stat(path)
					require.NoError(t, err)

					return walkFn(path, info, os.ErrPermission)
				}
				tmpDir := t.TempDir()
				mkdirAll(t, filepath.Join(tmpDir, "internal", "apps"))

				_, err := findDomainMigrationDirsWithOperations(tmpDir, filepath.Join(tmpDir, tplRelPath()), operations)

				return err
			},
			wantErr: migrationNumberingWalkAppsDirectoryErrorPrefix,
		},
		{
			name: "migration directory absolute path error",
			run: func(t *testing.T) error {
				operations := productionMigrationNumberingOperations()
				tmpDir := t.TempDir()
				migrationDir := filepath.Join(tmpDir, domainRelPath("myservice"))
				mkdirAll(t, migrationDir)

				operations.pathAbsFn = func(path string) (string, error) {
					if path == migrationDir {
						return "", fmt.Errorf("injected migration directory absolute path error")
					}

					return filepath.Abs(path)
				}

				_, err := findDomainMigrationDirsWithOperations(tmpDir, filepath.Join(tmpDir, tplRelPath()), operations)

				return err
			},
			wantErr: migrationNumberingWalkAppsDirectoryErrorPrefix,
		},
		{
			name: "check in directory wraps domain discovery error",
			run: func(t *testing.T) error {
				operations := productionMigrationNumberingOperations()
				operations.pathAbsFn = func(string) (string, error) {
					return "", fmt.Errorf("injected domain discovery error")
				}
				tmpDir := t.TempDir()
				mkdirAll(t, filepath.Join(tmpDir, tplRelPath()))

				return checkInDirWithOperations(cryptoutilCmdCicdCommon.NewLogger("test"), tmpDir, operations)
			},
			wantErr: "failed to find domain migration directories",
		},
		{
			name: "read directory error",
			run: func(t *testing.T) error {
				operations := productionMigrationNumberingOperations()
				operations.readDirFn = func(string) ([]os.DirEntry, error) {
					return nil, fmt.Errorf("injected read directory error")
				}
				errs := checkMigrationDirWithOperations(t.TempDir(), domainMigrationMin, 0, false, operations)

				return fmt.Errorf("%s", errs[0])
			},
			wantErr: "failed to read directory",
		},
		{
			name: "version parser error",
			run: func(t *testing.T) error {
				operations := productionMigrationNumberingOperations()
				operations.atoiFn = func(string) (int, error) {
					return 0, fmt.Errorf("injected integer parsing error")
				}
				migrationsDir := filepath.Join(t.TempDir(), "migrations")
				mkdirAll(t, migrationsDir)
				writeSQL(t, migrationsDir, test2001InitUpSQL)

				errs := checkMigrationDirWithOperations(migrationsDir, domainMigrationMin, 0, false, operations)

				return fmt.Errorf("%s", errs[0])
			},
			wantErr: "failed to parse version number",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.run(t)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestFindDomainMigrationDirs_ExcludesTemplateMigrationDirectory(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	templateDir := filepath.Join(rootDir, "internal", "apps", cryptoutilSharedMagic.SkeletonTemplateServiceName, "migrations")
	mkdirAll(t, templateDir)

	dirs, err := findDomainMigrationDirsWithOperations(rootDir, templateDir, productionMigrationNumberingOperations())
	require.NoError(t, err)
	require.Empty(t, dirs)
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_Integration(t *testing.T) {
	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(root))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	require.NoError(t, Check(logger))
}
