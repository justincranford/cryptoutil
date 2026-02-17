package lint_deployments

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifyDeployment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "service jose-ja", input: "jose-ja", expected: DeploymentTypeProductService},
		{name: "service cipher-im", input: "cipher-im", expected: DeploymentTypeProductService},
		{name: "service pki-ca", input: "pki-ca", expected: DeploymentTypeProductService},
		{name: "service sm-kms", input: "sm-kms", expected: DeploymentTypeProductService},
		{name: "service identity-authz", input: "identity-authz", expected: DeploymentTypeProductService},
		{name: "service identity-idp", input: "identity-idp", expected: DeploymentTypeProductService},
		{name: "service identity-rp", input: "identity-rp", expected: DeploymentTypeProductService},
		{name: "service identity-rs", input: "identity-rs", expected: DeploymentTypeProductService},
		{name: "service identity-spa", input: "identity-spa", expected: DeploymentTypeProductService},
		{name: "product identity", input: "identity", expected: DeploymentTypeProduct},
		{name: "product sm", input: "sm", expected: DeploymentTypeProduct},
		{name: "product pki", input: "pki", expected: DeploymentTypeProduct},
		{name: "product cipher", input: "cipher", expected: DeploymentTypeProduct},
		{name: "product jose", input: "jose", expected: DeploymentTypeProduct},
		{name: "suite cryptoutil", input: "cryptoutil", expected: DeploymentTypeSuite},
		{name: "template", input: "template", expected: DeploymentTypeTemplate},
		{name: "infrastructure shared-postgres", input: "shared-postgres", expected: DeploymentTypeInfrastructure},
		{name: "infrastructure shared-citus", input: "shared-citus", expected: DeploymentTypeInfrastructure},
		{name: "infrastructure shared-telemetry", input: "shared-telemetry", expected: DeploymentTypeInfrastructure},
		{name: "infrastructure compose", input: "compose", expected: DeploymentTypeInfrastructure},
		{name: "unknown dir", input: "random-dir", expected: DeploymentTypeInfrastructure},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := classifyDeployment(tc.input)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestDiscoverDeploymentDirs(t *testing.T) {
	t.Parallel()

	t.Run("valid directory with mixed entries", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "jose-ja"), 0o755))
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "identity"), 0o755))
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "cryptoutil"), 0o755))
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "shared-postgres"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("readme"), 0o600))

		result := discoverDeploymentDirs(dir)
		assert.Len(t, result, 4)

		nameMap := make(map[string]string)
		for _, d := range result {
			nameMap[d.name] = d.level
		}

		assert.Equal(t, DeploymentTypeProductService, nameMap["jose-ja"])
		assert.Equal(t, DeploymentTypeProduct, nameMap["identity"])
		assert.Equal(t, DeploymentTypeSuite, nameMap["cryptoutil"])
		assert.Equal(t, DeploymentTypeInfrastructure, nameMap["shared-postgres"])
	})

	t.Run("nonexistent directory returns empty", func(t *testing.T) {
		t.Parallel()

		result := discoverDeploymentDirs("/nonexistent/path/abc123")
		assert.Empty(t, result)
	})

	t.Run("empty directory", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		result := discoverDeploymentDirs(dir)
		assert.Empty(t, result)
	})
}

func TestDiscoverConfigFiles(t *testing.T) {
	t.Parallel()

	t.Run("finds yaml files recursively", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yml"), []byte("key: val"), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "nested.yaml"), []byte("key: val"), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "readme.md"), []byte("readme"), 0o600))

		files := discoverConfigFiles(dir)
		assert.Len(t, files, 2)
	})

	t.Run("nonexistent directory returns empty", func(t *testing.T) {
		t.Parallel()

		files := discoverConfigFiles("/nonexistent/path/xyz789")
		assert.Empty(t, files)
	})

	t.Run("no yaml files", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "data.json"), []byte("{}"), 0o600))

		files := discoverConfigFiles(dir)
		assert.Empty(t, files)
	})
}

func TestAllValidationResult_AllPassed(t *testing.T) {
	t.Parallel()

	t.Run("empty results returns true", func(t *testing.T) {
		t.Parallel()

		r := &AllValidationResult{}
		assert.True(t, r.AllPassed())
	})

	t.Run("all passed returns true", func(t *testing.T) {
		t.Parallel()

		r := &AllValidationResult{
			Results: []ValidatorResult{
				{Name: "a", Passed: true},
				{Name: "b", Passed: true},
			},
		}
		assert.True(t, r.AllPassed())
	})

	t.Run("one failed returns false", func(t *testing.T) {
		t.Parallel()

		r := &AllValidationResult{
			Results: []ValidatorResult{
				{Name: "a", Passed: true},
				{Name: "b", Passed: false},
				{Name: "c", Passed: true},
			},
		}
		assert.False(t, r.AllPassed())
	})

	t.Run("first failed returns false", func(t *testing.T) {
		t.Parallel()

		r := &AllValidationResult{
			Results: []ValidatorResult{
				{Name: "a", Passed: false},
			},
		}
		assert.False(t, r.AllPassed())
	})
}

func TestAllValidationResult_AddResult(t *testing.T) {
	t.Parallel()

	r := &AllValidationResult{}
	r.addResult("test-validator", "/tmp/target", true, "output text", 100*time.Millisecond)

	require.Len(t, r.Results, 1)
	assert.Equal(t, "test-validator", r.Results[0].Name)
	assert.Equal(t, "/tmp/target", r.Results[0].Target)
	assert.True(t, r.Results[0].Passed)
	assert.Equal(t, "output text", r.Results[0].Output)
	assert.Equal(t, 100*time.Millisecond, r.Results[0].Duration)
}

func TestFormatAllValidationResult(t *testing.T) {
	t.Parallel()

	t.Run("all passed", func(t *testing.T) {
		t.Parallel()

		r := &AllValidationResult{
			Results: []ValidatorResult{
				{Name: "naming", Target: "deployments", Passed: true, Duration: 10 * time.Millisecond},
				{Name: "schema", Target: "configs/test.yml", Passed: true, Duration: 5 * time.Millisecond},
			},
			TotalDuration: 15 * time.Millisecond,
		}

		output := FormatAllValidationResult(r)
		assert.Contains(t, output, "=== Validate All: Aggregated Results ===")
		assert.Contains(t, output, "[PASS] naming")
		assert.Contains(t, output, "[PASS] schema")
		assert.Contains(t, output, "Passed:   2")
		assert.Contains(t, output, "Failed:   0")
		assert.Contains(t, output, "ALL VALIDATORS PASSED")
		assert.NotContains(t, output, "VALIDATION FAILED")
	})

	t.Run("some failed", func(t *testing.T) {
		t.Parallel()

		r := &AllValidationResult{
			Results: []ValidatorResult{
				{Name: "naming", Target: "deployments", Passed: true, Duration: 10 * time.Millisecond},
				{Name: "ports", Target: "deployments/jose-ja", Passed: false, Duration: 20 * time.Millisecond},
				{Name: "admin", Target: "deployments/cipher-im", Passed: false, Duration: 15 * time.Millisecond},
			},
			TotalDuration: 45 * time.Millisecond,
		}

		output := FormatAllValidationResult(r)
		assert.Contains(t, output, "[PASS] naming")
		assert.Contains(t, output, "[FAIL] ports")
		assert.Contains(t, output, "[FAIL] admin")
		assert.Contains(t, output, "Passed:   1")
		assert.Contains(t, output, "Failed:   2")
		assert.Contains(t, output, "VALIDATION FAILED")
		assert.Contains(t, output, "Failed validators:")
		assert.Contains(t, output, "- ports (deployments/jose-ja)")
		assert.Contains(t, output, "- admin (deployments/cipher-im)")
		assert.NotContains(t, output, "ALL VALIDATORS PASSED")
	})

	t.Run("empty results", func(t *testing.T) {
		t.Parallel()

		r := &AllValidationResult{
			TotalDuration: 0,
		}

		output := FormatAllValidationResult(r)
		assert.Contains(t, output, "Total:    0 validators")
		assert.Contains(t, output, "ALL VALIDATORS PASSED")
	})
}

func TestValidateAll_EmptyDirs(t *testing.T) {
	t.Parallel()

	deploymentsDir := t.TempDir()
	configsDir := t.TempDir()

	result := ValidateAll(deploymentsDir, configsDir)
	assert.NotNil(t, result)
	assert.True(t, result.AllPassed())
	assert.Greater(t, len(result.Results), 0)
}

func TestValidateAll_WithDeployments(t *testing.T) {
	t.Parallel()

	deploymentsDir := t.TempDir()
	configsDir := t.TempDir()

	// Create a service deployment with compose.yml.
	svcDir := filepath.Join(deploymentsDir, "jose-ja")
	require.NoError(t, os.MkdirAll(filepath.Join(svcDir, "secrets"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(svcDir, "config"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(svcDir, "compose.yml"), []byte("services:\n  jose-ja:\n    image: test\n"), 0o600))

	// Create a config file.
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, "test.yml"), []byte("bind-public-port: 8080\n"), 0o600))

	result := ValidateAll(deploymentsDir, configsDir)
	assert.NotNil(t, result)
	assert.Greater(t, len(result.Results), 0)
	assert.Greater(t, result.TotalDuration, time.Duration(0))
}

func TestValidateAll_RealDeployments(t *testing.T) {
	t.Parallel()

	deploymentsDir := "../../../../deployments"
	configsDir := "../../../../configs"

	if _, err := os.Stat(deploymentsDir); os.IsNotExist(err) {
		t.Skip("deployments/ directory not found (not running from project root)")
	}

	if _, err := os.Stat(configsDir); os.IsNotExist(err) {
		t.Skip("configs/ directory not found (not running from project root)")
	}

	start := time.Now()

	result := ValidateAll(deploymentsDir, configsDir)
	elapsed := time.Since(start)

	assert.NotNil(t, result)
	assert.Greater(t, len(result.Results), 0)

	// Performance target: <5s (Decision 5:C).
	maxDuration := 5 * time.Second
	assert.Less(t, elapsed, maxDuration, "validate-all should complete in <5s, took %s", elapsed)

	// Log results for debugging.
	t.Logf("Total validators: %d, Duration: %s", len(result.Results), elapsed)

	for i := range result.Results {
		vr := &result.Results[i]

		status := statusPass
		if !vr.Passed {
			status = statusFail
		}

		t.Logf("[%s] %s (%s) [%s]", status, vr.Name, vr.Target, vr.Duration)
	}
}

func TestMainValidateAll(t *testing.T) {
	t.Parallel()

	t.Run("missing deployments dir", func(t *testing.T) {
		t.Parallel()

		exitCode := mainValidateAll([]string{"/nonexistent/deployments", "/nonexistent/configs"})
		assert.Equal(t, 1, exitCode)
	})

	t.Run("missing configs dir", func(t *testing.T) {
		t.Parallel()

		deploymentsDir := t.TempDir()
		exitCode := mainValidateAll([]string{deploymentsDir, "/nonexistent/configs"})
		assert.Equal(t, 1, exitCode)
	})

	t.Run("empty dirs succeed", func(t *testing.T) {
		t.Parallel()

		deploymentsDir := t.TempDir()
		configsDir := t.TempDir()
		exitCode := mainValidateAll([]string{deploymentsDir, configsDir})
		assert.Equal(t, 0, exitCode)
	})

	t.Run("defaults used when no args", func(t *testing.T) {
		t.Parallel()

		// With no args, uses defaultDeploymentsDir/defaultConfigsDir which may not exist.
		// This tests the default path logic.
		exitCode := mainValidateAll([]string{})
		// May return 0 or 1 depending on whether deployments/ and configs/ exist from CWD.
		assert.Contains(t, []int{0, 1}, exitCode)
	})

	t.Run("only one arg uses defaults for configs", func(t *testing.T) {
		t.Parallel()

		deploymentsDir := t.TempDir()
		// Only 1 arg: deploymentsDir provided but configsDir defaults.
		exitCode := mainValidateAll([]string{deploymentsDir})
		// configsDir defaults to "configs" which may not exist.
		assert.Contains(t, []int{0, 1}, exitCode)
	})
}

func TestValidatorResultFields(t *testing.T) {
	t.Parallel()

	vr := ValidatorResult{
		Name:     validatorNameNaming,
		Target:   "/tmp/test",
		Passed:   true,
		Output:   "sample output",
		Duration: 50 * time.Millisecond,
	}

	assert.Equal(t, "naming", vr.Name)
	assert.Equal(t, "/tmp/test", vr.Target)
	assert.True(t, vr.Passed)
	assert.Equal(t, "sample output", vr.Output)
	assert.Equal(t, 50*time.Millisecond, vr.Duration)
}

func TestValidatorNameConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{name: "naming", constant: validatorNameNaming, expected: "naming"},
		{name: "kebab-case", constant: validatorNameKebabCase, expected: "kebab-case"},
		{name: "schema", constant: validatorNameSchema, expected: "schema"},
		{name: "template-pattern", constant: validatorNameTemplatePattern, expected: "template-pattern"},
		{name: "ports", constant: validatorNamePorts, expected: "ports"},
		{name: "telemetry", constant: validatorNameTelemetry, expected: "telemetry"},
		{name: "admin", constant: validatorNameAdmin, expected: "admin"},
		{name: "secrets", constant: validatorNameSecrets, expected: "secrets"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, tc.constant)
		})
	}
}

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
