// Copyright (c) 2025-2026 Justin Cranford.
package no_postgres_in_non_e2e

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestCheckInDir_ExcludedDirectoriesDoNotHideRealViolation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dirName string
	}{
		{name: cryptoutilSharedMagic.CICDExcludeDirTestOutput, dirName: cryptoutilSharedMagic.CICDExcludeDirTestOutput},
		{name: cryptoutilSharedMagic.CICDExcludeDirVendor, dirName: cryptoutilSharedMagic.CICDExcludeDirVendor},
		{name: "git", dirName: cryptoutilSharedMagic.CICDExcludeDirGit},
		{name: "Python virtual environment", dirName: cryptoutilSharedMagic.CICDExcludeDirVenv},
		{name: "Node.js dependencies", dirName: cryptoutilSharedMagic.CICDExcludeDirNodeModules},
		{name: "workflow reports", dirName: cryptoutilSharedMagic.CICDExcludeDirWorkflowReports},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := t.TempDir()
			realFile := filepath.Join(rootDir, "repository", "real_test.go")
			excludedFile := filepath.Join(rootDir, "nested", cryptoutilSharedMagic.CICDExcludeDirWorktree, tc.dirName, "nested", "excluded_test.go")

			writeNoPostgresScanFixtureFile(t, realFile, postgresRunContainerViolation)
			writeNoPostgresScanFixtureFile(t, excludedFile, postgresRunContainerViolation)

			err := CheckInDir(cryptoutilCmdCicdCommon.NewLogger(t.Name()), rootDir)

			require.ErrorContains(t, err, "found 1 PostgreSQL container violation(s)")
		})
	}
}

func writeNoPostgresScanFixtureFile(t *testing.T, path, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))
}
