// Copyright (c) 2025 Justin Cranford

package standalone_config_presence_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessStandaloneConfigPresence "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/standalone_config_presence"
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

// writeDeploymentConfigFile creates a config YAML file under deployments/{psID}/config/.
func writeDeploymentConfigFile(t *testing.T, tmpDir, psID, filename, content string) {
	t.Helper()

	configDir := filepath.Join(tmpDir, "deployments", psID, "config")
	require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, filename), []byte(content), cryptoutilSharedMagic.CICDOutputFilePermissions))
}

// setupAllRequiredConfigs creates all required config files for both allowlist PS.
func setupAllRequiredConfigs(t *testing.T, tmpDir string) {
	t.Helper()

	for _, psID := range []string{cryptoutilSharedMagic.OTLPServiceSMIM, cryptoutilSharedMagic.OTLPServiceSMKMS} {
		for _, suffix := range []string{
			"-app-framework-common.yml",
			"-app-framework-sqlite-1.yml",
			"-app-framework-sqlite-2.yml",
			"-app-framework-postgresql-1.yml",
			"-app-framework-postgresql-2.yml",
			"-app-domain-common.yml",
			"-app-domain-sqlite-1.yml",
			"-app-domain-sqlite-2.yml",
			"-app-domain-postgresql-1.yml",
			"-app-domain-postgresql-2.yml",
		} {
			writeDeploymentConfigFile(t, tmpDir, psID, psID+suffix, "# placeholder\n")
		}
	}
}

func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)

	err := lintFitnessStandaloneConfigPresence.CheckInDir(newTestLogger(), root)
	require.NoError(t, err)
}

func TestCheckInDir_AllPresent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllRequiredConfigs(t, tmpDir)

	err := lintFitnessStandaloneConfigPresence.CheckInDir(newTestLogger(), tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_MissingSQLiteConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllRequiredConfigs(t, tmpDir)

	require.NoError(t, os.Remove(filepath.Join(tmpDir, "deployments", cryptoutilSharedMagic.OTLPServiceSMKMS, "config", "sm-kms-app-framework-sqlite-1.yml")))

	err := lintFitnessStandaloneConfigPresence.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMKMS)
	assert.Contains(t, err.Error(), "sm-kms-app-framework-sqlite-1.yml")
}

func TestCheckInDir_MissingPostgresConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllRequiredConfigs(t, tmpDir)

	require.NoError(t, os.Remove(filepath.Join(tmpDir, "deployments", cryptoutilSharedMagic.OTLPServiceSMIM, "config", "sm-im-app-framework-postgresql-1.yml")))

	err := lintFitnessStandaloneConfigPresence.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMIM)
	assert.Contains(t, err.Error(), "sm-im-app-framework-postgresql-1.yml")
}

func TestCheckInDir_MissingConfigDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllRequiredConfigs(t, tmpDir)

	// Remove the entire sm-kms deployment config directory.
	require.NoError(t, os.RemoveAll(filepath.Join(tmpDir, "deployments", cryptoutilSharedMagic.OTLPServiceSMKMS, "config")))

	err := lintFitnessStandaloneConfigPresence.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMKMS)
}
