// Copyright (c) 2025 Justin Cranford

package migration_numbering

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
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
	return filepath.Join("internal", "apps", cryptoutilSharedMagic.FrameworkProductName, "service", "server", "repository", "migrations")
}

func domainRelPath(parts ...string) string {
	elems := []string{"internal", "apps"}
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
	archivedDir := filepath.Join("internal", "apps", cryptoutilSharedMagic.PKIProductName, "_ca-archived", "repository", "migrations")

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
		{name: "valid domain migrations", stubTemplate: true, files: []fileEntry{{dom, []string{"2001_init.up.sql", "2001_init.down.sql"}}}},
		{name: "valid template migrations", files: []fileEntry{{tpl, []string{"1001_sessions.up.sql", "1001_sessions.down.sql"}}}},
		{name: "template at max boundary", files: []fileEntry{{tpl, []string{"1999_last.up.sql", "1999_last.down.sql"}}}},
		{name: "empty migrations dir", stubTemplate: true, files: []fileEntry{{dom, nil}}},
		{name: "archived dirs skipped", stubTemplate: true, files: []fileEntry{{archivedDir, []string{"0001_bad.up.sql"}}}},
		{name: "multiple domain dirs", stubTemplate: true, files: []fileEntry{
			{domainRelPath("product1", "svc1"), []string{"2001_init.up.sql", "2001_init.down.sql"}},
			{domainRelPath("product2", "svc2"), []string{"2002_tables.up.sql", "2002_tables.down.sql"}},
		}},
		{name: "multiple versions", stubTemplate: true, files: []fileEntry{
			{dom, []string{"2001_init.up.sql", "2001_init.down.sql", "2002_add_column.up.sql", "2002_add_column.down.sql"}},
		}},
		{name: "domain version below minimum", files: []fileEntry{{dom, []string{"0001_init.up.sql", "0001_init.down.sql"}}}, wantErr: "below minimum 2001"},
		{name: "template version above maximum", files: []fileEntry{{tpl, []string{"2001_too_high.up.sql", "2001_too_high.down.sql"}}}, wantErr: "exceeds maximum 1999"},
		{name: "missing down file", files: []fileEntry{{dom, []string{"2001_init.up.sql"}}}, wantErr: "missing .down.sql"},
		{name: "missing up file", files: []fileEntry{{dom, []string{"2001_init.down.sql"}}}, wantErr: "missing .up.sql"},
		{name: "invalid filename", files: []fileEntry{{dom, []string{"bad_file.sql"}}}, wantErr: "does not match migration naming pattern"},
		{name: "no apps dir", wantErr: "internal/apps directory not found"},
		{name: "template below minimum", files: []fileEntry{{tpl, []string{"0500_too_low.up.sql", "0500_too_low.down.sql"}}}, wantErr: "below minimum 1001"},
		{name: "template exceeds max", files: []fileEntry{{tpl, []string{"2500_above.up.sql", "2500_above.down.sql"}}}, wantErr: "exceeds maximum 1999"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			if tc.stubTemplate {
				mkdirAll(t, filepath.Join(tmpDir, tpl))
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
				writeSQL(t, dir, "2001_init.up.sql")
				writeSQL(t, dir, "2001_init.down.sql")

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

func TestCheckMigrationDir_ReadDirError(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("os.Chmod 0o000 does not restrict access on Windows NTFS")
	}

	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "migrations")
	mkdirAll(t, migrationsDir)
	writeSQL(t, migrationsDir, "2001_init.up.sql")
	require.NoError(t, os.Chmod(migrationsDir, 0o000))

	t.Cleanup(func() {
		_ = os.Chmod(migrationsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
	})

	errs := checkMigrationDir(migrationsDir, domainMigrationMin, 0, false)
	require.NotEmpty(t, errs)
	require.Contains(t, errs[0], "failed to read directory")
}

func TestFindDomainMigrationDirs_NoAppsDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	dirs, err := findDomainMigrationDirs(tmpDir, filepath.Join(tmpDir, tplRelPath()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "internal/apps directory not found")
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
	writeSQL(t, dir, "2001_init.up.sql")
	writeSQL(t, dir, "2001_init.down.sql")

	dirs, err := findDomainMigrationDirs(tmpDir, filepath.Join(tmpDir, tplRelPath()))
	require.NoError(t, err)
	require.Len(t, dirs, 1)
	require.Contains(t, dirs[0], "myservice")
}

func TestFindDomainMigrationDirs_WalkPermissionError(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("os.Chmod 0o000 does not restrict access on Windows NTFS")
	}

	tmpDir := t.TempDir()
	badDir := filepath.Join(tmpDir, "internal", "apps", "badservice")
	mkdirAll(t, badDir)
	require.NoError(t, os.Chmod(badDir, 0o000))

	t.Cleanup(func() {
		_ = os.Chmod(badDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
	})

	_, err := findDomainMigrationDirs(tmpDir, filepath.Join(tmpDir, tplRelPath()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to walk apps directory")
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

// Sequential: modifies package-level seam variables.
func TestSeam_FindDomainDirs_AbsTemplateError(t *testing.T) {
	saveRestoreSeams(t)

	pathAbsFunc = func(_ string) (string, error) {
		return "", fmt.Errorf("injected abs error")
	}

	tmpDir := t.TempDir()

	_, err := findDomainMigrationDirs(tmpDir, filepath.Join(tmpDir, tplRelPath()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get absolute path for template dir")
}

// Sequential: modifies package-level seam variables.
func TestSeam_FindDomainDirs_WalkAbsError(t *testing.T) {
	saveRestoreSeams(t)

	callCount := 0

	pathAbsFunc = func(path string) (string, error) {
		callCount++
		if callCount > 3 {
			return "", fmt.Errorf("injected walk abs error")
		}

		return filepath.Abs(path)
	}

	tmpDir := t.TempDir()
	migDir := filepath.Join(tmpDir, domainRelPath("myservice"))
	mkdirAll(t, migDir)
	writeSQL(t, migDir, "2001_init.up.sql")

	tplDir := filepath.Join(tmpDir, tplRelPath())
	mkdirAll(t, tplDir)

	_, err := findDomainMigrationDirs(tmpDir, tplDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to walk apps directory")
}

// Sequential: modifies package-level seam variables.
func TestSeam_CheckInDir_FindDomainDirsError(t *testing.T) {
	saveRestoreSeams(t)

	pathAbsFunc = func(_ string) (string, error) {
		return "", fmt.Errorf("injected abs error")
	}

	tmpDir := t.TempDir()
	mkdirAll(t, filepath.Join(tmpDir, tplRelPath()))

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
	mkdirAll(t, migrationsDir)
	writeSQL(t, migrationsDir, "2001_init.up.sql")

	errs := checkMigrationDir(migrationsDir, domainMigrationMin, 0, false)
	require.NotEmpty(t, errs)
	require.Contains(t, errs[0], "failed to parse version number")
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
