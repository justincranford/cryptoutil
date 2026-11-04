// Copyright (c) 2025-2026 Justin Cranford.
package if_else_chain_test

import (
	"os"
	"path/filepath"
	"testing"

	lintFitnessIfElseChain "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/if_else_chain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const excludedDirectoryIfElseViolation = `package fixture
func check(x int) {
	if x > 0 {
		println("first")
	}
	if x < 10 {
		println("second")
	}
}
`

func TestFindViolationsInDir_ExcludedDirectoriesDoNotContaminateScan(t *testing.T) {
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

			writeIfElseScanFixtureFile(t, realFile)
			writeIfElseScanFixtureFile(t, excludedFile)

			violations, err := lintFitnessIfElseChain.FindViolationsInDir(rootDir)

			require.NoError(t, err)
			require.Len(t, violations, 1)
			require.Equal(t, realFile, violations[0].File)
		})
	}
}

func writeIfElseScanFixtureFile(t *testing.T, path string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, []byte(excludedDirectoryIfElseViolation), 0o600))
}
