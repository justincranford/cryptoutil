package lint_deployments_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "cryptoutil/internal/cmd/cicd/lint_deployments"
)

// createRequiredSecrets creates all 14 required secret files in the secrets directory.
func createRequiredSecrets(t *testing.T, baseDir string) {
	t.Helper()

	secretsPath := filepath.Join(baseDir, "secrets")
	requiredSecrets := []string{
		"unseal_1of5.secret", "unseal_2of5.secret", "unseal_3of5.secret",
		"unseal_4of5.secret", "unseal_5of5.secret", "hash_pepper_v3.secret",
		"postgres_url.secret", "postgres_username.secret",
		"postgres_password.secret", "postgres_database.secret",
		"browser_username.secret", "browser_password.secret",
		"service_username.secret", "service_password.secret",
	}

	for _, secret := range requiredSecrets {
		require.NoError(t, os.WriteFile(filepath.Join(secretsPath, secret), []byte("secret-value"), 0o600))
	}
}

// createRequiredConfigFiles creates the 4 required config files for a PRODUCT-SERVICE.
func createRequiredConfigFiles(t *testing.T, baseDir string, productService string) {
	t.Helper()

	configPath := filepath.Join(baseDir, "config")
	requiredConfigs := []string{
		productService + "-app-common.yml",
		productService + "-app-sqlite-1.yml",
		productService + "-app-postgresql-1.yml",
		productService + "-app-postgresql-2.yml",
	}

	for _, cfg := range requiredConfigs {
		require.NoError(t, os.WriteFile(filepath.Join(configPath, cfg), []byte("# config"), 0o600))
	}
}

// createValidProductServiceDeployment creates a fully valid PRODUCT-SERVICE deployment directory.
func createValidProductServiceDeployment(t *testing.T, baseDir string, productService string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "secrets"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "config"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(baseDir, "compose.yml"), []byte("version: '3'"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(baseDir, "Dockerfile"), []byte("FROM alpine"), 0o600))

	createRequiredSecrets(t, baseDir)
	createRequiredConfigFiles(t, baseDir, productService)
}

func TestValidateDeploymentStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		deploymentName     string
		structType         string
		setupFunc          func(t *testing.T, baseDir string)
		wantValid          bool
		wantMissingDirs    []string
		wantMissingFiles   []string
		wantMissingSecrets []string
		wantErrors         []string
	}{
		{
			name:           "valid PRODUCT-SERVICE deployment",
			deploymentName: "sm-kms",
			structType:     "PRODUCT-SERVICE",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()
				createValidProductServiceDeployment(t, baseDir, "sm-kms")
			},
			wantValid: true,
		},
		{
			name:           "valid PRODUCT-SERVICE deployment with optional demo and e2e",
			deploymentName: "sm-kms",
			structType:     "PRODUCT-SERVICE",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()
				createValidProductServiceDeployment(t, baseDir, "sm-kms")
				// Add optional demo and e2e files (should not cause errors).
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "config", "sm-kms-demo.yml"), []byte("# demo"), 0o600))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "config", "sm-kms-e2e.yml"), []byte("# e2e"), 0o600))
			},
			wantValid: true,
		},
		{
			name:           "valid PRODUCT-SERVICE deployment with identity-authz multi-part name",
			deploymentName: "identity-authz",
			structType:     "PRODUCT-SERVICE",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()
				createValidProductServiceDeployment(t, baseDir, "identity-authz")
			},
			wantValid: true,
		},
		{
			name:           "missing secrets directory",
			deploymentName: "sm-kms",
			structType:     "PRODUCT-SERVICE",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "config"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "compose.yml"), []byte("version: '3'"), 0o600))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "Dockerfile"), []byte("FROM alpine"), 0o600))
			},
			wantValid:       false,
			wantMissingDirs: []string{"secrets"},
		},
		{
			name:           "missing config directory",
			deploymentName: "sm-kms",
			structType:     "PRODUCT-SERVICE",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "secrets"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "compose.yml"), []byte("version: '3'"), 0o600))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "Dockerfile"), []byte("FROM alpine"), 0o600))
				createRequiredSecrets(t, baseDir)
			},
			wantValid:       false,
			wantMissingDirs: []string{"config"},
		},
		{
			name:           "missing compose.yml",
			deploymentName: "sm-kms",
			structType:     "PRODUCT-SERVICE",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "secrets"), 0o755))
				require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "config"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "Dockerfile"), []byte("FROM alpine"), 0o600))
				createRequiredSecrets(t, baseDir)
				createRequiredConfigFiles(t, baseDir, "sm-kms")
			},
			wantValid:        false,
			wantMissingFiles: []string{"compose.yml"},
		},
		{
			name:           "missing Dockerfile",
			deploymentName: "sm-kms",
			structType:     "PRODUCT-SERVICE",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "secrets"), 0o755))
				require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "config"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "compose.yml"), []byte("version: '3'"), 0o600))
				createRequiredSecrets(t, baseDir)
				createRequiredConfigFiles(t, baseDir, "sm-kms")
			},
			wantValid:        false,
			wantMissingFiles: []string{"Dockerfile"},
		},
		{
			name:           "missing some secrets",
			deploymentName: "sm-kms",
			structType:     "PRODUCT-SERVICE",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "secrets"), 0o755))
				require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "config"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "compose.yml"), []byte("version: '3'"), 0o600))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "Dockerfile"), []byte("FROM alpine"), 0o600))
				// Only create 2 of 10 required secrets.
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "secrets", "unseal_1of5.secret"), []byte("s"), 0o600))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "secrets", "hash_pepper_v3.secret"), []byte("s"), 0o600))
				createRequiredConfigFiles(t, baseDir, "sm-kms")
			},
			wantValid: false,
			wantMissingSecrets: []string{
				"unseal_2of5.secret", "unseal_3of5.secret",
				"unseal_4of5.secret", "unseal_5of5.secret",
				"postgres_url.secret", "postgres_username.secret",
				"postgres_password.secret", "postgres_database.secret",
			},
		},
		{
			name:           "valid template deployment",
			deploymentName: "template",
			structType:     "template",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "secrets"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "compose.yml"), []byte("version: '3'"), 0o600))
				createRequiredSecrets(t, baseDir)
			},
			wantValid: true,
		},
		{
			name:           "valid infrastructure deployment",
			deploymentName: "grafana-otel-lgtm",
			structType:     "infrastructure",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "compose.yml"), []byte("version: '3'"), 0o600))
			},
			wantValid: true,
		},
		{
			name:           "infrastructure with optional files",
			deploymentName: "grafana-otel-lgtm",
			structType:     "infrastructure",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "compose.yml"), []byte("version: '3'"), 0o600))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "init-db.sql"), []byte("CREATE TABLE test;"), 0o600))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "README.md"), []byte("# README"), 0o600))
			},
			wantValid: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			tc.setupFunc(t, tmpDir)

			result, err := ValidateDeploymentStructure(tmpDir, tc.deploymentName, tc.structType)
			require.NoError(t, err)

			assert.Equal(t, tc.wantValid, result.Valid, "validity mismatch")
			assert.ElementsMatch(t, tc.wantMissingDirs, result.MissingDirs, "missing dirs mismatch")
			assert.ElementsMatch(t, tc.wantMissingFiles, result.MissingFiles, "missing files mismatch")
			assert.ElementsMatch(t, tc.wantMissingSecrets, result.MissingSecrets, "missing secrets mismatch")
		})
	}
}

func TestValidateDeploymentStructure_UnknownType(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	result, err := ValidateDeploymentStructure(tmpDir, "test", "UNKNOWN-TYPE")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unknown structure type")
}

// TestValidateAllDeployments_ProductAndSuiteAndTemplate tests PRODUCT, SUITE, template paths.
func TestValidateAllDeployments_ProductAndSuiteAndTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(t *testing.T) string
		wantCount int
	}{
		{
			name: "product directory triggers PRODUCT validation",
			setup: func(t *testing.T) string {
				t.Helper()

				tmpDir := t.TempDir()
				productDir := filepath.Join(tmpDir, "identity")
				require.NoError(t, os.MkdirAll(filepath.Join(productDir, "secrets"), 0o750))
				require.NoError(t, os.WriteFile(
					filepath.Join(productDir, "compose.yml"),
					[]byte("name: identity\n"), 0o600))

				return tmpDir
			},
			wantCount: 1,
		},
		{
			name: "suite directory triggers SUITE validation",
			setup: func(t *testing.T) string {
				t.Helper()

				tmpDir := t.TempDir()
				suiteDir := filepath.Join(tmpDir, "cryptoutil")
				require.NoError(t, os.MkdirAll(filepath.Join(suiteDir, "secrets"), 0o750))
				require.NoError(t, os.WriteFile(
					filepath.Join(suiteDir, "compose.yml"),
					[]byte("name: cryptoutil\n"), 0o600))

				return tmpDir
			},
			wantCount: 1,
		},
		{
			name: "template directory triggers template validation",
			setup: func(t *testing.T) string {
				t.Helper()

				tmpDir := t.TempDir()
				templateDir := filepath.Join(tmpDir, "template")
				require.NoError(t, os.MkdirAll(filepath.Join(templateDir, "secrets"), 0o750))
				require.NoError(t, os.WriteFile(
					filepath.Join(templateDir, "compose.yml"),
					[]byte("name: template\n"), 0o600))

				return tmpDir
			},
			wantCount: 1,
		},
		{
			name: "empty root returns no results",
			setup: func(t *testing.T) string {
				t.Helper()

				return t.TempDir()
			},
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := tc.setup(t)
			results, err := ValidateAllDeployments(root)
			require.NoError(t, err)
			assert.Len(t, results, tc.wantCount)
		})
	}
}

// TestValidateDeploymentStructure_ProductType tests PRODUCT type triggers product secrets.
func TestValidateDeploymentStructure_ProductType(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "secrets"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "compose.yml"), []byte("n: t\n"), 0o600))

	result, err := ValidateDeploymentStructure(tmpDir, "identity", "PRODUCT")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Missing product secrets should cause invalid.
	assert.False(t, result.Valid)
}

// TestValidateDeploymentStructure_SuiteType tests SUITE type triggers suite secrets.
func TestValidateDeploymentStructure_SuiteType(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "secrets"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "compose.yml"), []byte("n: t\n"), 0o600))

	result, err := ValidateDeploymentStructure(tmpDir, "cryptoutil", "SUITE")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Missing suite secrets should cause invalid.
	assert.False(t, result.Valid)
}
