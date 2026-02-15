package lint_deployments

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateDeploymentStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		structType         string
		setupFunc          func(t *testing.T, baseDir string)
		wantValid          bool
		wantMissingDirs    []string
		wantMissingFiles   []string
		wantMissingSecrets []string
	}{
		{
			name:       "valid PRODUCT-SERVICE deployment",
			structType: "PRODUCT-SERVICE",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()
				// Create required directories
				require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "secrets"), 0o755))
				require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "config"), 0o755))

				// Create required files
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "compose.yml"), []byte("version: '3'"), 0o600))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "Dockerfile"), []byte("FROM alpine"), 0o600))

				// Create required secrets
				secrets := []string{
					"unseal_1of5.secret", "unseal_2of5.secret", "unseal_3of5.secret",
					"unseal_4of5.secret", "unseal_5of5.secret", "hash_pepper_v3.secret",
					"postgres_url.secret", "postgres_username.secret",
					"postgres_password.secret", "postgres_database.secret",
				}
				for _, secret := range secrets {
					require.NoError(t, os.WriteFile(filepath.Join(baseDir, "secrets", secret), []byte("secret"), 0o600))
				}
			},
			wantValid: true,
		},
		{
			name:       "missing secrets directory",
			structType: "PRODUCT-SERVICE",
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
			name:       "missing config directory",
			structType: "PRODUCT-SERVICE",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "secrets"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "compose.yml"), []byte("version: '3'"), 0o600))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "Dockerfile"), []byte("FROM alpine"), 0o600))
			},
			wantValid:       false,
			wantMissingDirs: []string{"config"},
			wantMissingSecrets: []string{
				"unseal_1of5.secret", "unseal_2of5.secret", "unseal_3of5.secret",
				"unseal_4of5.secret", "unseal_5of5.secret", "hash_pepper_v3.secret",
				"postgres_url.secret", "postgres_username.secret",
				"postgres_password.secret", "postgres_database.secret",
			},
		},
		{
			name:       "missing compose.yml",
			structType: "PRODUCT-SERVICE",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "secrets"), 0o755))
				require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "config"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "Dockerfile"), []byte("FROM alpine"), 0o600))
			},
			wantValid:        false,
			wantMissingFiles: []string{"compose.yml"},
			wantMissingSecrets: []string{
				"unseal_1of5.secret", "unseal_2of5.secret", "unseal_3of5.secret",
				"unseal_4of5.secret", "unseal_5of5.secret", "hash_pepper_v3.secret",
				"postgres_url.secret", "postgres_username.secret",
				"postgres_password.secret", "postgres_database.secret",
			},
		},
		{
			name:       "missing Dockerfile",
			structType: "PRODUCT-SERVICE",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "secrets"), 0o755))
				require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "config"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "compose.yml"), []byte("version: '3'"), 0o600))
			},
			wantValid:        false,
			wantMissingFiles: []string{"Dockerfile"},
			wantMissingSecrets: []string{
				"unseal_1of5.secret", "unseal_2of5.secret", "unseal_3of5.secret",
				"unseal_4of5.secret", "unseal_5of5.secret", "hash_pepper_v3.secret",
				"postgres_url.secret", "postgres_username.secret",
				"postgres_password.secret", "postgres_database.secret",
			},
		},
		{
			name:       "missing some secrets",
			structType: "PRODUCT-SERVICE",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "secrets"), 0o755))
				require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "config"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "compose.yml"), []byte("version: '3'"), 0o600))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "Dockerfile"), []byte("FROM alpine"), 0o600))

				// Create only some secrets
				for i := 1; i <= 5; i++ {
					require.NoError(t, os.WriteFile(
						filepath.Join(baseDir, "secrets", "unseal_"+string(rune('0'+i))+"of5.secret"),
						[]byte("secret"), 0o600))
				}
				// Missing: hash_pepper, postgres_*
			},
			wantValid: false,
			wantMissingSecrets: []string{
				"hash_pepper_v3.secret", "postgres_url.secret",
				"postgres_username.secret", "postgres_password.secret", "postgres_database.secret",
			},
		},
		{
			name:       "valid template deployment",
			structType: "template",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "secrets"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "compose.yml"), []byte("version: '3'"), 0o600))

				secrets := []string{
					"unseal_1of5.secret", "unseal_2of5.secret", "unseal_3of5.secret",
					"unseal_4of5.secret", "unseal_5of5.secret", "hash_pepper_v3.secret",
					"postgres_url.secret", "postgres_username.secret",
					"postgres_password.secret", "postgres_database.secret",
				}
				for _, secret := range secrets {
					require.NoError(t, os.WriteFile(filepath.Join(baseDir, "secrets", secret), []byte("secret"), 0o600))
				}
			},
			wantValid: true,
		},
		{
			name:       "valid infrastructure deployment",
			structType: "infrastructure",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "compose.yml"), []byte("version: '3'"), 0o600))
			},
			wantValid: true,
		},
		{
			name:       "infrastructure with optional files",
			structType: "infrastructure",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "compose.yml"), []byte("version: '3'"), 0o600))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "init-db.sql"), []byte("CREATE TABLE..."), 0o600))
				require.NoError(t, os.WriteFile(filepath.Join(baseDir, "README.md"), []byte("# Infrastructure"), 0o600))
			},
			wantValid: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create temporary directory
			tmpDir := t.TempDir()

			// Setup test structure
			tc.setupFunc(t, tmpDir)

			// Validate
			result, err := ValidateDeploymentStructure(tmpDir, tc.name, tc.structType)
			require.NoError(t, err)

			// Assert results
			assert.Equal(t, tc.wantValid, result.Valid, "validity mismatch")
			assert.ElementsMatch(t, tc.wantMissingDirs, result.MissingDirs, "missing dirs mismatch")
			assert.ElementsMatch(t, tc.wantMissingFiles, result.MissingFiles, "missing files mismatch")
			assert.ElementsMatch(t, tc.wantMissingSecrets, result.MissingSecrets, "missing secrets mismatch")
		})
	}
}

func TestValidateAllDeployments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, baseDir string)
		wantValidCount int
		wantTotalCount int
	}{
		{
			name: "all valid deployments",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()

				// Create valid jose-ja deployment
				createValidProductService(t, filepath.Join(baseDir, "jose-ja"))

				// Create valid template deployment
				createValidTemplate(t, filepath.Join(baseDir, "template"))

				// Create valid infrastructure
				createValidInfrastructure(t, filepath.Join(baseDir, "postgres"))
			},
			wantValidCount: 3,
			wantTotalCount: 3,
		},
		{
			name: "mixed valid and invalid",
			setupFunc: func(t *testing.T, baseDir string) {
				t.Helper()

				// Valid jose-ja
				createValidProductService(t, filepath.Join(baseDir, "jose-ja"))

				// Invalid cipher-im (missing Dockerfile)
				invalidDir := filepath.Join(baseDir, "cipher-im")
				require.NoError(t, os.MkdirAll(filepath.Join(invalidDir, "secrets"), 0o755))
				require.NoError(t, os.MkdirAll(filepath.Join(invalidDir, "config"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(invalidDir, "compose.yml"), []byte("version: '3'"), 0o600))

				// Valid template
				createValidTemplate(t, filepath.Join(baseDir, "template"))
			},
			wantValidCount: 2,
			wantTotalCount: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			tc.setupFunc(t, tmpDir)

			results, err := ValidateAllDeployments(tmpDir)
			require.NoError(t, err)

			assert.Equal(t, tc.wantTotalCount, len(results), "total count mismatch")

			validCount := 0

			for _, r := range results {
				if r.Valid {
					validCount++
				}
			}

			assert.Equal(t, tc.wantValidCount, validCount, "valid count mismatch")
		})
	}
}

func TestFormatResults(t *testing.T) {
	t.Parallel()

	results := []ValidationResult{
		{
			Path:         "deployments/jose-ja",
			Type:         "PRODUCT-SERVICE",
			Valid:        true,
			MissingDirs:  []string{},
			MissingFiles: []string{},
		},
		{
			Path:         "deployments/cipher-im",
			Type:         "PRODUCT-SERVICE",
			Valid:        false,
			MissingDirs:  []string{"config"},
			MissingFiles: []string{"Dockerfile"},
		},
	}

	output := FormatResults(results)

	assert.Contains(t, output, "Validated 2 deployments")
	assert.Contains(t, output, "1 valid")
	assert.Contains(t, output, "1 with issues")
	assert.Contains(t, output, "✅ VALID")
	assert.Contains(t, output, "❌ INVALID")
	assert.Contains(t, output, "Missing directories: config")
	assert.Contains(t, output, "Missing files: Dockerfile")
}

// Helper functions for test setup.

func createValidProductService(t *testing.T, baseDir string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "secrets"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "config"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(baseDir, "compose.yml"), []byte("version: '3'"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(baseDir, "Dockerfile"), []byte("FROM alpine"), 0o600))

	secrets := []string{
		"unseal_1of5.secret", "unseal_2of5.secret", "unseal_3of5.secret",
		"unseal_4of5.secret", "unseal_5of5.secret", "hash_pepper_v3.secret",
		"postgres_url.secret", "postgres_username.secret",
		"postgres_password.secret", "postgres_database.secret",
	}
	for _, secret := range secrets {
		require.NoError(t, os.WriteFile(filepath.Join(baseDir, "secrets", secret), []byte("secret"), 0o600))
	}
}

func createValidTemplate(t *testing.T, baseDir string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "secrets"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(baseDir, "compose.yml"), []byte("version: '3'"), 0o600))

	secrets := []string{
		"unseal_1of5.secret", "unseal_2of5.secret", "unseal_3of5.secret",
		"unseal_4of5.secret", "unseal_5of5.secret", "hash_pepper_v3.secret",
		"postgres_url.secret", "postgres_username.secret",
		"postgres_password.secret", "postgres_database.secret",
	}
	for _, secret := range secrets {
		require.NoError(t, os.WriteFile(filepath.Join(baseDir, "secrets", secret), []byte("secret"), 0o600))
	}
}

func createValidInfrastructure(t *testing.T, baseDir string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(baseDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(baseDir, "compose.yml"), []byte("version: '3'"), 0o600))
}
