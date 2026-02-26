package lint_deployments

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckDelegationPattern_SuiteValid(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	compose := `include:
  - path: ../sm/compose.yml
  - path: ../pki/compose.yml
  - path: ../jose/compose.yml
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(compose), cryptoutilSharedMagic.CacheFilePermissions))

	result := &ValidationResult{Valid: true}
	checkDelegationPattern(dir, "cryptoutil-suite", DeploymentTypeSuite, result)

	assert.True(t, result.Valid, "expected valid for proper delegation")
	assert.Empty(t, result.Errors)
	assert.Empty(t, result.Warnings)
}

func TestCheckDelegationPattern_SuiteInvalidServiceLevel(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	compose := `include:
  - path: ../sm-kms/compose.yml
  - path: ../pki-ca/compose.yml
  - path: ../sm-im/compose.yml
  - path: ../jose-ja/compose.yml
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(compose), cryptoutilSharedMagic.CacheFilePermissions))

	result := &ValidationResult{Valid: true}
	checkDelegationPattern(dir, "cryptoutil-suite", DeploymentTypeSuite, result)

	assert.False(t, result.Valid, "expected invalid for service-level delegation")
	assert.Len(t, result.Errors, 4, "expected 4 errors for 4 invalid patterns")
}

func TestCheckDelegationPattern_SuiteMissingProducts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	compose := `include:
  - path: ../sm/compose.yml
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(compose), cryptoutilSharedMagic.CacheFilePermissions))

	result := &ValidationResult{Valid: true}
	checkDelegationPattern(dir, "cryptoutil-suite", DeploymentTypeSuite, result)

	assert.True(t, result.Valid, "should still be valid, missing products is a warning")
	assert.NotEmpty(t, result.Warnings, "expected warning about missing products")
}

func TestCheckDelegationPattern_ProductValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		deploymentName string
		composeContent string
	}{
		{
			name:           "sm includes sm-kms and sm-im",
			deploymentName: "sm",
			composeContent: "include:\n  - path: ../sm-kms/compose.yml\n  - path: ../sm-im/compose.yml\n",
		},
		{
			name:           "pki includes pki-ca",
			deploymentName: cryptoutilSharedMagic.PKIProductName,
			composeContent: "include:\n  - path: ../pki-ca/compose.yml\n",
		},
		{
			name:           "jose includes jose-ja",
			deploymentName: cryptoutilSharedMagic.JoseProductName,
			composeContent: "include:\n  - path: ../jose-ja/compose.yml\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(tc.composeContent), cryptoutilSharedMagic.CacheFilePermissions))

			result := &ValidationResult{Valid: true}
			checkDelegationPattern(dir, tc.deploymentName, DeploymentTypeProduct, result)

			assert.True(t, result.Valid)
			assert.Empty(t, result.Errors)
		})
	}
}

func TestCheckDelegationPattern_ProductMissingService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		deploymentName string
		wantError      string
	}{
		{name: "sm missing sm-kms", deploymentName: "sm", wantError: cryptoutilSharedMagic.OTLPServiceSMKMS},
		{name: "sm missing sm-im", deploymentName: "sm", wantError: cryptoutilSharedMagic.OTLPServiceSMIM},
		{name: "pki missing pki-ca", deploymentName: cryptoutilSharedMagic.PKIProductName, wantError: cryptoutilSharedMagic.OTLPServicePKICA},
		{name: "jose missing jose-ja", deploymentName: cryptoutilSharedMagic.JoseProductName, wantError: cryptoutilSharedMagic.OTLPServiceJoseJA},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte("name: empty\n"), cryptoutilSharedMagic.CacheFilePermissions))

			result := &ValidationResult{Valid: true}
			checkDelegationPattern(dir, tc.deploymentName, DeploymentTypeProduct, result)

			assert.False(t, result.Valid)
			assert.NotEmpty(t, result.Errors)

			found := false

			for _, e := range result.Errors {
				if strings.Contains(e, tc.wantError) {
					found = true

					break
				}
			}

			assert.True(t, found, "expected error containing %q in %v", tc.wantError, result.Errors)
		})
	}
}

func TestCheckDelegationPattern_SkipsNonSuiteProduct(t *testing.T) {
	t.Parallel()

	result := &ValidationResult{Valid: true}
	checkDelegationPattern(t.TempDir(), cryptoutilSharedMagic.OTLPServiceJoseJA, DeploymentTypeProductService, result)

	assert.True(t, result.Valid, "should skip non-suite/product types")
	assert.Empty(t, result.Errors)
}

func TestCheckDelegationPattern_NoComposeFile(t *testing.T) {
	t.Parallel()

	result := &ValidationResult{Valid: true}
	checkDelegationPattern(t.TempDir(), "cryptoutil-suite", DeploymentTypeSuite, result)

	assert.True(t, result.Valid, "should skip when no compose file exists")
	assert.Empty(t, result.Errors)
}

func TestCheckOTLPProtocolOverride_NonServiceSkipped(t *testing.T) {
	t.Parallel()

	result := &ValidationResult{Valid: true}
	checkOTLPProtocolOverride(t.TempDir(), "sm", DeploymentTypeProduct, result)

	assert.True(t, result.Valid, "should skip non-product-service types")
	assert.Empty(t, result.Warnings)
}

func TestCheckOTLPProtocolOverride_WithProtocolPrefix(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configDir := filepath.Join(dir, "config")
	require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
	require.NoError(t, os.WriteFile(
		filepath.Join(configDir, "config-test.yml"),
		[]byte("otlp-endpoint: grpc://collector:4317\n"),
		cryptoutilSharedMagic.CacheFilePermissions,
	))

	result := &ValidationResult{Valid: true}
	checkOTLPProtocolOverride(dir, cryptoutilSharedMagic.OTLPServiceJoseJA, DeploymentTypeProductService, result)

	assert.NotEmpty(t, result.Warnings, "expected warning about protocol prefix")
}

func TestCheckOTLPProtocolOverride_NoProtocolPrefix(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configDir := filepath.Join(dir, "config")
	require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
	require.NoError(t, os.WriteFile(
		filepath.Join(configDir, "config-test.yml"),
		[]byte("otlp-endpoint: collector:4317\n"),
		cryptoutilSharedMagic.CacheFilePermissions,
	))

	result := &ValidationResult{Valid: true}
	checkOTLPProtocolOverride(dir, cryptoutilSharedMagic.OTLPServiceJoseJA, DeploymentTypeProductService, result)

	assert.Empty(t, result.Warnings, "should not warn when no protocol prefix")
}

func TestCheckOTLPProtocolOverride_NoConfigDir(t *testing.T) {
	t.Parallel()

	result := &ValidationResult{Valid: true}
	checkOTLPProtocolOverride(t.TempDir(), cryptoutilSharedMagic.OTLPServiceJoseJA, DeploymentTypeProductService, result)

	assert.True(t, result.Valid)
	assert.Empty(t, result.Warnings)
}

func TestCheckBrowserServiceCredentials_AllPresent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	secretsDir := filepath.Join(dir, "secrets")
	require.NoError(t, os.MkdirAll(secretsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))

	credFiles := []string{
		"browser_username.secret", "browser_password.secret",
		"service_username.secret", "service_password.secret",
	}
	for _, f := range credFiles {
		require.NoError(t, os.WriteFile(filepath.Join(secretsDir, f), []byte("val"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	result := &ValidationResult{Valid: true}
	checkBrowserServiceCredentials(dir, cryptoutilSharedMagic.OTLPServiceJoseJA, DeploymentTypeProductService, result)

	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestCheckBrowserServiceCredentials_Missing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "secrets"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))

	result := &ValidationResult{Valid: true}
	checkBrowserServiceCredentials(dir, cryptoutilSharedMagic.OTLPServiceJoseJA, DeploymentTypeProductService, result)

	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 4, "expected 4 missing credential files")
}

func TestCheckBrowserServiceCredentials_SkipsNonService(t *testing.T) {
	t.Parallel()

	result := &ValidationResult{Valid: true}
	checkBrowserServiceCredentials(t.TempDir(), "sm", DeploymentTypeProduct, result)

	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}
