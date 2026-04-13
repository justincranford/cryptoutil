// Copyright (c) 2025 Justin Cranford

package template_drift

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// projectRoot returns the path to the project root for integration tests.
func projectRoot() string {
	return filepath.Join("..", "..", "..", "..", "..", "..")
}

// failingInstantiate always returns an error to test the instantiation error path.
func failingInstantiate(_ string, _ map[string]string) (string, error) {
	return "", fmt.Errorf("injected template error")
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheckDockerfile_PublicWrapper(t *testing.T) {
	origDir, err := os.Getwd()
	require.NoError(t, err)

	root := projectRoot()

	require.NoError(t, os.Chdir(root))

	t.Cleanup(func() {
		require.NoError(t, os.Chdir(origDir))
	})

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-dockerfile-public")
	require.NoError(t, CheckDockerfile(logger))
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheckCompose_PublicWrapper(t *testing.T) {
	origDir, err := os.Getwd()
	require.NoError(t, err)

	root := projectRoot()

	require.NoError(t, os.Chdir(root))

	t.Cleanup(func() {
		require.NoError(t, os.Chdir(origDir))
	})

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-compose-public")
	require.NoError(t, CheckCompose(logger))
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheckConfigCommon_PublicWrapper(t *testing.T) {
	origDir, err := os.Getwd()
	require.NoError(t, err)

	root := projectRoot()

	require.NoError(t, os.Chdir(root))

	t.Cleanup(func() {
		require.NoError(t, os.Chdir(origDir))
	})

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-config-common-public")
	require.NoError(t, CheckConfigCommon(logger))
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheckConfigSQLite_PublicWrapper(t *testing.T) {
	origDir, err := os.Getwd()
	require.NoError(t, err)

	root := projectRoot()

	require.NoError(t, os.Chdir(root))

	t.Cleanup(func() {
		require.NoError(t, os.Chdir(origDir))
	})

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-config-sqlite-public")
	require.NoError(t, CheckConfigSQLite(logger))
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheckConfigPostgreSQL_PublicWrapper(t *testing.T) {
	origDir, err := os.Getwd()
	require.NoError(t, err)

	root := projectRoot()

	require.NoError(t, os.Chdir(root))

	t.Cleanup(func() {
		require.NoError(t, os.Chdir(origDir))
	})

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-config-postgresql-public")
	require.NoError(t, CheckConfigPostgreSQL(logger))
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheckStandaloneConfig_PublicWrapper(t *testing.T) {
	origDir, err := os.Getwd()
	require.NoError(t, err)

	root := projectRoot()

	require.NoError(t, os.Chdir(root))

	t.Cleanup(func() {
		require.NoError(t, os.Chdir(origDir))
	})

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-standalone-config-public")
	require.NoError(t, CheckStandaloneConfig(logger))
}

// --- Missing file tests ---

func TestCheckDockerfile_MissingFile(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-dockerfile-missing")
	err := checkDockerfileInDir(logger, t.TempDir(), instantiate)
	require.Error(t, err)
	require.Contains(t, err.Error(), "template-dockerfile violations")
}

func TestCheckCompose_MissingFile(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-compose-missing")
	err := checkComposeInDir(logger, t.TempDir(), instantiate)
	require.Error(t, err)
	require.Contains(t, err.Error(), "template-compose violations")
}

func TestCheckConfigCommon_MissingFile(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-config-common-missing")
	err := checkConfigCommonInDir(logger, t.TempDir(), instantiate)
	require.Error(t, err)
	require.Contains(t, err.Error(), "template-config-common violations")
}

func TestCheckConfigSQLite_MissingFile(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-config-sqlite-missing")
	err := checkConfigSQLiteInDir(logger, t.TempDir(), instantiate)
	require.Error(t, err)
	require.Contains(t, err.Error(), "template-config-sqlite violations")
}

func TestCheckConfigPostgreSQL_MissingFile(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-config-postgresql-missing")
	err := checkConfigPostgreSQLInDir(logger, t.TempDir(), instantiate)
	require.Error(t, err)
	require.Contains(t, err.Error(), "template-config-postgresql violations")
}

func TestCheckStandaloneConfig_MissingFile(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-standalone-config-missing")
	err := checkStandaloneConfigInDir(logger, t.TempDir(), instantiate)
	require.Error(t, err)
	require.Contains(t, err.Error(), "template-standalone-config violations")
}

// --- Template instantiation error tests ---

func TestCheckDockerfile_InstantiateError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-dockerfile-inst-err")
	err := checkDockerfileInDir(logger, t.TempDir(), failingInstantiate)
	require.Error(t, err)
	require.Contains(t, err.Error(), "injected template error")
}

func TestCheckCompose_InstantiateError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-compose-inst-err")
	err := checkComposeInDir(logger, t.TempDir(), failingInstantiate)
	require.Error(t, err)
	require.Contains(t, err.Error(), "injected template error")
}

func TestCheckConfigCommon_InstantiateError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-config-common-inst-err")
	err := checkConfigCommonInDir(logger, t.TempDir(), failingInstantiate)
	require.Error(t, err)
	require.Contains(t, err.Error(), "injected template error")
}

func TestCheckConfigSQLite_InstantiateError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-config-sqlite-inst-err")
	err := checkConfigSQLiteInDir(logger, t.TempDir(), failingInstantiate)
	require.Error(t, err)
	require.Contains(t, err.Error(), "injected template error")
}

func TestCheckConfigPostgreSQL_InstantiateError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-config-postgresql-inst-err")
	err := checkConfigPostgreSQLInDir(logger, t.TempDir(), failingInstantiate)
	require.Error(t, err)
	require.Contains(t, err.Error(), "injected template error")
}

func TestCheckStandaloneConfig_InstantiateError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-standalone-config-inst-err")
	err := checkStandaloneConfigInDir(logger, t.TempDir(), failingInstantiate)
	require.Error(t, err)
	require.Contains(t, err.Error(), "injected template error")
}

// --- Drift detection tests ---

func TestCheckDockerfile_DriftDetection(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	for _, ps := range cryptoutilRegistry.AllProductServices() {
		dockerfilePath := filepath.Join(tmpDir, "deployments", ps.PSID, "Dockerfile")
		require.NoError(t, os.MkdirAll(filepath.Dir(dockerfilePath), cryptoutilSharedMagic.CICDTempDirPermissions))
		require.NoError(t, os.WriteFile(dockerfilePath, []byte("FROM scratch\n"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-dockerfile-drift")
	err := checkDockerfileInDir(logger, tmpDir, instantiate)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Dockerfile drift")
}

func TestCheckCompose_DriftDetection(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	for _, ps := range cryptoutilRegistry.AllProductServices() {
		composePath := filepath.Join(tmpDir, "deployments", ps.PSID, "compose.yml")
		require.NoError(t, os.MkdirAll(filepath.Dir(composePath), cryptoutilSharedMagic.CICDTempDirPermissions))
		require.NoError(t, os.WriteFile(composePath, []byte("version: '3'\n"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-compose-drift")
	err := checkComposeInDir(logger, tmpDir, instantiate)
	require.Error(t, err)
	require.Contains(t, err.Error(), "compose.yml drift")
}

func TestCheckConfigCommon_DriftDetection(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	for _, ps := range cryptoutilRegistry.AllProductServices() {
		configPath := filepath.Join(tmpDir, "deployments", ps.PSID, "config", ps.PSID+"-app-common.yml")
		require.NoError(t, os.MkdirAll(filepath.Dir(configPath), cryptoutilSharedMagic.CICDTempDirPermissions))
		require.NoError(t, os.WriteFile(configPath, []byte("wrong: content\n"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-config-common-drift")
	err := checkConfigCommonInDir(logger, tmpDir, instantiate)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config-common drift")
}

func TestCheckConfigSQLite_DriftDetection(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	for _, ps := range cryptoutilRegistry.AllProductServices() {
		suffixes := []string{
			cryptoutilRegistry.DeploymentConfigSuffixSQLite1,
			cryptoutilRegistry.DeploymentConfigSuffixSQLite2,
		}
		for _, suffix := range suffixes {
			configPath := filepath.Join(tmpDir, "deployments", ps.PSID, "config", ps.PSID+suffix)
			require.NoError(t, os.MkdirAll(filepath.Dir(configPath), cryptoutilSharedMagic.CICDTempDirPermissions))
			require.NoError(t, os.WriteFile(configPath, []byte("wrong: content\n"), cryptoutilSharedMagic.CacheFilePermissions))
		}
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-config-sqlite-drift")
	err := checkConfigSQLiteInDir(logger, tmpDir, instantiate)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config-sqlite drift")
}

func TestCheckConfigPostgreSQL_DriftDetection(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	for _, ps := range cryptoutilRegistry.AllProductServices() {
		suffixes := []string{
			cryptoutilRegistry.DeploymentConfigSuffixPostgresql1,
			cryptoutilRegistry.DeploymentConfigSuffixPostgresql2,
		}
		for _, suffix := range suffixes {
			configPath := filepath.Join(tmpDir, "deployments", ps.PSID, "config", ps.PSID+suffix)
			require.NoError(t, os.MkdirAll(filepath.Dir(configPath), cryptoutilSharedMagic.CICDTempDirPermissions))
			require.NoError(t, os.WriteFile(configPath, []byte("wrong: content\n"), cryptoutilSharedMagic.CacheFilePermissions))
		}
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-config-postgresql-drift")
	err := checkConfigPostgreSQLInDir(logger, tmpDir, instantiate)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config-postgresql drift")
}

func TestCheckStandaloneConfig_DriftDetection(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	for _, ps := range cryptoutilRegistry.AllProductServices() {
		configPath := filepath.Join(tmpDir, cryptoutilSharedMagic.CICDConfigsDir, ps.PSID, ps.PSID+"-framework.yml")
		require.NoError(t, os.MkdirAll(filepath.Dir(configPath), cryptoutilSharedMagic.CICDTempDirPermissions))
		require.NoError(t, os.WriteFile(configPath, []byte("wrong: content\n"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-standalone-config-drift")
	err := checkStandaloneConfigInDir(logger, tmpDir, instantiate)
	require.Error(t, err)
	require.Contains(t, err.Error(), "standalone config drift")
}
