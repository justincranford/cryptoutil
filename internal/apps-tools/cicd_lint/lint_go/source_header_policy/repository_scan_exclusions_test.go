// Copyright (c) 2025-2026 Justin Cranford.
package source_header_policy

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestFindViolations_ExcludesRepositoryScanDirectoriesAndFindsRealViolation(t *testing.T) {
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
			currentYear := time.Now().UTC().Year()
			invalidHeader := fmt.Sprintf(
				"// Copyright (c) %d Justin Cranford.\n// SPDX-License-Identifier: MIT\n\npackage fixture\n",
				currentYear,
			)
			realFile := filepath.Join(rootDir, "repository", "real.go")
			excludedFile := filepath.Join(rootDir, "nested", cryptoutilSharedMagic.CICDExcludeDirWorktree, tc.dirName, "nested", "excluded.go")

			writeRepositoryScanFixtureFile(t, realFile, invalidHeader)
			writeRepositoryScanFixtureFile(t, excludedFile, invalidHeader)

			violations, err := findViolations(rootDir, currentYear)

			require.NoError(t, err)
			require.Len(t, violations, 1)
			require.Contains(t, violations[0], realFile)
		})
	}
}

func writeRepositoryScanFixtureFile(t *testing.T, path, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))
}
