package lint_deployments_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	. "cryptoutil/internal/cmd/cicd/lint_deployments"
)

// TestIntegrationFullPipeline tests the complete CICD validation pipeline:
// generate listings -> validate mirror -> validate compose -> validate config.
func TestIntegrationFullPipeline(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	deploymentsDir := filepath.Join(tmpDir, "deployments")
	configsDir := filepath.Join(tmpDir, "configs")

	// Create a realistic deployment structure.
	// Note: mapDeploymentToConfig maps PRODUCT-SERVICE "jose-ja" -> PRODUCT "jose".
	deployName := cryptoutilSharedMagic.OTLPServiceJoseJA
	configName := cryptoutilSharedMagic.JoseProductName
	svcDeployDir := filepath.Join(deploymentsDir, deployName)
	svcConfigDir := filepath.Join(configsDir, configName)

	require.NoError(t, os.MkdirAll(filepath.Join(svcDeployDir, "config"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(svcDeployDir, "secrets"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(svcConfigDir, "config"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	// Write compose file.
	composeContent := `services:
  test-svc-app:
    image: test-svc:latest
    ports:
      - "127.0.0.1:8080:8080"
    healthcheck:
      test: ["CMD", "wget", "--spider", "https://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    secrets:
      - test_secret.secret
  test-svc-postgres:
    image: postgres:18
    healthcheck:
      test: ["CMD", "pg_isready"]
      interval: 10s
      timeout: 5s
      retries: 3
secrets:
  test_secret.secret:
    file: ./secrets/test_secret.secret
`
	require.NoError(t, os.WriteFile(filepath.Join(svcDeployDir, "compose.yml"),
		[]byte(composeContent), cryptoutilSharedMagic.CacheFilePermissions))

	// Write config file.
	configContent := `bind-public-protocol: https
bind-public-address: 0.0.0.0
bind-public-port: 8080
bind-private-protocol: https
bind-private-address: 127.0.0.1
bind-private-port: 9090
database-url: "file:///run/secrets/db_url"
`
	require.NoError(t, os.WriteFile(
		filepath.Join(svcConfigDir, "config", "config.yml"),
		[]byte(configContent), cryptoutilSharedMagic.CacheFilePermissions))

	// Write deployment config file (empty placeholder).
	require.NoError(t, os.WriteFile(
		filepath.Join(svcDeployDir, "config", "config.yml"),
		[]byte(configContent), cryptoutilSharedMagic.CacheFilePermissions))

	// Write secrets.
	require.NoError(t, os.WriteFile(
		filepath.Join(svcDeployDir, "secrets", "test_secret.secret"),
		[]byte("secret-value"), cryptoutilSharedMagic.CacheFilePermissions))

	// Step 1: Generate listings.
	deploymentsOutput := filepath.Join(deploymentsDir, "deployments-all-files.json")
	configsOutput := filepath.Join(configsDir, "configs-all-files.json")

	err := WriteListingFile(deploymentsDir, deploymentsOutput)
	require.NoError(t, err, "generate deployments listing")

	err = WriteListingFile(configsDir, configsOutput)
	require.NoError(t, err, "generate configs listing")

	require.FileExists(t, deploymentsOutput)
	require.FileExists(t, configsOutput)

	// Step 2: Validate mirror.
	mirrorResult, err := ValidateStructuralMirror(deploymentsDir, configsDir)
	require.NoError(t, err, "validate mirror")
	require.True(t, mirrorResult.Valid, "mirror should be valid: %v", mirrorResult.Errors)

	// Step 3: Validate compose.
	composePath := filepath.Join(svcDeployDir, "compose.yml")
	composeResult, err := ValidateComposeFile(composePath)
	require.NoError(t, err, "validate compose")
	require.True(t, composeResult.Valid,
		"compose should be valid: errors=%v warnings=%v",
		composeResult.Errors, composeResult.Warnings)

	// Step 4: Validate config.
	configPath := filepath.Join(svcConfigDir, "config", "config.yml")
	configResult, err := ValidateConfigFile(configPath)
	require.NoError(t, err, "validate config")
	require.True(t, configResult.Valid,
		"config should be valid: errors=%v warnings=%v",
		configResult.Errors, configResult.Warnings)
}

// TestIntegrationFullPipeline_DetectsErrors validates the pipeline catches errors at each stage.
func TestIntegrationFullPipeline_DetectsErrors(t *testing.T) {
	t.Parallel()

	t.Run("compose errors detected", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		composePath := filepath.Join(tmpDir, "compose.yml")

		badCompose := `services:
  app:
    image: test:latest
    ports:
      - "8080:8080"
    environment:
      DB_PASSWORD: supersecret123
`
		require.NoError(t, os.WriteFile(composePath, []byte(badCompose), cryptoutilSharedMagic.CacheFilePermissions))

		result, err := ValidateComposeFile(composePath)
		require.NoError(t, err)
		require.False(t, result.Valid, "should detect hardcoded credentials")
	})

	t.Run("config errors detected", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yml")

		badConfig := `bind-public-protocol: http
bind-private-address: 0.0.0.0
database-url: "postgres://user:pass@db:5432/mydb"
`
		require.NoError(t, os.WriteFile(configPath, []byte(badConfig), cryptoutilSharedMagic.CacheFilePermissions))

		result, err := ValidateConfigFile(configPath)
		require.NoError(t, err)
		require.False(t, result.Valid, "should detect config violations")
		require.GreaterOrEqual(t, len(result.Errors), 3,
			"should have at least 3 errors (protocol, admin, db)")
	})

	t.Run("mirror errors detected", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		deploymentsDir := filepath.Join(tmpDir, "deployments")
		configsDir := filepath.Join(tmpDir, "configs")

		// Create deployment without matching config.
		require.NoError(t, os.MkdirAll(filepath.Join(deploymentsDir, "svc-a", "config"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(
			filepath.Join(deploymentsDir, "svc-a", "compose.yml"),
			[]byte("services: {}"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.MkdirAll(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

		result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
		require.NoError(t, err)
		// Missing config dir for deployment should generate warning or error.
		require.True(t, len(result.Warnings) > 0 || len(result.Errors) > 0,
			"should detect missing config mirror")
	})
}

// TestIntegrationRealFiles validates real project files when running from project root.
func TestIntegrationRealFiles(t *testing.T) {
	t.Parallel()

	// Only run when real directories exist.
	if _, err := os.Stat("../../../../../../deployments"); os.IsNotExist(err) {
		t.Skip("not running from project root context")
	}

	deploymentsDir := "../../../../../../deployments"
	configsDir := "../../../../../../configs"

	t.Run("real compose files validate", func(t *testing.T) {
		t.Parallel()

		composeFiles, err := filepath.Glob(filepath.Join(deploymentsDir, "*/compose.yml"))
		require.NoError(t, err)
		require.NotEmpty(t, composeFiles, "should find compose files")

		for _, f := range composeFiles {
			result, err := ValidateComposeFile(f)
			require.NoError(t, err, "ValidateComposeFile(%s)", f)

			// Log but don't fail on warnings.
			if len(result.Warnings) > 0 {
				t.Logf("warnings for %s: %v", f, result.Warnings)
			}
		}
	})

	t.Run("real config files validate", func(t *testing.T) {
		t.Parallel()

		configFiles, err := filepath.Glob(filepath.Join(configsDir, "*/*.yml"))
		require.NoError(t, err)

		configFiles2, err := filepath.Glob(filepath.Join(configsDir, "*/*/*.yml"))
		require.NoError(t, err)

		allConfigs := append(configFiles, configFiles2...)
		require.NotEmpty(t, allConfigs, "should find config files")

		for _, f := range allConfigs {
			result, err := ValidateConfigFile(f)
			require.NoError(t, err, "ValidateConfigFile(%s)", f)

			if len(result.Warnings) > 0 {
				t.Logf("warnings for %s: %v", f, result.Warnings)
			}
		}
	})
}
