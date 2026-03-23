// Copyright (c) 2025 Justin Cranford

package archive_detector

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

func mkdir(t *testing.T, path string) {
	t.Helper()

	err := os.MkdirAll(path, cryptoutilSharedMagic.DirPermissions)
	require.NoError(t, err)
}

func TestCheckInDir_EmptyDir_Passes(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_NoBannedDirs_Passes(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	mkdir(t, filepath.Join(tmp, "internal", "apps", "sm"))
	mkdir(t, filepath.Join(tmp, "deployments", cryptoutilSharedMagic.OTLPServiceSMIM))
	mkdir(t, filepath.Join(tmp, cryptoutilSharedMagic.CICDConfigsDir, "sm", "im"))

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_ArchivedDir_Detected(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	mkdir(t, filepath.Join(tmp, "archived"))

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "archived/orphaned directories")
}

func TestCheckInDir_UnderscoreArchivedDir_Detected(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	mkdir(t, filepath.Join(tmp, "internal", "_archived"))

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "archived/orphaned directories")
}

func TestCheckInDir_OrphanedDir_Detected(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	mkdir(t, filepath.Join(tmp, "deployments", "orphaned"))

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "archived/orphaned directories")
}

func TestCheckInDir_NestedBannedDir_Detected(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	mkdir(t, filepath.Join(tmp, "internal", "apps", "sm", "_archived", "old-service"))

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "archived/orphaned directories")
}

func TestCheckInDir_MultipleBannedDirs_ReportsAll(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	mkdir(t, filepath.Join(tmp, "archived"))
	mkdir(t, filepath.Join(tmp, "orphaned"))

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "2 archived/orphaned directories")
}

func TestCheckInDir_PartialNameMatch_Passes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		dir  string
	}{
		{"archived-prefix", filepath.Join("internal", "archived-code")},
		{"archived-suffix", filepath.Join("internal", "code-archived")},
		{"orphaned-prefix", filepath.Join("deployments", "orphaned-service")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmp := t.TempDir()
			mkdir(t, filepath.Join(tmp, tc.dir))

			err := CheckInDir(newTestLogger(), tmp)
			require.NoError(t, err)
		})
	}
}

func TestCheckInDir_NonExistentRoot_ReturnsError(t *testing.T) {
	t.Parallel()

	err := CheckInDir(newTestLogger(), "/nonexistent/path/that/does/not/exist")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to check for archived directories")
}

func TestFindViolationsInDir_EmptyDir_NoViolations(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	violations, err := FindViolationsInDir(tmp)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestFindViolationsInDir_BannedDir_ReturnsPath(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	bannedPath := filepath.Join(tmp, "_archived")
	mkdir(t, bannedPath)

	violations, err := FindViolationsInDir(tmp)
	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Equal(t, bannedPath, violations[0])
}

func TestFindViolationsInDir_WalkError_NonExistentRoot_ReturnsError(t *testing.T) {
	t.Parallel()

	_, err := FindViolationsInDir("/nonexistent/path/that/does/not/exist")
	require.Error(t, err)
}

func TestFindViolationsInDir_WalkError_PermissionDenied_ReturnsError(t *testing.T) {
	t.Parallel()

	if isWindows() {
		t.Skip("os.Chmod 0o000 does not restrict directory access on Windows NTFS")
	}

	tmp := t.TempDir()
	lockedDir := filepath.Join(tmp, "locked")
	mkdir(t, lockedDir)
	innerDir := filepath.Join(lockedDir, "inner")
	mkdir(t, innerDir)

	err := os.Chmod(lockedDir, 0o000)
	require.NoError(t, err)

	t.Cleanup(func() { _ = os.Chmod(lockedDir, 0o700) })

	_, err = FindViolationsInDir(tmp)
	require.Error(t, err)
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

	err = Check(newTestLogger())
	require.NoError(t, err)
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
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

func isWindows() bool {
	return os.PathSeparator == '\\'
}
