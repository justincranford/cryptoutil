package lint_deployments

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const testNonexistentPath = "/nonexistent/path/abc123"

func TestRunNamingValidation_InvalidPaths(t *testing.T) {
	t.Parallel()

	result := &AllValidationResult{}
	runNamingValidation(testNonexistentPath, "/nonexistent/path/def456", result)

	require.Len(t, result.Results, 2)
	require.False(t, result.Results[0].Passed)
	require.Equal(t, validatorNameNaming, result.Results[0].Name)
	require.False(t, result.Results[1].Passed)
	require.Equal(t, validatorNameNaming, result.Results[1].Name)
}

func TestRunKebabCaseValidation_InvalidPath(t *testing.T) {
	t.Parallel()

	result := &AllValidationResult{}
	runKebabCaseValidation(testNonexistentPath, result)

	require.Len(t, result.Results, 1)
	require.False(t, result.Results[0].Passed)
	require.Equal(t, validatorNameKebabCase, result.Results[0].Name)
}

func TestRunSchemaValidation_ErrorPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create an unreadable YAML file matching service template config pattern.
	unreadable := filepath.Join(dir, "config-bad.yml")
	require.NoError(t, os.WriteFile(unreadable, []byte("key: val"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.Chmod(unreadable, 0o000))

	t.Cleanup(func() {
		_ = os.Chmod(unreadable, cryptoutilSharedMagic.CacheFilePermissions)
	})

	result := &AllValidationResult{}
	runSchemaValidation(dir, result)

	// Schema validator may error or report invalid depending on permissions.
	require.Greater(t, len(result.Results), 0)
}

func TestRunTemplatePatternValidation_NonexistentSkips(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	result := &AllValidationResult{}
	runTemplatePatternValidation(dir, result)

	// No template/ subdir means no results added.
	require.Empty(t, result.Results)
}

func TestRunTemplatePatternValidation_ErrorPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	templateDir := filepath.Join(dir, cryptoutilSharedMagic.SkeletonTemplateServiceName)
	require.NoError(t, os.MkdirAll(templateDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	result := &AllValidationResult{}
	runTemplatePatternValidation(dir, result)

	// Template dir exists but is empty - validator runs and reports.
	require.Greater(t, len(result.Results), 0)
}

func TestRunPortsValidation_ErrorPath(t *testing.T) {
	t.Parallel()

	deployments := []deploymentEntry{
		{path: testNonexistentPath, name: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, level: DeploymentTypeProductService},
	}

	result := &AllValidationResult{}
	runPortsValidation(deployments, result)

	require.Len(t, result.Results, 1)

	// May error or report invalid.
	require.Equal(t, validatorNamePorts, result.Results[0].Name)
}

func TestRunPortsValidation_SkipsInfrastructure(t *testing.T) {
	t.Parallel()

	deployments := []deploymentEntry{
		{path: "/tmp/shared-postgres", name: postgresSharedDirName, level: DeploymentTypeInfrastructure},
		{path: "/tmp/template", name: cryptoutilSharedMagic.SkeletonTemplateServiceName, level: DeploymentTypeTemplate},
	}

	result := &AllValidationResult{}
	runPortsValidation(deployments, result)

	// Infrastructure and template are skipped.
	require.Empty(t, result.Results)
}

func TestRunTelemetryValidation_ErrorPath(t *testing.T) {
	t.Parallel()

	result := &AllValidationResult{}
	runTelemetryValidation(testNonexistentPath, result)

	require.Len(t, result.Results, 1)
	require.Equal(t, validatorNameTelemetry, result.Results[0].Name)
}

func TestRunAdminValidation_ErrorPath(t *testing.T) {
	t.Parallel()

	deployments := []deploymentEntry{
		{path: testNonexistentPath, name: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, level: DeploymentTypeProductService},
	}

	result := &AllValidationResult{}
	runAdminValidation(deployments, result)

	require.Len(t, result.Results, 1)
	require.Equal(t, validatorNameAdmin, result.Results[0].Name)
}

func TestRunSecretsValidation_ErrorPath(t *testing.T) {
	t.Parallel()

	deployments := []deploymentEntry{
		{path: testNonexistentPath, name: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, level: DeploymentTypeProductService},
	}

	result := &AllValidationResult{}
	runSecretsValidation(deployments, result)

	require.Len(t, result.Results, 1)
	require.Equal(t, validatorNameSecrets, result.Results[0].Name)
}

func TestRunCertVolumePolicyValidation(t *testing.T) {
	t.Parallel()

	t.Run("runs for service deployment levels", func(t *testing.T) {
		t.Parallel()

		deployments := []deploymentEntry{
			{path: "/nonexistent/path/service", name: cryptoutilSharedMagic.OTLPServiceSMKMS, level: DeploymentTypeProductService},
			{path: "/nonexistent/path/template", name: cryptoutilSharedMagic.SkeletonTemplateServiceName, level: DeploymentTypeTemplate},
			{path: "/nonexistent/path/product", name: cryptoutilSharedMagic.SMProductName, level: DeploymentTypeProduct},
			{path: "/nonexistent/path/suite", name: cryptoutilSharedMagic.DefaultOTLPServiceDefault, level: DeploymentTypeSuite},
			{path: "/nonexistent/path/infra", name: postgresSharedDirName, level: DeploymentTypeInfrastructure},
		}

		result := &AllValidationResult{}
		runCertVolumePolicyValidation(deployments, result)

		require.Len(t, result.Results, 4)

		for _, vr := range result.Results {
			require.Equal(t, validatorNameCertVolumes, vr.Name)
		}
	})
}

func TestRunPostgresSecretsDirSyncValidation_ErrorPath(t *testing.T) {
	t.Parallel()

	result := &AllValidationResult{}
	runPostgresSecretsDirSyncValidation(testNonexistentPath, result)

	require.Len(t, result.Results, 1)
	require.Equal(t, validatorNamePostgresSync, result.Results[0].Name)
	require.False(t, result.Results[0].Passed)
}

func TestIsServiceFrameworkConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{name: "sm-kms-app-framework-postgresql-1.yml", path: "deployments/sm-kms/config/sm-kms-app-framework-postgresql-1.yml", expected: true},
		{name: "sm-kms-app-framework-sqlite-1.yml", path: "deployments/sm-kms/config/sm-kms-app-framework-sqlite-1.yml", expected: true},
		{name: "sm-kms-app-framework-postgresql-2.yml", path: "deployments/sm-kms/config/sm-kms-app-framework-postgresql-2.yml", expected: true},
		{name: "sm-kms-app-framework-common.yml", path: "configs/sm-kms/sm-kms-app-framework-common.yml", expected: true},
		{name: "config-pg-2.yaml", path: "/tmp/config-pg-2.yaml", expected: true},
		{name: testSMKMSServerConfigFilename, path: filepath.Join(cryptoutilSharedMagic.CICDConfigsDir, cryptoutilSharedMagic.OTLPServiceSMKMS, testSMKMSServerConfigFilename), expected: false},
		{name: testSMKMSServerConfigFilename, path: filepath.Join(cryptoutilSharedMagic.CICDConfigsDir, cryptoutilSharedMagic.OTLPServiceSMKMS, testSMKMSServerConfigFilename), expected: false},
		{name: testSMKMSConfigSchemaFilename, path: filepath.Join(cryptoutilSharedMagic.CICDConfigsDir, cryptoutilSharedMagic.OTLPServiceSMKMS, testSMKMSConfigSchemaFilename), expected: false},
		{name: "adaptive-auth.yml", path: "configs/identity-authz/domain/policies/adaptive-auth.yml", expected: false},
		{name: "authz-docker.yml", path: "configs/identity-authz/authz-docker.yml", expected: false},
		{name: cryptoutilSharedMagic.COMPOSE_YML, path: "deployments/compose/compose.yml", expected: false},
		{name: "random.yml", path: "/tmp/random.yml", expected: false},
		{name: "non-yaml readme.md", path: "/tmp/readme.md", expected: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := isServiceFrameworkConfig(tc.path)
			require.Equal(t, tc.expected, got)
		})
	}
}
