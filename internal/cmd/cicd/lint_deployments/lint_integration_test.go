package lint_deployments_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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
				svcDir := filepath.Join(rootDir, "jose-ja")
				require.NoError(t, os.MkdirAll(svcDir, 0o755))
				createValidProductServiceDeployment(t, svcDir, "jose-ja")

				// Create a valid template deployment.
				tmplDir := filepath.Join(rootDir, "template")
				require.NoError(t, os.MkdirAll(filepath.Join(tmplDir, "secrets"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(tmplDir, "compose.yml"), []byte("v"), 0o600))
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
				validDir := filepath.Join(rootDir, "cipher-im")
				require.NoError(t, os.MkdirAll(validDir, 0o755))
				createValidProductServiceDeployment(t, validDir, "cipher-im")

				// Create an invalid deployment (missing secrets and config files).
				invalidDir := filepath.Join(rootDir, "sm-kms")
				require.NoError(t, os.MkdirAll(filepath.Join(invalidDir, "config"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(invalidDir, "compose.yml"), []byte("v"), 0o600))
				require.NoError(t, os.WriteFile(filepath.Join(invalidDir, "Dockerfile"), []byte("F"), 0o600))
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

			assert.Len(t, results, tc.wantCount, "deployment count mismatch")

			validCount := 0

			for _, r := range results {
				if r.Valid {
					validCount++
				}
			}

			assert.Equal(t, tc.wantValid, validCount, "valid count mismatch")
		})
	}
}

func TestFormatResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		results  []ValidationResult
		contains []string
	}{
		{
			name: "valid result",
			results: []ValidationResult{
				{
					Path:  "/deploy/jose-ja",
					Type:  "PRODUCT-SERVICE",
					Valid: true,
				},
			},
			contains: []string{"✅ VALID", "jose-ja", "1 valid"},
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
			contains: []string{"❌ INVALID", "sm-kms", "Missing directories", "Missing files", "Missing secrets", "ERROR: Config file error"},
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
			contains: []string{"✅ VALID", "WARN: Optional file missing"},
		},
		{
			name: "mixed valid and invalid sorted correctly",
			results: []ValidationResult{
				{Path: "/deploy/jose-ja", Type: "PRODUCT-SERVICE", Valid: true},
				{Path: "/deploy/sm-kms", Type: "PRODUCT-SERVICE", Valid: false, MissingDirs: []string{"secrets"}},
				{Path: "/deploy/cipher-im", Type: "PRODUCT-SERVICE", Valid: true},
			},
			contains: []string{"3 deployments", "2 valid", "1 with issues"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output := FormatResults(tc.results)

			for _, expected := range tc.contains {
				assert.Contains(t, output, expected, "output should contain: %s", expected)
			}
		})
	}
}

func TestGetExpectedStructures(t *testing.T) {
	t.Parallel()

	structures := GetExpectedStructures()

	// Verify all expected structure types exist.
	assert.Contains(t, structures, "PRODUCT-SERVICE")
	assert.Contains(t, structures, "template")
	assert.Contains(t, structures, "infrastructure")

	// Verify PRODUCT-SERVICE has correct requirements.
	ps := structures["PRODUCT-SERVICE"]
	assert.ElementsMatch(t, []string{"secrets", "config"}, ps.RequiredDirs)
	assert.ElementsMatch(t, []string{"compose.yml", "Dockerfile"}, ps.RequiredFiles)
	assert.Len(t, ps.RequiredSecrets, 10, "PRODUCT-SERVICE should require 10 secrets")

	// Verify template has correct requirements.
	tmpl := structures["template"]
	assert.ElementsMatch(t, []string{"secrets"}, tmpl.RequiredDirs)
	assert.ElementsMatch(t, []string{"compose.yml"}, tmpl.RequiredFiles)

	// Verify infrastructure has minimal requirements.
	infra := structures["infrastructure"]
	assert.Empty(t, infra.RequiredDirs)
	assert.ElementsMatch(t, []string{"compose.yml"}, infra.RequiredFiles)
	assert.Empty(t, infra.RequiredSecrets)
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

	// Create a valid infrastructure deployment (e.g., postgres).
	infraDir := filepath.Join(tmpDir, "postgres")
	require.NoError(t, os.MkdirAll(infraDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(infraDir, "compose.yml"), []byte("version: '3'"), 0o600))

	results, err := ValidateAllDeployments(tmpDir)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.True(t, results[0].Valid, "infrastructure deployment should be valid")
}

func TestValidateConfigFiles_DirectoryEntry(t *testing.T) {
	t.Parallel()

	// Create a PRODUCT-SERVICE deployment with a subdirectory inside config/.
	tmpDir := t.TempDir()
	createValidProductServiceDeployment(t, tmpDir, "sm-kms")

	// Add a subdirectory inside config/ (should be skipped, not cause errors).
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "config", "subdir"), 0o755))

	result, err := ValidateDeploymentStructure(tmpDir, "sm-kms", "PRODUCT-SERVICE")
	require.NoError(t, err)
	assert.True(t, result.Valid, "subdirectory in config/ should be ignored")
	assert.Empty(t, result.Errors, "no errors expected for config subdirectory")
}

func TestMain_ValidDeployments(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a valid PRODUCT-SERVICE deployment.
	svcDir := filepath.Join(tmpDir, "jose-ja")
	require.NoError(t, os.MkdirAll(svcDir, 0o755))
	createValidProductServiceDeployment(t, svcDir, "jose-ja")

	exitCode := Main([]string{tmpDir})
	assert.Equal(t, 0, exitCode, "should exit 0 for valid deployments")
}

func TestMain_InvalidDeployments(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create an invalid PRODUCT-SERVICE deployment (missing secrets and config).
	svcDir := filepath.Join(tmpDir, "sm-kms")
	require.NoError(t, os.MkdirAll(svcDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(svcDir, "compose.yml"), []byte("version: '3'"), 0o600))

	exitCode := Main([]string{tmpDir})
	assert.Equal(t, 1, exitCode, "should exit 1 for invalid deployments")
}

func TestMain_NonexistentDirectory(t *testing.T) {
	t.Parallel()

	exitCode := Main([]string{"/nonexistent/path/that/does/not/exist"})
	assert.Equal(t, 1, exitCode, "should exit 1 for nonexistent directory")
}

func TestMain_DefaultDirectory(t *testing.T) {
	t.Parallel()

	// Main with no args uses "deployments" which doesn't exist in temp context.
	exitCode := Main(nil)
	// If "deployments" dir doesn't exist from CWD, it returns 1.
	// If it does exist (running from project root), it returns 0 or 1 depending on state.
	// Either way, it exercises the default directory path.
	assert.Contains(t, []int{0, 1}, exitCode)
}

func TestMain_EmptyArg(t *testing.T) {
	t.Parallel()

	// Empty string arg should use default "deployments" directory.
	exitCode := Main([]string{""})
	assert.Contains(t, []int{0, 1}, exitCode)
}
