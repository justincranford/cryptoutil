// Copyright (c) 2025 Justin Cranford

package standalone_config_otlp_names_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessStandaloneConfigOTLPNames "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/standalone_config_otlp_names"
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

// correctConfigContent generates a minimal config YAML with the correct otlp-service value.
func correctConfigContent(psID, suffix string) string {
	return fmt.Sprintf("otlp-service: %q\n", psID+suffix)
}

// setupAllCorrectOTLPConfigs creates all required config files with correct otlp-service values.
func setupAllCorrectOTLPConfigs(t *testing.T, tmpDir string) {
	t.Helper()

	configs := []struct {
		psID     string
		filename string
		suffix   string
	}{
		{cryptoutilSharedMagic.OTLPServiceSMIM, "sm-im-app-sqlite-1.yml", "-sqlite-1"},
		{cryptoutilSharedMagic.OTLPServiceSMIM, "sm-im-app-postgresql-1.yml", "-postgres-1"},
		{cryptoutilSharedMagic.OTLPServiceSMIM, "sm-im-app-postgresql-2.yml", "-postgres-2"},
		{cryptoutilSharedMagic.OTLPServiceSMKMS, "sm-kms-app-sqlite-1.yml", "-sqlite-1"},
		{cryptoutilSharedMagic.OTLPServiceSMKMS, "sm-kms-app-postgresql-1.yml", "-postgres-1"},
		{cryptoutilSharedMagic.OTLPServiceSMKMS, "sm-kms-app-postgresql-2.yml", "-postgres-2"},
	}

	for _, c := range configs {
		writeDeploymentConfigFile(t, tmpDir, c.psID, c.filename, correctConfigContent(c.psID, c.suffix))
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

	// Overwrite sm-kms sm-kms-app-sqlite-1.yml with a wrong otlp-service value.
	writeDeploymentConfigFile(t, tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS, "sm-kms-app-sqlite-1.yml",
		"otlp-service: \"sm-kms-wrong-name\"\n",
	)

	err := lintFitnessStandaloneConfigOTLPNames.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMKMS)
	assert.Contains(t, err.Error(), "sm-kms-wrong-name")
	assert.Contains(t, err.Error(), "sm-kms-sqlite-1")
}

func TestCheckInDir_MissingOTLPServiceKey(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllCorrectOTLPConfigs(t, tmpDir)

	// Overwrite sm-im sm-im-app-postgresql-1.yml with no otlp-service key.
	writeDeploymentConfigFile(t, tmpDir, cryptoutilSharedMagic.OTLPServiceSMIM, "sm-im-app-postgresql-1.yml",
		"bind-public-port: 8700\n",
	)

	err := lintFitnessStandaloneConfigOTLPNames.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMIM)
	assert.Contains(t, err.Error(), "missing required otlp-service key")
}

func TestCheckInDir_MissingConfigFile_Skipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllCorrectOTLPConfigs(t, tmpDir)

	// Remove sm-kms sm-kms-app-postgresql-2.yml — file absence is a presence violation, not OTLP names.
	require.NoError(t, os.Remove(filepath.Join(tmpDir, "deployments", cryptoutilSharedMagic.OTLPServiceSMKMS, "config", "sm-kms-app-postgresql-2.yml")))

	err := lintFitnessStandaloneConfigOTLPNames.CheckInDir(newTestLogger(), tmpDir)
	require.NoError(t, err)
}
