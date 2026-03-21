// Copyright (c) 2025 Justin Cranford

package standalone_config_otlp_names_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintFitnessStandaloneConfigOTLPNames "cryptoutil/internal/apps/cicd/lint_fitness/standalone_config_otlp_names"
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

	configDir := filepath.Join(tmpDir, "configs", product, service)
	require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, filename), []byte(content), cryptoutilSharedMagic.CICDOutputFilePermissions))
}

// correctConfigContent generates a minimal config YAML with the correct otlp-service value.
func correctConfigContent(psID, suffix string) string {
	return fmt.Sprintf("otlp-service: %q\n", psID+suffix)
}

// setupAllCorrectOTLPConfigs creates all required config files with correct otlp-service values.
func setupAllCorrectOTLPConfigs(t *testing.T, tmpDir string) {
	t.Helper()

	configs := []struct {
		product  string
		service  string
		psID     string
		filename string
		suffix   string
	}{
		{cryptoutilSharedMagic.SMProductName, cryptoutilSharedMagic.IMServiceName, cryptoutilSharedMagic.OTLPServiceSMIM, "config-sqlite.yml", "-sqlite-1"},
		{cryptoutilSharedMagic.SMProductName, cryptoutilSharedMagic.IMServiceName, cryptoutilSharedMagic.OTLPServiceSMIM, "config-pg-1.yml", "-postgres-1"},
		{cryptoutilSharedMagic.SMProductName, cryptoutilSharedMagic.IMServiceName, cryptoutilSharedMagic.OTLPServiceSMIM, "config-pg-2.yml", "-postgres-2"},
		{cryptoutilSharedMagic.SMProductName, cryptoutilSharedMagic.KMSServiceName, cryptoutilSharedMagic.OTLPServiceSMKMS, "config-sqlite.yml", "-sqlite-1"},
		{cryptoutilSharedMagic.SMProductName, cryptoutilSharedMagic.KMSServiceName, cryptoutilSharedMagic.OTLPServiceSMKMS, "config-pg-1.yml", "-postgres-1"},
		{cryptoutilSharedMagic.SMProductName, cryptoutilSharedMagic.KMSServiceName, cryptoutilSharedMagic.OTLPServiceSMKMS, "config-pg-2.yml", "-postgres-2"},
	}

	for _, c := range configs {
		writeConfigFile(t, tmpDir, c.product, c.service, c.filename, correctConfigContent(c.psID, c.suffix))
	}
}

func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)

	err := lintFitnessStandaloneConfigOTLPNames.CheckInDir(newTestLogger(), root)
	require.NoError(t, err)
}

func TestCheckInDir_AllCorrect(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllCorrectOTLPConfigs(t, tmpDir)

	err := lintFitnessStandaloneConfigOTLPNames.CheckInDir(newTestLogger(), tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_WrongOTLPServiceValue(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllCorrectOTLPConfigs(t, tmpDir)

	// Overwrite sm-kms config-sqlite.yml with a wrong otlp-service value.
	writeConfigFile(t, tmpDir, cryptoutilSharedMagic.SMProductName, cryptoutilSharedMagic.KMSServiceName, "config-sqlite.yml",
		"otlp-service: \"sm-kms-wrong-name\"\n",
	)

	err := lintFitnessStandaloneConfigOTLPNames.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.KMSServiceName)
	assert.Contains(t, err.Error(), "sm-kms-wrong-name")
	assert.Contains(t, err.Error(), "sm-kms-sqlite-1")
}

func TestCheckInDir_MissingOTLPServiceKey(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllCorrectOTLPConfigs(t, tmpDir)

	// Overwrite sm-im config-pg-1.yml with no otlp-service key.
	writeConfigFile(t, tmpDir, cryptoutilSharedMagic.SMProductName, cryptoutilSharedMagic.IMServiceName, "config-pg-1.yml",
		"bind-public-port: 8700\n",
	)

	err := lintFitnessStandaloneConfigOTLPNames.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.IMServiceName)
	assert.Contains(t, err.Error(), "missing required otlp-service key")
}

func TestCheckInDir_MissingConfigFile_Skipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllCorrectOTLPConfigs(t, tmpDir)

	// Remove sm-kms config-pg-2.yml — file absence is a presence violation, not OTLP names.
	require.NoError(t, os.Remove(filepath.Join(tmpDir, "configs", cryptoutilSharedMagic.SMProductName, cryptoutilSharedMagic.KMSServiceName, "config-pg-2.yml")))

	err := lintFitnessStandaloneConfigOTLPNames.CheckInDir(newTestLogger(), tmpDir)
	require.NoError(t, err)
}
