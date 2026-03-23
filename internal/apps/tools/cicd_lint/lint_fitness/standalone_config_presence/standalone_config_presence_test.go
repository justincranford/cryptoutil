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

// writeConfigFile creates a config YAML file under configs/{product}/{service}/.
func writeConfigFile(t *testing.T, tmpDir, product, service, filename, content string) {
	t.Helper()

	configDir := filepath.Join(tmpDir, cryptoutilSharedMagic.CICDConfigsDir, product, service)
	require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, filename), []byte(content), cryptoutilSharedMagic.CICDOutputFilePermissions))
}

// setupAllRequiredConfigs creates all three required config files for both allowlist PS.
func setupAllRequiredConfigs(t *testing.T, tmpDir string) {
	t.Helper()

	for _, product := range []string{cryptoutilSharedMagic.SMProductName} {
		for _, svc := range []struct {
			service string
			psID    string
		}{
			{cryptoutilSharedMagic.IMServiceName, cryptoutilSharedMagic.OTLPServiceSMIM},
			{cryptoutilSharedMagic.KMSServiceName, cryptoutilSharedMagic.OTLPServiceSMKMS},
		} {
			for _, suffix := range []string{"-sqlite.yml", "-pg-1.yml", "-pg-2.yml"} {
				writeConfigFile(t, tmpDir, product, svc.service, svc.psID+suffix, "# placeholder\n")
			}
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

	require.NoError(t, os.Remove(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDConfigsDir, cryptoutilSharedMagic.SMProductName, cryptoutilSharedMagic.KMSServiceName, "sm-kms-sqlite.yml")))

	err := lintFitnessStandaloneConfigPresence.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMKMS)
	assert.Contains(t, err.Error(), "sm-kms-sqlite.yml")
}

func TestCheckInDir_MissingPG1Config(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllRequiredConfigs(t, tmpDir)

	require.NoError(t, os.Remove(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDConfigsDir, cryptoutilSharedMagic.SMProductName, cryptoutilSharedMagic.IMServiceName, "sm-im-pg-1.yml")))

	err := lintFitnessStandaloneConfigPresence.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMIM)
	assert.Contains(t, err.Error(), "sm-im-pg-1.yml")
}

func TestCheckInDir_MissingConfigDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllRequiredConfigs(t, tmpDir)

	// Remove the entire sm-kms config directory.
	require.NoError(t, os.RemoveAll(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDConfigsDir, cryptoutilSharedMagic.SMProductName, cryptoutilSharedMagic.KMSServiceName)))

	err := lintFitnessStandaloneConfigPresence.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMKMS)
}
