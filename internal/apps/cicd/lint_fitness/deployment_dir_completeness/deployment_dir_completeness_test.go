// Copyright (c) 2025 Justin Cranford

package deployment_dir_completeness_test

import (
"os"
"path/filepath"
"testing"

cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
lintFitnessRegistry "cryptoutil/internal/apps/cicd/lint_fitness/registry"
cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

"github.com/stretchr/testify/require"

lintFitnessDeploymentDirCompleteness "cryptoutil/internal/apps/cicd/lint_fitness/deployment_dir_completeness"
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

// setupAllDeploymentDirs creates a complete deployment directory structure for all 10 PS.
func setupAllDeploymentDirs(t *testing.T, tmpDir string) {
t.Helper()

for _, ps := range lintFitnessRegistry.AllProductServices() {
createDeploymentDir(t, tmpDir, ps.PSID)
}
}

// createDeploymentDir creates a complete deployment directory for a single PS.
func createDeploymentDir(t *testing.T, tmpDir, psID string) {
t.Helper()

deployDir := filepath.Join(tmpDir, "deployments", psID)
configDir := filepath.Join(deployDir, "config")
secretsDir := filepath.Join(deployDir, "secrets")

require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.DirPermissions))
require.NoError(t, os.MkdirAll(secretsDir, cryptoutilSharedMagic.DirPermissions))

require.NoError(t, os.WriteFile(filepath.Join(deployDir, "Dockerfile"), []byte("FROM scratch\n"), cryptoutilSharedMagic.CacheFilePermissions))
require.NoError(t, os.WriteFile(filepath.Join(deployDir, "compose.yml"), []byte("services: {}\n"), cryptoutilSharedMagic.CacheFilePermissions))

for _, suffix := range []string{"-app-common.yml", "-app-sqlite-1.yml", "-app-postgresql-1.yml", "-app-postgresql-2.yml"} {
fp := filepath.Join(configDir, psID+suffix)
require.NoError(t, os.WriteFile(fp, []byte("# config\n"), cryptoutilSharedMagic.CacheFilePermissions))
}
}

func TestCheck_RealWorkspace(t *testing.T) {
t.Parallel()

root := findProjectRoot(t)

err := lintFitnessDeploymentDirCompleteness.CheckInDir(newTestLogger(), root)
require.NoError(t, err)
}

func TestCheckInDir_AllPresent(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
setupAllDeploymentDirs(t, tmpDir)

err := lintFitnessDeploymentDirCompleteness.CheckInDir(newTestLogger(), tmpDir)
require.NoError(t, err)
}

func TestCheckInDir_MissingDockerfile(t *testing.T) {
t.Parallel()

tests := []struct {
name string
psID string
}{
{name: cryptoutilSharedMagic.OTLPServiceSMIM, psID: cryptoutilSharedMagic.OTLPServiceSMIM},
{name: cryptoutilSharedMagic.OTLPServiceSMKMS, psID: cryptoutilSharedMagic.OTLPServiceSMKMS},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
setupAllDeploymentDirs(t, tmpDir)

require.NoError(t, os.Remove(filepath.Join(tmpDir, "deployments", tc.psID, "Dockerfile")))

err := lintFitnessDeploymentDirCompleteness.CheckInDir(newTestLogger(), tmpDir)
require.Error(t, err)
require.Contains(t, err.Error(), "Dockerfile")
require.Contains(t, err.Error(), tc.psID)
})
}
}

func TestCheckInDir_MissingComposeYML(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
setupAllDeploymentDirs(t, tmpDir)

require.NoError(t, os.Remove(filepath.Join(tmpDir, "deployments", cryptoutilSharedMagic.OTLPServiceSMIM, "compose.yml")))

err := lintFitnessDeploymentDirCompleteness.CheckInDir(newTestLogger(), tmpDir)
require.Error(t, err)
require.Contains(t, err.Error(), "compose.yml")
}

func TestCheckInDir_MissingSecretsDir(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
setupAllDeploymentDirs(t, tmpDir)

require.NoError(t, os.RemoveAll(filepath.Join(tmpDir, "deployments", cryptoutilSharedMagic.OTLPServiceJoseJA, "secrets")))

err := lintFitnessDeploymentDirCompleteness.CheckInDir(newTestLogger(), tmpDir)
require.Error(t, err)
require.Contains(t, err.Error(), "secrets/")
}

func TestCheckInDir_MissingConfigDir(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
setupAllDeploymentDirs(t, tmpDir)

require.NoError(t, os.RemoveAll(filepath.Join(tmpDir, "deployments", cryptoutilSharedMagic.OTLPServicePKICA, "config")))

err := lintFitnessDeploymentDirCompleteness.CheckInDir(newTestLogger(), tmpDir)
require.Error(t, err)
require.Contains(t, err.Error(), "config/")
}

func TestCheckInDir_MissingConfigFile(t *testing.T) {
t.Parallel()

psID := cryptoutilSharedMagic.OTLPServiceSMIM
suffixes := []string{"-app-common.yml", "-app-sqlite-1.yml", "-app-postgresql-1.yml", "-app-postgresql-2.yml"}

for _, suffix := range suffixes {
cfgFile := psID + suffix
t.Run(cfgFile, func(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
setupAllDeploymentDirs(t, tmpDir)

require.NoError(t, os.Remove(filepath.Join(tmpDir, "deployments", psID, "config", cfgFile)))

err := lintFitnessDeploymentDirCompleteness.CheckInDir(newTestLogger(), tmpDir)
require.Error(t, err)
require.Contains(t, err.Error(), cfgFile)
})
}
}
