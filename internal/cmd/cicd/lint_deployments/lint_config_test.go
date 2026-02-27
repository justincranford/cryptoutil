package lint_deployments_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	. "cryptoutil/internal/cmd/cicd/lint_deployments"
)

func TestValidateConfigFiles_MissingRequiredConfigs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "secrets"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "config"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "compose.yml"), []byte("version: '3'"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), []byte("FROM alpine"), cryptoutilSharedMagic.CacheFilePermissions))

	createRequiredSecrets(t, tmpDir)

	// Config dir exists but has no config files.
	result, err := ValidateDeploymentStructure(tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS, "PRODUCT-SERVICE")
	require.NoError(t, err)

	require.False(t, result.Valid, "should be invalid with missing config files")
	require.GreaterOrEqual(t, len(result.Errors), 4, "should have at least 4 errors for missing config files")

	// Verify specific missing files mentioned in errors.
	errorStr := joinErrors(result.Errors)
	require.Contains(t, errorStr, "sm-kms-app-common.yml")
	require.Contains(t, errorStr, "sm-kms-app-sqlite-1.yml")
	require.Contains(t, errorStr, "sm-kms-app-postgresql-1.yml")
	require.Contains(t, errorStr, "sm-kms-app-postgresql-2.yml")
}

func TestValidateConfigFiles_WrongPrefix(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "secrets"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "config"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "compose.yml"), []byte("version: '3'"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), []byte("FROM alpine"), cryptoutilSharedMagic.CacheFilePermissions))

	createRequiredSecrets(t, tmpDir)
	createRequiredConfigFiles(t, tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS)

	// Add a file with wrong prefix (kms-app.yml instead of sm-kms-app.yml).
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config", "kms-app.yml"), []byte("# wrong"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateDeploymentStructure(tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS, "PRODUCT-SERVICE")
	require.NoError(t, err)

	require.False(t, result.Valid, "should be invalid with wrong-prefix config file")

	errorStr := joinErrors(result.Errors)
	require.Contains(t, errorStr, "kms-app.yml")
	require.Contains(t, errorStr, "does not match required pattern")
}

func TestValidateConfigFiles_WrongSuffix(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "secrets"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "config"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "compose.yml"), []byte("version: '3'"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), []byte("FROM alpine"), cryptoutilSharedMagic.CacheFilePermissions))

	createRequiredSecrets(t, tmpDir)

	// Create config files with wrong suffix (no instance number).
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config", "sm-kms-app-common.yml"), []byte("# cfg"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config", "sm-kms-app-sqlite.yml"), []byte("# wrong"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config", "sm-kms-app-postgresql-1.yml"), []byte("# cfg"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config", "sm-kms-app-postgresql-2.yml"), []byte("# cfg"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateDeploymentStructure(tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS, "PRODUCT-SERVICE")
	require.NoError(t, err)

	// Should be invalid because sm-kms-app-sqlite-1.yml is missing (sm-kms-app-sqlite.yml is not the right name).
	require.False(t, result.Valid, "should be invalid with missing sqlite-1 config file")

	errorStr := joinErrors(result.Errors)
	require.Contains(t, errorStr, "sm-kms-app-sqlite-1.yml")
}

func TestValidateConfigFiles_DeprecatedDemoSeed(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createValidProductServiceDeployment(t, tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS)

	// Add deprecated demo-seed.yml.
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config", "demo-seed.yml"), []byte("# deprecated"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateDeploymentStructure(tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS, "PRODUCT-SERVICE")
	require.NoError(t, err)

	require.False(t, result.Valid, "should be invalid with deprecated demo-seed.yml")

	errorStr := joinErrors(result.Errors)
	require.Contains(t, errorStr, "demo-seed.yml")
	require.Contains(t, errorStr, "DEPRECATED")
	require.Contains(t, errorStr, "sm-kms-demo.yml")
}

func TestValidateConfigFiles_DeprecatedIntegration(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createValidProductServiceDeployment(t, tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS)

	// Add deprecated integration.yml.
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config", "integration.yml"), []byte("# deprecated"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateDeploymentStructure(tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS, "PRODUCT-SERVICE")
	require.NoError(t, err)

	require.False(t, result.Valid, "should be invalid with deprecated integration.yml")

	errorStr := joinErrors(result.Errors)
	require.Contains(t, errorStr, "integration.yml")
	require.Contains(t, errorStr, "DEPRECATED")
	require.Contains(t, errorStr, "sm-kms-e2e.yml")
}

func TestValidateConfigFiles_SinglePartDeploymentName(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "secrets"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "config"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "compose.yml"), []byte("version: '3'"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), []byte("FROM alpine"), cryptoutilSharedMagic.CacheFilePermissions))

	createRequiredSecrets(t, tmpDir)

	// Single-part name (not PRODUCT-SERVICE pattern) should produce error.
	result, err := ValidateDeploymentStructure(tmpDir, cryptoutilSharedMagic.KMSServiceName, "PRODUCT-SERVICE")
	require.NoError(t, err)

	require.False(t, result.Valid, "should be invalid with single-part deployment name")

	errorStr := joinErrors(result.Errors)
	require.Contains(t, errorStr, "does not match PRODUCT-SERVICE pattern")
}

func TestValidateConfigFiles_WrongProductPrefix(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createValidProductServiceDeployment(t, tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS)

	// Add a file with wrong product prefix (pki-kms instead of sm-kms).
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config", "pki-kms-app-common.yml"), []byte("# wrong product"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateDeploymentStructure(tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS, "PRODUCT-SERVICE")
	require.NoError(t, err)

	require.False(t, result.Valid, "should be invalid with wrong product prefix")

	errorStr := joinErrors(result.Errors)
	require.Contains(t, errorStr, "pki-kms-app-common.yml")
	require.Contains(t, errorStr, "does not match required pattern")
}

func TestValidateConfigFiles_NonYAMLFilesIgnored(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createValidProductServiceDeployment(t, tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS)

	// Add non-YAML files that should be ignored.
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config", "README.md"), []byte("# readme"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config", ".gitkeep"), []byte(""), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config", "notes.txt"), []byte("notes"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateDeploymentStructure(tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS, "PRODUCT-SERVICE")
	require.NoError(t, err)

	require.True(t, result.Valid, "non-YAML files should be ignored")
	require.Empty(t, result.Errors, "should have no errors for valid deployment with non-YAML files")
}

func TestValidateConfigFiles_IdentityMultiPartServiceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		productService string
	}{
		{name: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, productService: cryptoutilSharedMagic.OTLPServiceIdentityAuthz},
		{name: cryptoutilSharedMagic.OTLPServiceIdentityIDP, productService: cryptoutilSharedMagic.OTLPServiceIdentityIDP},
		{name: cryptoutilSharedMagic.OTLPServiceIdentityRP, productService: cryptoutilSharedMagic.OTLPServiceIdentityRP},
		{name: cryptoutilSharedMagic.OTLPServiceIdentityRS, productService: cryptoutilSharedMagic.OTLPServiceIdentityRS},
		{name: cryptoutilSharedMagic.OTLPServiceIdentitySPA, productService: cryptoutilSharedMagic.OTLPServiceIdentitySPA},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			createValidProductServiceDeployment(t, tmpDir, tc.productService)

			result, err := ValidateDeploymentStructure(tmpDir, tc.productService, "PRODUCT-SERVICE")
			require.NoError(t, err)

			require.True(t, result.Valid, "identity service %s should be valid", tc.productService)
			require.Empty(t, result.Errors, "should have no errors for valid %s deployment", tc.productService)
		})
	}
}
