package lint_deployments

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestClassifyDeployment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "service jose-ja", input: cryptoutilSharedMagic.OTLPServiceJoseJA, expected: DeploymentTypeProductService},
		{name: "service sm-im", input: cryptoutilSharedMagic.OTLPServiceSMIM, expected: DeploymentTypeProductService},
		{name: "service pki-ca", input: cryptoutilSharedMagic.OTLPServicePKICA, expected: DeploymentTypeProductService},
		{name: "service sm-kms", input: cryptoutilSharedMagic.OTLPServiceSMKMS, expected: DeploymentTypeProductService},
		{name: "service identity-authz", input: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, expected: DeploymentTypeProductService},
		{name: "service identity-idp", input: cryptoutilSharedMagic.OTLPServiceIdentityIDP, expected: DeploymentTypeProductService},
		{name: "service identity-rp", input: cryptoutilSharedMagic.OTLPServiceIdentityRP, expected: DeploymentTypeProductService},
		{name: "service identity-rs", input: cryptoutilSharedMagic.OTLPServiceIdentityRS, expected: DeploymentTypeProductService},
		{name: "service identity-spa", input: cryptoutilSharedMagic.OTLPServiceIdentitySPA, expected: DeploymentTypeProductService},
		{name: "service skeleton-template", input: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, expected: DeploymentTypeProductService},
		{name: "product identity", input: cryptoutilSharedMagic.IdentityProductName, expected: DeploymentTypeProduct},
		{name: "product sm", input: cryptoutilSharedMagic.SMProductName, expected: DeploymentTypeProduct},
		{name: "product pki", input: cryptoutilSharedMagic.PKIProductName, expected: DeploymentTypeProduct},
		{name: "product jose", input: cryptoutilSharedMagic.JoseProductName, expected: DeploymentTypeProduct},
		{name: "product skeleton", input: cryptoutilSharedMagic.SkeletonProductName, expected: DeploymentTypeProduct},
		{name: "suite cryptoutil-suite", input: "cryptoutil-suite", expected: DeploymentTypeSuite},
		{name: cryptoutilSharedMagic.SkeletonTemplateServiceName, input: cryptoutilSharedMagic.SkeletonTemplateServiceName, expected: DeploymentTypeTemplate},
		{name: "infrastructure shared-postgres", input: "shared-postgres", expected: DeploymentTypeInfrastructure},
		{name: "infrastructure shared-citus", input: "shared-citus", expected: DeploymentTypeInfrastructure},
		{name: "infrastructure shared-telemetry", input: "shared-telemetry", expected: DeploymentTypeInfrastructure},

		{name: "unknown dir", input: "random-dir", expected: DeploymentTypeInfrastructure},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := classifyDeployment(tc.input)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestDiscoverDeploymentDirs(t *testing.T) {
	t.Parallel()

	t.Run("valid directory with mixed entries", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, cryptoutilSharedMagic.OTLPServiceJoseJA), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(filepath.Join(dir, cryptoutilSharedMagic.IdentityProductName), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "cryptoutil-suite"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "shared-postgres"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("readme"), cryptoutilSharedMagic.CacheFilePermissions))

		result := discoverDeploymentDirs(dir)
		require.Len(t, result, 4)

		nameMap := make(map[string]string)
		for _, d := range result {
			nameMap[d.name] = d.level
		}

		require.Equal(t, DeploymentTypeProductService, nameMap[cryptoutilSharedMagic.OTLPServiceJoseJA])
		require.Equal(t, DeploymentTypeProduct, nameMap[cryptoutilSharedMagic.IdentityProductName])
		require.Equal(t, DeploymentTypeSuite, nameMap["cryptoutil-suite"])
		require.Equal(t, DeploymentTypeInfrastructure, nameMap["shared-postgres"])
	})

	t.Run("nonexistent directory returns empty", func(t *testing.T) {
		t.Parallel()

		result := discoverDeploymentDirs("/nonexistent/path/abc123")
		require.Empty(t, result)
	})

	t.Run("empty directory", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		result := discoverDeploymentDirs(dir)
		require.Empty(t, result)
	})
}

func TestDiscoverConfigFiles(t *testing.T) {
	t.Parallel()

	t.Run("finds yaml files recursively", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, cryptoutilSharedMagic.ClaimSub), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yml"), []byte("key: val"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(dir, cryptoutilSharedMagic.ClaimSub, "nested.yaml"), []byte("key: val"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "readme.md"), []byte("readme"), cryptoutilSharedMagic.CacheFilePermissions))

		files := discoverConfigFiles(dir)
		require.Len(t, files, 2)
	})

	t.Run("nonexistent directory returns empty", func(t *testing.T) {
		t.Parallel()

		files := discoverConfigFiles("/nonexistent/path/xyz789")
		require.Empty(t, files)
	})

	t.Run("no yaml files", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "data.json"), []byte("{}"), cryptoutilSharedMagic.CacheFilePermissions))

		files := discoverConfigFiles(dir)
		require.Empty(t, files)
	})
}

func TestAllValidationResult_AllPassed(t *testing.T) {
	t.Parallel()

	t.Run("empty results returns true", func(t *testing.T) {
		t.Parallel()

		r := &AllValidationResult{}
		require.True(t, r.AllPassed())
	})

	t.Run("all passed returns true", func(t *testing.T) {
		t.Parallel()

		r := &AllValidationResult{
			Results: []ValidatorResult{
				{Name: "a", Passed: true},
				{Name: "b", Passed: true},
			},
		}
		require.True(t, r.AllPassed())
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
		require.False(t, r.AllPassed())
	})

	t.Run("first failed returns false", func(t *testing.T) {
		t.Parallel()

		r := &AllValidationResult{
			Results: []ValidatorResult{
				{Name: "a", Passed: false},
			},
		}
		require.False(t, r.AllPassed())
	})
}

func TestAllValidationResult_AddResult(t *testing.T) {
	t.Parallel()

	r := &AllValidationResult{}
	r.addResult("test-validator", "/tmp/target", true, "output text", cryptoutilSharedMagic.JoseJAMaxMaterials*time.Millisecond)

	require.Len(t, r.Results, 1)
	require.Equal(t, "test-validator", r.Results[0].Name)
	require.Equal(t, "/tmp/target", r.Results[0].Target)
	require.True(t, r.Results[0].Passed)
	require.Equal(t, "output text", r.Results[0].Output)
	require.Equal(t, cryptoutilSharedMagic.JoseJAMaxMaterials*time.Millisecond, r.Results[0].Duration)
}

func TestFormatAllValidationResult(t *testing.T) {
	t.Parallel()

	t.Run("all passed", func(t *testing.T) {
		t.Parallel()

		r := &AllValidationResult{
			Results: []ValidatorResult{
				{Name: "naming", Target: "deployments", Passed: true, Duration: cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Millisecond},
				{Name: "schema", Target: "configs/test.yml", Passed: true, Duration: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Millisecond},
			},
			TotalDuration: 15 * time.Millisecond,
		}

		output := FormatAllValidationResult(r)
		require.Contains(t, output, "=== Validate All: Aggregated Results ===")
		require.Contains(t, output, "[PASS] naming")
		require.Contains(t, output, "[PASS] schema")
		require.Contains(t, output, "Passed:   2")
		require.Contains(t, output, "Failed:   0")
		require.Contains(t, output, "ALL VALIDATORS PASSED")
		require.NotContains(t, output, "VALIDATION FAILED")
	})

	t.Run("some failed", func(t *testing.T) {
		t.Parallel()

		r := &AllValidationResult{
			Results: []ValidatorResult{
				{Name: "naming", Target: "deployments", Passed: true, Duration: cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Millisecond},
				{Name: "ports", Target: "deployments/jose-ja", Passed: false, Duration: cryptoutilSharedMagic.MaxErrorDisplay * time.Millisecond},
				{Name: "admin", Target: "deployments/sm-im", Passed: false, Duration: 15 * time.Millisecond},
			},
			TotalDuration: 45 * time.Millisecond,
		}

		output := FormatAllValidationResult(r)
		require.Contains(t, output, "[PASS] naming")
		require.Contains(t, output, "[FAIL] ports")
		require.Contains(t, output, "[FAIL] admin")
		require.Contains(t, output, "Passed:   1")
		require.Contains(t, output, "Failed:   2")
		require.Contains(t, output, "VALIDATION FAILED")
		require.Contains(t, output, "Failed validators:")
		require.Contains(t, output, "- ports (deployments/jose-ja)")
		require.Contains(t, output, "- admin (deployments/sm-im)")
		require.NotContains(t, output, "ALL VALIDATORS PASSED")
	})

	t.Run("empty results", func(t *testing.T) {
		t.Parallel()

		r := &AllValidationResult{
			TotalDuration: 0,
		}

		output := FormatAllValidationResult(r)
		require.Contains(t, output, "Total:    0 validators")
		require.Contains(t, output, "ALL VALIDATORS PASSED")
	})
}

func TestValidateAll_EmptyDirs(t *testing.T) {
	t.Parallel()

	deploymentsDir := t.TempDir()
	configsDir := t.TempDir()

	result := ValidateAll(deploymentsDir, configsDir)
	require.NotNil(t, result)
	require.True(t, result.AllPassed())
	require.Greater(t, len(result.Results), 0)
}

func TestValidateAll_WithDeployments(t *testing.T) {
	t.Parallel()

	deploymentsDir := t.TempDir()
	configsDir := t.TempDir()

	// Create a service deployment with compose.yml.
	svcDir := filepath.Join(deploymentsDir, cryptoutilSharedMagic.OTLPServiceJoseJA)
	require.NoError(t, os.MkdirAll(filepath.Join(svcDir, "secrets"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(svcDir, "config"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(svcDir, "compose.yml"), []byte("services:\n  jose-ja:\n    image: test\n"), cryptoutilSharedMagic.CacheFilePermissions))

	// Create a config file.
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, "test.yml"), []byte("bind-public-port: 8080\n"), cryptoutilSharedMagic.CacheFilePermissions))

	result := ValidateAll(deploymentsDir, configsDir)
	require.NotNil(t, result)
	require.Greater(t, len(result.Results), 0)
	require.Greater(t, result.TotalDuration, time.Duration(0))
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

	start := time.Now().UTC()

	result := ValidateAll(deploymentsDir, configsDir)
	elapsed := time.Since(start)

	require.NotNil(t, result)
	require.Greater(t, len(result.Results), 0)

	// Performance target: <5s (Decision 5:C).
	maxDuration := cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second
	require.Less(t, elapsed, maxDuration, "validate-all should complete in <5s, took %s", elapsed)

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
		require.Equal(t, 1, exitCode)
	})

	t.Run("missing configs dir", func(t *testing.T) {
		t.Parallel()

		deploymentsDir := t.TempDir()
		exitCode := mainValidateAll([]string{deploymentsDir, "/nonexistent/configs"})
		require.Equal(t, 1, exitCode)
	})

	t.Run("empty dirs succeed", func(t *testing.T) {
		t.Parallel()

		deploymentsDir := t.TempDir()
		configsDir := t.TempDir()
		exitCode := mainValidateAll([]string{deploymentsDir, configsDir})
		require.Equal(t, 0, exitCode)
	})

	t.Run("defaults used when no args", func(t *testing.T) {
		t.Parallel()

		// With no args, uses defaultDeploymentsDir/defaultConfigsDir which may not exist.
		// This tests the default path logic.
		exitCode := mainValidateAll([]string{})
		// May return 0 or 1 depending on whether deployments/ and configs/ exist from CWD.
		require.Contains(t, []int{0, 1}, exitCode)
	})

	t.Run("only one arg uses defaults for configs", func(t *testing.T) {
		t.Parallel()

		deploymentsDir := t.TempDir()
		// Only 1 arg: deploymentsDir provided but configsDir defaults.
		exitCode := mainValidateAll([]string{deploymentsDir})
		// configsDir defaults to "configs" which may not exist.
		require.Contains(t, []int{0, 1}, exitCode)
	})
}

func TestValidatorResultFields(t *testing.T) {
	t.Parallel()

	vr := ValidatorResult{
		Name:     validatorNameNaming,
		Target:   "/tmp/test",
		Passed:   true,
		Output:   "sample output",
		Duration: cryptoutilSharedMagic.IMMaxUsernameLength * time.Millisecond,
	}

	require.Equal(t, "naming", vr.Name)
	require.Equal(t, "/tmp/test", vr.Target)
	require.True(t, vr.Passed)
	require.Equal(t, "sample output", vr.Output)
	require.Equal(t, cryptoutilSharedMagic.IMMaxUsernameLength*time.Millisecond, vr.Duration)
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
			require.Equal(t, tc.expected, tc.constant)
		})
	}
}
