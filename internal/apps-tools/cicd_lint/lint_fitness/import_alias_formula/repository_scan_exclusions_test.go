// Copyright (c) 2025-2026 Justin Cranford.
package import_alias_formula

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const wrongImportAliasScanFixture = `package fixture
import (
	json "encoding/json"
)
var _ = json.Marshal
`

func TestCheckInDir_ExcludedDirectoriesDoNotContaminateScan(t *testing.T) {
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

			rootDir := buildAliasRoot(t)
			writeGoFile(t, filepath.Join(rootDir, "repository"), "real.go", wrongImportAliasScanFixture)
			writeGoFile(t, filepath.Join(rootDir, "nested", cryptoutilSharedMagic.CICDExcludeDirWorktree, tc.dirName, "nested"), "excluded.go", wrongImportAliasScanFixture)

			err := CheckInDir(
				cryptoutilCmdCicdCommon.NewLogger(t.Name()),
				rootDir,
				os.ReadFile,
				filepath.WalkDir,
			)

			require.ErrorContains(t, err, "1 violation(s) found")
		})
	}
}
