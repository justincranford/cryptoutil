// Copyright (c) 2025 Justin Cranford

package legacy_dir_detection

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

func TestCheckInDir_NoLegacyDirs_Passes(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	mkdir(t, filepath.Join(tmp, "internal", "apps", "sm"))
	mkdir(t, filepath.Join(tmp, "deployments", cryptoutilSharedMagic.OTLPServiceSMIM))
	mkdir(t, filepath.Join(tmp, cryptoutilSharedMagic.CICDConfigsDir, "sm", "im"))
	mkdir(t, filepath.Join(tmp, "cmd", cryptoutilSharedMagic.OTLPServiceSMIM))

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_InternalAppsCipher_Detected(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	mkdir(t, filepath.Join(tmp, "internal", "apps", "cipher"))

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "legacy directories")
}

func TestCheckInDir_CipherPrefixInDeployments(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	mkdir(t, filepath.Join(tmp, "deployments", "cipher-im"))

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "legacy directories")
}

func TestCheckInDir_CipherPrefixInConfigs(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	mkdir(t, filepath.Join(tmp, cryptoutilSharedMagic.CICDConfigsDir, "cipher-kms"))

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "legacy directories")
}

func TestCheckInDir_CipherPrefixInCmd(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	mkdir(t, filepath.Join(tmp, "cmd", "cipher-app"))

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "legacy directories")
}

func TestCheckInDir_NonCipherPrefixNotDetected(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		dir  string
	}{
		{
			name: "sm-im in deployments",
			dir:  filepath.Join("deployments", cryptoutilSharedMagic.OTLPServiceSMIM),
		},
		{
			name: "cipher-less name in cmd",
			dir:  filepath.Join("cmd", cryptoutilSharedMagic.OTLPServiceSMKMS),
		},
		{
			name: "decipher prefix not banned",
			dir:  filepath.Join("deployments", "decipher-tool"),
		},
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

func TestFindViolationsInDir_MissingScanDir_Passes(t *testing.T) {
	t.Parallel()

	// If deployments/ does not exist, no violation should be reported.
	tmp := t.TempDir()

	violations, err := FindViolationsInDir(tmp)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestFindViolationsInDir_ReadDirError(t *testing.T) {
	t.Parallel()

	// If we cannot read a scan dir, FindViolationsInDir should return an error.
	// On Windows os.Chmod 0o000 does not restrict access, so skip.
	if isWindows() {
		t.Skip("os.Chmod 0o000 does not restrict access on Windows NTFS")
	}

	tmp := t.TempDir()
	scanDir := filepath.Join(tmp, "deployments")
	err := os.MkdirAll(scanDir, 0o000)
	require.NoError(t, err)

	t.Cleanup(func() { _ = os.Chmod(scanDir, 0o700) })

	violations, err := FindViolationsInDir(tmp)
	require.Error(t, err)
	require.Nil(t, violations)
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
