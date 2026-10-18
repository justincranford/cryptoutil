// Copyright (c) 2025-2026 Justin Cranford.
package migration_range_compliance

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	testAppsDirectoryName   = "apps"
	testCallbackPathMarker  = "callback"
	testThirdSequenceNumber = 3
)

func TestCheckAndCheckDirErrors(t *testing.T) {
	t.Parallel()

	err := Check(newTestLogger())
	require.ErrorContains(t, err, "checking template migrations")

	_, err = checkDir(filepath.Join(t.TempDir(), "missing"), 1, 2, false)
	require.ErrorContains(t, err, "directory not found")

	fileAsDirectory := filepath.Join(t.TempDir(), "not-a-directory")
	require.NoError(t, os.WriteFile(fileAsDirectory, nil, cryptoutilSharedMagic.CacheFilePermissions))
	_, err = checkDir(fileAsDirectory, 1, 2, false)
	require.ErrorContains(t, err, "directory not found")

	_, err = checkDirWithReadDir(fileAsDirectory, 1, 2, false, func(string) ([]os.DirEntry, error) {
		return nil, errors.New("injected read failure")
	})
	require.ErrorContains(t, err, "read dir")
}

func TestFindDomainMigrationDirsWithPSID_Filtering(t *testing.T) {
	t.Parallel()

	appsDir := filepath.Join(t.TempDir(), "internal", testAppsDirectoryName)
	require.NoError(t, os.MkdirAll(appsDir, cryptoutilSharedMagic.DirPermissions))
	templateDir := filepath.Join(appsDir, cryptoutilSharedMagic.FrameworkProductName, "service", cryptoutilSharedMagic.CMD_SERVER, "repository", "migrations")
	identityDir := filepath.Join(appsDir, cryptoutilSharedMagic.OTLPServiceIdentityIDP, "repository", "migrations")
	serviceDir := filepath.Join(appsDir, cryptoutilSharedMagic.OTLPServiceSMKMS, "repository", "migrations")
	archivedDir := filepath.Join(appsDir, "_archived", "legacy", "migrations")
	templateUnderApps := filepath.Join(appsDir, cryptoutilSharedMagic.FrameworkProductName, "service", cryptoutilSharedMagic.CMD_SERVER, "repository", "migrations")

	for _, dir := range []string{templateDir, identityDir, serviceDir, archivedDir, templateUnderApps} {
		require.NoError(t, os.MkdirAll(dir, cryptoutilSharedMagic.DirPermissions))
	}

	entries, err := findDomainMigrationDirsWithPSID(appsDir, templateUnderApps)

	require.NoError(t, err)
	require.Equal(t, []migrationDirEntry{{dir: serviceDir, psID: cryptoutilSharedMagic.OTLPServiceSMKMS}}, entries)
}

func TestCheckInDir_InjectedDomainErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		entries   []migrationDirEntry
		findErr   bool
		failCheck bool
		wantErr   string
	}{
		{name: "domain discovery error", findErr: true, wantErr: "finding domain migration dirs"},
		{
			name:      "unknown service directory check error",
			entries:   []migrationDirEntry{{dir: "unknown-migrations", psID: "unknown-service"}},
			failCheck: true,
			wantErr:   "checking migrations in unknown-migrations",
		},
		{
			name:      "registered service directory check error",
			entries:   []migrationDirEntry{{dir: "known-migrations", psID: cryptoutilSharedMagic.OTLPServiceSMKMS}},
			failCheck: true,
			wantErr:   "checking migrations in known-migrations",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			checkCount := 0
			err := checkInDirWithOperations(
				newTestLogger(),
				t.TempDir(),
				func(string, int, int, bool) ([]string, error) {
					checkCount++
					if tc.failCheck && checkCount == 2 {
						return nil, errors.New("injected directory check error")
					}

					return nil, nil
				},
				func(string, string) ([]migrationDirEntry, error) {
					if tc.findErr {
						return nil, errors.New("injected discovery error")
					}

					return tc.entries, nil
				},
			)

			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestFindDomainMigrationDirsWithPSID_OperationErrors(t *testing.T) {
	t.Parallel()

	appsDir := filepath.Join(t.TempDir(), testAppsDirectoryName)
	require.NoError(t, os.MkdirAll(appsDir, cryptoutilSharedMagic.DirPermissions))
	templateDir := filepath.Join(t.TempDir(), cryptoutilSharedMagic.SkeletonTemplateServiceName)
	info, err := os.Stat(appsDir)
	require.NoError(t, err)

	tests := []struct {
		name   string
		absErr string
		walk   func(string, filepath.WalkFunc) error
		want   string
	}{
		{
			name:   "template absolute path error",
			absErr: cryptoutilSharedMagic.SkeletonTemplateServiceName,
			want:   "abs template dir",
		},
		{
			name:   "apps absolute path error",
			absErr: testAppsDirectoryName,
			want:   "abs apps dir",
		},
		{
			name: "walk callback error",
			walk: func(_ string, walkFn filepath.WalkFunc) error {
				return walkFn(appsDir, nil, errors.New("injected callback walk error"))
			},
			want: "walk " + appsDir,
		},
		{
			name: "walk operation error",
			walk: func(string, filepath.WalkFunc) error {
				return errors.New("injected walk error")
			},
			want: "injected walk error",
		},
		{
			name:   "callback absolute path error",
			absErr: testCallbackPathMarker,
			walk: func(_ string, walkFn filepath.WalkFunc) error {
				return walkFn(appsDir, info, nil)
			},
			want: "abs path",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			absCalls := 0
			absFn := func(path string) (string, error) {
				absCalls++
				switch {
				case tc.absErr == cryptoutilSharedMagic.SkeletonTemplateServiceName && absCalls == 1:
					return "", errors.New("injected template absolute error")
				case tc.absErr == testAppsDirectoryName && absCalls == 2:
					return "", errors.New("injected apps absolute error")
				case tc.absErr == testCallbackPathMarker && absCalls == testThirdSequenceNumber:
					return "", errors.New("injected callback absolute error")
				default:
					return filepath.Abs(path)
				}
			}

			walkFn := filepath.Walk
			if tc.walk != nil {
				walkFn = tc.walk
			}

			_, err := findDomainMigrationDirsWithOperations(appsDir, templateDir, os.Stat, absFn, walkFn)

			require.ErrorContains(t, err, tc.want)
		})
	}
}

func TestCheckDirInvalidVersionOverflow(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filename := "999999999999999999999999999999999999_init.up.sql"
	require.NoError(t, os.WriteFile(filepath.Join(dir, filename), nil, cryptoutilSharedMagic.CacheFilePermissions))

	violations, err := checkDir(dir, 1, 0, false)

	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Contains(t, violations[0], "failed to parse version")
}

func TestCheckInDir_MissingTemplateMigrations(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(rootDir, "internal", testAppsDirectoryName), cryptoutilSharedMagic.DirPermissions))

	err := CheckInDir(cryptoutilCmdCicdCommon.NewLogger("migration-range-compliance"), rootDir)

	require.ErrorContains(t, err, "checking template migrations")
}
