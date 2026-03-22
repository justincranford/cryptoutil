// Copyright (c) 2025 Justin Cranford

package magic_e2e_compose_path_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessMagicE2EComposePath "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/magic_e2e_compose_path"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// smIME2EMagicSrc is the magic Go source fragment for sm-im with a correct E2EComposeFile constant.
const smIME2EMagicSrc = "package magic\nconst (\n\tIME2EComposeFile = \"../../../../../deployments/sm-im/compose.yml\"\n)\n"

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

// writeMagicFile writes a magic Go source under internal/shared/magic/filename.
func writeMagicFile(t *testing.T, tmpDir, filename, content string) {
	t.Helper()

	magicDir := filepath.Join(tmpDir, "internal", "shared", "magic")
	require.NoError(t, os.MkdirAll(magicDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(magicDir, filename), []byte(content), cryptoutilSharedMagic.FilePermissions))
}

// createComposeFile creates an empty compose.yml at deployments/{psID}/compose.yml.
func createComposeFile(t *testing.T, tmpDir, psID string) {
	t.Helper()

	deployDir := filepath.Join(tmpDir, "deployments", psID)
	require.NoError(t, os.MkdirAll(deployDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(deployDir, "compose.yml"), []byte("services: {}\n"), cryptoutilSharedMagic.FilePermissions))
}

// createE2EDir creates an empty e2e directory for a PS.
func createE2EDir(t *testing.T, tmpDir, internalAppsDir string) string {
	t.Helper()

	e2eDir := filepath.Join(tmpDir, "internal", "apps", filepath.FromSlash(internalAppsDir), "e2e")
	require.NoError(t, os.MkdirAll(e2eDir, cryptoutilSharedMagic.CICDTempDirPermissions))

	return e2eDir
}

func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)

	err := lintFitnessMagicE2EComposePath.CheckInDir(newTestLogger(), root)
	require.NoError(t, err)
}

func TestCheckInDir_AllCorrect(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Use sm-im as representative: magic_sm_im.go with E2EComposeFile pointing 5 levels up.
	// e2e dir: tmpDir/internal/apps/sm/im/e2e
	// compose:  tmpDir/deployments/sm-im/compose.yml
	// relative: ../../../../../deployments/sm-im/compose.yml
	createE2EDir(t, tmpDir, "sm/im/")
	createComposeFile(t, tmpDir, cryptoutilSharedMagic.OTLPServiceSMIM)

	writeMagicFile(t, tmpDir, "magic_sm_im.go", smIME2EMagicSrc)

	// All other magic files: no E2EComposeFile constant.
	for _, mf := range []string{"magic_identity.go", "magic_jose.go", "magic_skeleton.go", "magic_sm.go", "magic_pki.go"} {
		writeMagicFile(t, tmpDir, mf, "package magic\nconst (\n)\n")
	}

	err := lintFitnessMagicE2EComposePath.CheckInDir(newTestLogger(), tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_NonExistentComposePath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Write sm-im magic file pointing to a compose file that doesn't exist.
	createE2EDir(t, tmpDir, "sm/im/")
	// Do NOT create the compose file.

	writeMagicFile(t, tmpDir, "magic_sm_im.go", smIME2EMagicSrc)

	for _, mf := range []string{"magic_identity.go", "magic_jose.go", "magic_skeleton.go", "magic_sm.go", "magic_pki.go"} {
		writeMagicFile(t, tmpDir, mf, "package magic\nconst (\n)\n")
	}

	err := lintFitnessMagicE2EComposePath.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMIM)
	assert.Contains(t, err.Error(), "non-existent")
}

func TestCheckInDir_NoE2EComposeFile_Skipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// All magic files have no E2EComposeFile constant — all should be skipped.
	for _, mf := range []string{"magic_identity.go", "magic_jose.go", "magic_skeleton.go", "magic_sm.go", "magic_pki.go", "magic_sm_im.go"} {
		writeMagicFile(t, tmpDir, mf, "package magic\nconst (\n)\n")
	}

	err := lintFitnessMagicE2EComposePath.CheckInDir(newTestLogger(), tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_MissingMagicFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create magic dir but leave it empty.
	magicDir := filepath.Join(tmpDir, "internal", "shared", "magic")
	require.NoError(t, os.MkdirAll(magicDir, cryptoutilSharedMagic.CICDTempDirPermissions))

	err := lintFitnessMagicE2EComposePath.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
}
