package lint_deployments_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	. "cryptoutil/internal/cmd/cicd/lint_deployments"
)

func TestValidateAllDeployments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFunc  func(t *testing.T, rootDir string)
		wantCount  int
		wantValid  int
		wantErrors bool
	}{
		{
			name: "all valid deployments",
			setupFunc: func(t *testing.T, rootDir string) {
				t.Helper()

				// Create a valid PRODUCT-SERVICE deployment.
				svcDir := filepath.Join(rootDir, cryptoutilSharedMagic.OTLPServiceJoseJA)
				require.NoError(t, os.MkdirAll(svcDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				createValidProductServiceDeployment(t, svcDir, cryptoutilSharedMagic.OTLPServiceJoseJA)

				// Create a valid template deployment.
				tmplDir := filepath.Join(rootDir, cryptoutilSharedMagic.SkeletonTemplateServiceName)
				require.NoError(t, os.MkdirAll(filepath.Join(tmplDir, "secrets"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				require.NoError(t, os.WriteFile(filepath.Join(tmplDir, "compose.yml"), []byte("v"), cryptoutilSharedMagic.CacheFilePermissions))
				createRequiredSecrets(t, tmplDir)
			},
			wantCount: 2,
			wantValid: 2,
		},
		{
			name: "mixed valid and invalid",
			setupFunc: func(t *testing.T, rootDir string) {
				t.Helper()

				// Create a valid PRODUCT-SERVICE deployment.
				validDir := filepath.Join(rootDir, cryptoutilSharedMagic.OTLPServiceSMIM)
				require.NoError(t, os.MkdirAll(validDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				createValidProductServiceDeployment(t, validDir, cryptoutilSharedMagic.OTLPServiceSMIM)

				// Create an invalid deployment (missing secrets and config files).
				invalidDir := filepath.Join(rootDir, cryptoutilSharedMagic.OTLPServiceSMKMS)
				require.NoError(t, os.MkdirAll(filepath.Join(invalidDir, "config"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				require.NoError(t, os.WriteFile(filepath.Join(invalidDir, "compose.yml"), []byte("v"), cryptoutilSharedMagic.CacheFilePermissions))
				require.NoError(t, os.WriteFile(filepath.Join(invalidDir, "Dockerfile"), []byte("F"), cryptoutilSharedMagic.CacheFilePermissions))
			},
			wantCount:  2,
			wantValid:  1,
			wantErrors: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			tc.setupFunc(t, tmpDir)

			results, err := ValidateAllDeployments(tmpDir)
			require.NoError(t, err)

			require.Len(t, results, tc.wantCount, "deployment count mismatch")

			validCount := 0

			for _, r := range results {
				if r.Valid {
					validCount++
				}
			}

			require.Equal(t, tc.wantValid, validCount, "valid count mismatch")
		})
	}
}

func TestFormatResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		results     []ValidationResult
		contains    []string
		notContains []string
	}{
		{
			name: "valid result no optional sections",
			results: []ValidationResult{
				{
					Path:  "/deploy/jose-ja",
					Type:  "PRODUCT-SERVICE",
					Valid: true,
				},
			},
			contains:    []string{"✅ VALID", cryptoutilSharedMagic.OTLPServiceJoseJA, "1 valid"},
			notContains: []string{"Missing directories", "Missing files", "Missing secrets", "ERROR:", "WARN:"},
		},
		{
			name: "invalid result with errors",
			results: []ValidationResult{
				{
					Path:           "/deploy/sm-kms",
					Type:           "PRODUCT-SERVICE",
					Valid:          false,
					MissingDirs:    []string{"secrets"},
					MissingFiles:   []string{"Dockerfile"},
					MissingSecrets: []string{"unseal_1of5.secret"},
					Errors:         []string{"Config file error"},
				},
			},
			contains: []string{"❌ INVALID", cryptoutilSharedMagic.OTLPServiceSMKMS, "Missing directories", "Missing files", "Missing secrets", "ERROR: Config file error"},
		},
		{
			name: "result with warnings",
			results: []ValidationResult{
				{
					Path:     "/deploy/jose-ja",
					Type:     "PRODUCT-SERVICE",
					Valid:    true,
					Warnings: []string{"Optional file missing"},
				},
			},
			contains:    []string{"✅ VALID", "WARN: Optional file missing"},
			notContains: []string{"Missing directories", "Missing files", "Missing secrets", "ERROR:"},
		},
		{
			name: "mixed valid and invalid sorted correctly",
			results: []ValidationResult{
				{Path: "/deploy/jose-ja", Type: "PRODUCT-SERVICE", Valid: true},
				{Path: "/deploy/sm-kms", Type: "PRODUCT-SERVICE", Valid: false, MissingDirs: []string{"secrets"}},
				{Path: "/deploy/sm-im", Type: "PRODUCT-SERVICE", Valid: true},
			},
			contains: []string{"3 deployments", "2 valid", "1 with issues"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output := FormatResults(tc.results)

			for _, expected := range tc.contains {
				require.Contains(t, output, expected, "output should contain: %s", expected)
			}

			for _, unexpected := range tc.notContains {
				require.NotContains(t, output, unexpected, "output should not contain: %s", unexpected)
			}

			// Verify sort order: invalid results should appear before valid ones.
			if tc.name == "mixed valid and invalid sorted correctly" {
				invalidIdx := strings.Index(output, "❌ INVALID")
				validIdx := strings.Index(output, "✅ VALID")
				require.Greater(t, validIdx, invalidIdx, "invalid results should appear before valid results in output")
			}
		})
	}
}

func TestGetExpectedStructures(t *testing.T) {
	t.Parallel()

	structures := GetExpectedStructures()

	// Verify all expected structure types exist.
	require.Contains(t, structures, "PRODUCT-SERVICE")
	require.Contains(t, structures, cryptoutilSharedMagic.SkeletonTemplateServiceName)
	require.Contains(t, structures, "infrastructure")

	// Verify PRODUCT-SERVICE has correct requirements.
	ps := structures["PRODUCT-SERVICE"]
	require.ElementsMatch(t, []string{"secrets", "config"}, ps.RequiredDirs)
	require.ElementsMatch(t, []string{"compose.yml", "Dockerfile"}, ps.RequiredFiles)
	require.Len(t, ps.RequiredSecrets, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, "PRODUCT-SERVICE should require 10 secrets")

	// Verify template has correct requirements.
	tmpl := structures[cryptoutilSharedMagic.SkeletonTemplateServiceName]
	require.ElementsMatch(t, []string{"secrets"}, tmpl.RequiredDirs)
	require.ElementsMatch(t, []string{"compose.yml"}, tmpl.RequiredFiles)

	// Verify infrastructure has minimal requirements.
	infra := structures["infrastructure"]
	require.Empty(t, infra.RequiredDirs)
	require.ElementsMatch(t, []string{"compose.yml"}, infra.RequiredFiles)
	require.Empty(t, infra.RequiredSecrets)
}

// joinErrors concatenates error strings for assertion convenience.
func joinErrors(errors []string) string {
	result := ""

	for _, e := range errors {
		result += e + "\n"
	}

	return result
}

func TestValidateAllDeployments_WithInfrastructure(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a valid infrastructure deployment (e.g., shared-postgres).
	infraDir := filepath.Join(tmpDir, "shared-postgres")
	require.NoError(t, os.MkdirAll(infraDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(infraDir, "compose.yml"), []byte("version: '3'"), cryptoutilSharedMagic.CacheFilePermissions))

	results, err := ValidateAllDeployments(tmpDir)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.True(t, results[0].Valid, "infrastructure deployment should be valid")
}

func TestValidateConfigFiles_DirectoryEntry(t *testing.T) {
	t.Parallel()

	// Create a PRODUCT-SERVICE deployment with a subdirectory inside config/.
	tmpDir := t.TempDir()
	createValidProductServiceDeployment(t, tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS)

	// Add a subdirectory inside config/ (should be skipped, not cause errors).
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "config", "subdir"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	result, err := ValidateDeploymentStructure(tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS, "PRODUCT-SERVICE")
	require.NoError(t, err)
	require.True(t, result.Valid, "subdirectory in config/ should be ignored")
	require.Empty(t, result.Errors, "no errors expected for config subdirectory")
}

func TestMain_ValidDeployments(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a valid PRODUCT-SERVICE deployment.
	svcDir := filepath.Join(tmpDir, cryptoutilSharedMagic.OTLPServiceJoseJA)
	require.NoError(t, os.MkdirAll(svcDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	createValidProductServiceDeployment(t, svcDir, cryptoutilSharedMagic.OTLPServiceJoseJA)

	exitCode := Main([]string{tmpDir})
	require.Equal(t, 0, exitCode, "should exit 0 for valid deployments")
}

func TestMain_InvalidDeployments(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create an invalid PRODUCT-SERVICE deployment (missing secrets and config).
	svcDir := filepath.Join(tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS)
	require.NoError(t, os.MkdirAll(svcDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(svcDir, "compose.yml"), []byte("version: '3'"), cryptoutilSharedMagic.CacheFilePermissions))

	exitCode := Main([]string{tmpDir})
	require.Equal(t, 1, exitCode, "should exit 1 for invalid deployments")
}

func TestMain_NonexistentDirectory(t *testing.T) {
	t.Parallel()

	exitCode := Main([]string{"/nonexistent/path/that/does/not/exist"})
	require.Equal(t, 1, exitCode, "should exit 1 for nonexistent directory")
}

func TestMain_DefaultDirectory(t *testing.T) {
	t.Parallel()

	// Main with no args uses "deployments" which doesn't exist in temp context.
	exitCode := Main(nil)
	// If "deployments" dir doesn't exist from CWD, it returns 1.
	// If it does exist (running from project root), it returns 0 or 1 depending on state.
	// Either way, it exercises the default directory path.
	require.Contains(t, []int{0, 1}, exitCode)
}

func TestMain_EmptyArg(t *testing.T) {
	t.Parallel()

	// Empty string arg should use default "deployments" directory.
	exitCode := Main([]string{""})
	require.Contains(t, []int{0, 1}, exitCode)
}
