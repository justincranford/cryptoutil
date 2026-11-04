// Copyright (c) 2025-2026 Justin Cranford.
package no_unaliased_cryptoutil_imports

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const unaliasedImportScanFixture = "package fixture\n" +
	"import (" + "\n\t\"cryptoutil/" +
	"internal/shared/magic\"\n)\nvar _ = magic.CICDExcludeDirGit\n"

func TestFindUnaliasedCryptoutilImports_ExcludedDirectoriesDoNotContaminateScan(t *testing.T) {
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

			writeUnaliasedImportScanFixtureFile(t, realFile)
			writeUnaliasedImportScanFixtureFile(t, excludedFile)

			walkFn := func(_ string, walkFn filepath.WalkFunc) error {
				return filepath.Walk(rootDir, walkFn)
			}
			violations, err := FindUnaliasedCryptoutilImports(walkFn)

			require.NoError(t, err)
			require.Len(t, violations, 1)
			require.Contains(t, violations[0], realFile)
		})
	}
}

func writeUnaliasedImportScanFixtureFile(t *testing.T, path string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, []byte(unaliasedImportScanFixture), 0o600))
}
