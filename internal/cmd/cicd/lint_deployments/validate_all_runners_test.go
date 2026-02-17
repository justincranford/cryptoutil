package lint_deployments

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunNamingValidation_InvalidPaths(t *testing.T) {
	t.Parallel()

	result := &AllValidationResult{}
	runNamingValidation("/nonexistent/path/abc123", "/nonexistent/path/def456", result)

	require.Len(t, result.Results, 2)
	assert.False(t, result.Results[0].Passed)
	assert.Equal(t, validatorNameNaming, result.Results[0].Name)
	assert.False(t, result.Results[1].Passed)
	assert.Equal(t, validatorNameNaming, result.Results[1].Name)
}

func TestRunKebabCaseValidation_InvalidPath(t *testing.T) {
	t.Parallel()

	result := &AllValidationResult{}
	runKebabCaseValidation("/nonexistent/path/abc123", result)

	require.Len(t, result.Results, 1)
	assert.False(t, result.Results[0].Passed)
	assert.Equal(t, validatorNameKebabCase, result.Results[0].Name)
}

func TestRunSchemaValidation_ErrorPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create an unreadable YAML file.
	unreadable := filepath.Join(dir, "bad.yml")
	require.NoError(t, os.WriteFile(unreadable, []byte("key: val"), 0o600))
	require.NoError(t, os.Chmod(unreadable, 0o000))

	t.Cleanup(func() {
		_ = os.Chmod(unreadable, 0o600)
	})

	result := &AllValidationResult{}
	runSchemaValidation(dir, result)

	// Schema validator may error or report invalid depending on permissions.
	assert.Greater(t, len(result.Results), 0)
}

func TestRunTemplatePatternValidation_NonexistentSkips(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	result := &AllValidationResult{}
	runTemplatePatternValidation(dir, result)

	// No template/ subdir means no results added.
	assert.Empty(t, result.Results)
}

func TestRunTemplatePatternValidation_ErrorPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	templateDir := filepath.Join(dir, "template")
	require.NoError(t, os.MkdirAll(templateDir, 0o755))

	result := &AllValidationResult{}
	runTemplatePatternValidation(dir, result)

	// Template dir exists but is empty - validator runs and reports.
	assert.Greater(t, len(result.Results), 0)
}

func TestRunPortsValidation_ErrorPath(t *testing.T) {
	t.Parallel()

	deployments := []deploymentEntry{
		{path: "/nonexistent/path/abc123", name: "jose-ja", level: DeploymentTypeProductService},
	}

	result := &AllValidationResult{}
	runPortsValidation(deployments, result)

	require.Len(t, result.Results, 1)

	// May error or report invalid.
	assert.Equal(t, validatorNamePorts, result.Results[0].Name)
}

func TestRunPortsValidation_SkipsInfrastructure(t *testing.T) {
	t.Parallel()

	deployments := []deploymentEntry{
		{path: "/tmp/shared-postgres", name: "shared-postgres", level: DeploymentTypeInfrastructure},
		{path: "/tmp/template", name: "template", level: DeploymentTypeTemplate},
	}

	result := &AllValidationResult{}
	runPortsValidation(deployments, result)

	// Infrastructure and template are skipped.
	assert.Empty(t, result.Results)
}

func TestRunTelemetryValidation_ErrorPath(t *testing.T) {
	t.Parallel()

	result := &AllValidationResult{}
	runTelemetryValidation("/nonexistent/path/abc123", result)

	require.Len(t, result.Results, 1)
	assert.Equal(t, validatorNameTelemetry, result.Results[0].Name)
}

func TestRunAdminValidation_ErrorPath(t *testing.T) {
	t.Parallel()

	deployments := []deploymentEntry{
		{path: "/nonexistent/path/abc123", name: "jose-ja", level: DeploymentTypeProductService},
	}

	result := &AllValidationResult{}
	runAdminValidation(deployments, result)

	require.Len(t, result.Results, 1)
	assert.Equal(t, validatorNameAdmin, result.Results[0].Name)
}

func TestRunSecretsValidation_ErrorPath(t *testing.T) {
	t.Parallel()

	deployments := []deploymentEntry{
		{path: "/nonexistent/path/abc123", name: "jose-ja", level: DeploymentTypeProductService},
	}

	result := &AllValidationResult{}
	runSecretsValidation(deployments, result)

	require.Len(t, result.Results, 1)
	assert.Equal(t, validatorNameSecrets, result.Results[0].Name)
}
