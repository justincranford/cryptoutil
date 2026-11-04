// Copyright (c) 2025-2026 Justin Cranford.
package insecure_skip_verify

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const excludedDirectoryInsecureTLSFixture = `package fixture
var config = &tls.Config{InsecureSkipVerify: true}
`

func TestFindInsecureSkipVerifyViolationsInDir_ExcludedDirectoriesDoNotContaminateScan(t *testing.T) {
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
			realFile := filepath.Join(rootDir, "repository", "real.go")
			excludedFile := filepath.Join(rootDir, "nested", cryptoutilSharedMagic.CICDExcludeDirWorktree, tc.dirName, "nested", "excluded.go")

			writeInsecureTLSScanFixtureFile(t, realFile)
			writeInsecureTLSScanFixtureFile(t, excludedFile)

			violations, err := FindInsecureSkipVerifyViolationsInDir(rootDir)

			require.NoError(t, err)
			require.Len(t, violations, 1)
			require.Equal(t, realFile, violations[0].File)
		})
	}
}

func writeInsecureTLSScanFixtureFile(t *testing.T, path string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.WriteFile(path, []byte(excludedDirectoryInsecureTLSFixture), cryptoutilSharedMagic.FilePermissionsDefault))
}
