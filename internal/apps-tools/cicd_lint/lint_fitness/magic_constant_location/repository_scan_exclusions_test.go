// Copyright (c) 2025-2026 Justin Cranford.
package magic_constant_location_test

import (
	"path/filepath"
	"testing"

	lintFitnessMagicConstantLocation "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/magic_constant_location"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const excludedDirectoryMagicConstantFixture = `package fixture
const port = 8080
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

			writeFile(t, realFile, excludedDirectoryMagicConstantFixture)
			writeFile(t, excludedFile, excludedDirectoryMagicConstantFixture)

			violations, err := lintFitnessMagicConstantLocation.FindViolationsInDir(rootDir)

			require.NoError(t, err)
			require.Len(t, violations, 1)
			require.Contains(t, violations[0], filepath.ToSlash(filepath.Join("repository", "real.go")))
		})
	}
}
