// Copyright (c) 2025 Justin Cranford

package magic_e2e_container_names_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessMagicE2EContainerNames "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/magic_e2e_container_names"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

func findProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err, "failed to get working directory")

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("skipping integration test: cannot find project root (no go.mod)")
		}

		dir = parent
	}
}

// writeMagicFile writes a magic Go source file under internal/shared/magic/filename.
func writeMagicFile(t *testing.T, tmpDir, filename, content string) {
	t.Helper()

	magicDir := filepath.Join(tmpDir, "internal", "shared", "magic")
	require.NoError(t, os.MkdirAll(magicDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(magicDir, filename), []byte(content), cryptoutilSharedMagic.FilePermissions))
}

// correctMagicSource generates a magic Go source with correct container name constants for psID.
func correctMagicSource(psID string) string {
	return "package magic\nconst (\n" +
		"	TestE2ESQLiteContainer = \"" + psID + "-app-sqlite-1\"\n" +
		"	TestE2EPostgreSQL1Container = \"" + psID + "-app-postgres-1\"\n" +
		"	TestE2EPostgreSQL2Container = \"" + psID + "-app-postgres-2\"\n" +
		")\n"
}

func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)

	err := lintFitnessMagicE2EContainerNames.CheckInDir(newTestLogger(), root)
	require.NoError(t, err)
}

func TestCheckInDir_AllCorrect(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Write magic_sm_im.go with correct container constants for sm-im.
	writeMagicFile(t, tmpDir, "magic_sm_im.go", correctMagicSource(cryptoutilSharedMagic.OTLPServiceSMIM))

	// The test uses only the magic files referenced by the registry — inject a minimal registry
	// by specifying the MagicFile name that the checker will scan.
	// Because the checker iterates the real registry, fake magic files must match one of the
	// MagicFile names in the registry and contain valid constants.
	// We only write magic_sm_im.go containing sm-im constants; all other PS will either find no
	// *E2ESQLiteContainer constant (identity, pki-ca) or their magic file won't exist, causing an error.
	// Instead, write all required magic files with NO E2ESQLiteContainer to pass cleanly.
	for _, mf := range []string{"magic_identity.go", "magic_jose.go", "magic_skeleton.go", "magic_sm.go", "magic_pki.go"} {
		writeMagicFile(t, tmpDir, mf, "package magic\nconst (\n)\n")
	}

	// Replace magic_sm_im.go with correct content.
	writeMagicFile(t, tmpDir, "magic_sm_im.go", correctMagicSource(cryptoutilSharedMagic.OTLPServiceSMIM))

	err := lintFitnessMagicE2EContainerNames.CheckInDir(newTestLogger(), tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_WrongSQLiteContainer(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Write all magic files with no E2E constants.
	for _, mf := range []string{"magic_identity.go", "magic_jose.go", "magic_skeleton.go", "magic_sm.go", "magic_pki.go", "magic_sm_im.go"} {
		writeMagicFile(t, tmpDir, mf, "package magic\nconst (\n)\n")
	}

	// Write sm-im with wrong SQLite container name.
	wrongSrc := "package magic\nconst (\n" +
		"	IME2ESQLiteContainer = \"wrong-value\"\n" +
		"	IME2EPostgreSQL1Container = \"sm-im-app-postgres-1\"\n" +
		"	IME2EPostgreSQL2Container = \"sm-im-app-postgres-2\"\n" +
		")\n"
	writeMagicFile(t, tmpDir, "magic_sm_im.go", wrongSrc)

	err := lintFitnessMagicE2EContainerNames.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMIM)
	assert.Contains(t, err.Error(), "wrong-value")
}

func TestCheckInDir_MissingPostgreSQL1Container(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Write all magic files with no E2E constants.
	for _, mf := range []string{"magic_identity.go", "magic_jose.go", "magic_skeleton.go", "magic_sm.go", "magic_pki.go", "magic_sm_im.go"} {
		writeMagicFile(t, tmpDir, mf, "package magic\nconst (\n)\n")
	}

	// Write sm-im with SQLite present but PostgreSQL1 missing.
	partialSrc := "package magic\nconst (\n" +
		"	IME2ESQLiteContainer = \"sm-im-app-sqlite-1\"\n" +
		"	IME2EPostgreSQL2Container = \"sm-im-app-postgres-2\"\n" +
		")\n"
	writeMagicFile(t, tmpDir, "magic_sm_im.go", partialSrc)

	err := lintFitnessMagicE2EContainerNames.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "E2EPostgreSQL1Container")
}

func TestCheckInDir_NoE2EConstants_Skipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Write all magic files with no E2E constants — all should be skipped.
	for _, mf := range []string{"magic_identity.go", "magic_jose.go", "magic_skeleton.go", "magic_sm.go", "magic_pki.go", "magic_sm_im.go"} {
		writeMagicFile(t, tmpDir, mf, "package magic\nconst (\n)\n")
	}

	err := lintFitnessMagicE2EContainerNames.CheckInDir(newTestLogger(), tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_MissingMagicFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create the magic dir but leave it empty — all magic file reads will fail.
	magicDir := filepath.Join(tmpDir, "internal", "shared", "magic")
	require.NoError(t, os.MkdirAll(magicDir, cryptoutilSharedMagic.CICDTempDirPermissions))

	err := lintFitnessMagicE2EContainerNames.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
}
