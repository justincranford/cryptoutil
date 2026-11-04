// Copyright (c) 2025-2026 Justin Cranford.
package no_local_ps_id_server_stubs_test

import (
	"os"
	"path/filepath"
	"testing"

	lintFitnessNoLocalPSIDServerStubs "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/no_local_ps_id_server_stubs"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestFindViolations_ExcludedDirectoriesDoNotContaminateScan(t *testing.T) {
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
			realFile := filepath.Join(rootDir, "repository", "internal", "apps", "real-service", cryptoutilSharedMagic.CMD_SERVER, "server_test.go")
			excludedFile := filepath.Join(rootDir, "nested", cryptoutilSharedMagic.CICDExcludeDirWorktree, tc.dirName, "internal", "apps", "excluded-service", cryptoutilSharedMagic.CMD_SERVER, "server_test.go")

			writeLocalServerStubScanFixtureFile(t, realFile)
			writeLocalServerStubScanFixtureFile(t, excludedFile)

			violations, err := lintFitnessNoLocalPSIDServerStubs.FindViolations(rootDir)

			require.NoError(t, err)
			require.Len(t, violations, 1)
			require.Equal(t, realFile, violations[0].File)
		})
	}
}

func writeLocalServerStubScanFixtureFile(t *testing.T, path string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.WriteFile(path, []byte(stubPublicServerGoSrc), cryptoutilSharedMagic.FilePermissionsDefault))
}
