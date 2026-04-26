package lint_deployments

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// createValidTemplateDir sets up a minimal valid template directory structure.
func createValidTemplateDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	// Create required directories.
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "config"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "secrets"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	// Compose files with placeholders.
	serviceContent := "name: PRODUCT-SERVICE\nservices:\n  PRODUCT-SERVICE-sqlite:\n    ports:\n      - \"XXXX:8080\"\n"
	productContent := "name: PRODUCT\nservices:\n  PRODUCT-sqlite:\n    ports:\n      - \"18000:8080\"\n"
	suiteContent := "name: cryptoutil\nservices:\n  sm-kms-sqlite:\n    ports:\n      - \"28000:8080\"\n"
	baseContent := "name: PRODUCT-SERVICE\nservices:\n  PRODUCT-SERVICE-sqlite:\n    image: local\n"

	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose-cryptoutil-PRODUCT-SERVICE.yml"), []byte(serviceContent), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose-cryptoutil-PRODUCT.yml"), []byte(productContent), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose-cryptoutil.yml"), []byte(suiteContent), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(baseContent), cryptoutilSharedMagic.CacheFilePermissions))

	// Config files.
	for _, f := range requiredTemplateConfigFiles {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "config", f), []byte("# template config\n"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	// Secret files.
	for _, f := range requiredTemplateSecretFiles {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "secrets", f), []byte("secret-value"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	return dir
}

func TestValidateTemplatePattern_ValidCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setupFn func(t *testing.T) string
	}{
		{
			name: "valid template",
			setupFn: func(t *testing.T) string {
				t.Helper()

				return createValidTemplateDir(t)
			},
		},
		{
			name: "non-YAML config ignored",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := createValidTemplateDir(t)
				require.NoError(t, os.WriteFile(filepath.Join(dir, "config", "README.md"), []byte("# readme\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: "config subdir ignored",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := createValidTemplateDir(t)
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "config", "subdir"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

				return dir
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := ValidateTemplatePattern(tc.setupFn(t))
			require.NoError(t, err)
			require.NotNil(t, result)
			require.True(t, result.Valid)
			require.Empty(t, result.Errors)
		})
	}
}

func TestValidateTemplatePattern_Violations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setupFn      func(t *testing.T) string
		wantContains []string
	}{
		{
			name: "path not found",
			setupFn: func(t *testing.T) string {
				t.Helper()

				return filepath.Join(t.TempDir(), "nonexistent")
			},
			wantContains: []string{"does not exist"},
		},
		{
			name: "path is file",
			setupFn: func(t *testing.T) string {
				t.Helper()

				f := filepath.Join(t.TempDir(), "file.txt")
				require.NoError(t, os.WriteFile(f, []byte("data"), cryptoutilSharedMagic.CacheFilePermissions))

				return f
			},
			wantContains: []string{"not a directory"},
		},
		{
			name: "missing compose files",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := createValidTemplateDir(t)
				require.NoError(t, os.Remove(filepath.Join(dir, "compose.yml")))
				require.NoError(t, os.Remove(filepath.Join(dir, "compose-cryptoutil-PRODUCT-SERVICE.yml")))

				return dir
			},
			wantContains: []string{"compose.yml", "compose-cryptoutil-PRODUCT-SERVICE.yml"},
		},
		{
			name: "missing config dir",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := createValidTemplateDir(t)
				require.NoError(t, os.RemoveAll(filepath.Join(dir, "config")))

				return dir
			},
			wantContains: []string{"config/ directory"},
		},
		{
			name: "missing config file",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := createValidTemplateDir(t)
				require.NoError(t, os.Remove(filepath.Join(dir, "config", "template-app-common.yml")))

				return dir
			},
			wantContains: []string{"template-app-common.yml"},
		},
		{
			name: "missing secrets dir",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := createValidTemplateDir(t)
				require.NoError(t, os.RemoveAll(filepath.Join(dir, "secrets")))

				return dir
			},
			wantContains: []string{"secrets/ directory"},
		},
		{
			name: "missing secret file",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := createValidTemplateDir(t)
				require.NoError(t, os.Remove(filepath.Join(dir, "secrets", "unseal_3of5.secret")))

				return dir
			},
			wantContains: []string{"unseal_3of5.secret"},
		},
		{
			name: "missing service placeholder",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := createValidTemplateDir(t)
				content := "name: my-service\nservices:\n  PRODUCT-SERVICE-sqlite:\n    ports:\n      - \"8080:8080\"\n"
				require.NoError(t, os.WriteFile(filepath.Join(dir, "compose-cryptoutil-PRODUCT-SERVICE.yml"), []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
			wantContains: []string{"XXXX"},
		},
		{
			name: "missing product placeholder",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := createValidTemplateDir(t)
				content := "name: my-app\nservices:\n  my-sqlite:\n    ports:\n      - \"18000:8080\"\n"
				require.NoError(t, os.WriteFile(filepath.Join(dir, "compose-cryptoutil-PRODUCT.yml"), []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
			wantContains: []string{"PRODUCT"},
		},
		{
			name: "missing product compose for placeholders",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := createValidTemplateDir(t)
				require.NoError(t, os.Remove(filepath.Join(dir, "compose-cryptoutil-PRODUCT.yml")))

				return dir
			},
			wantContains: []string{"compose-cryptoutil-PRODUCT.yml"},
		},
		{
			name: "missing service compose for placeholders",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := createValidTemplateDir(t)
				require.NoError(t, os.Remove(filepath.Join(dir, "compose-cryptoutil-PRODUCT-SERVICE.yml")))

				return dir
			},
			wantContains: []string{"compose-cryptoutil-PRODUCT-SERVICE.yml"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := ValidateTemplatePattern(tc.setupFn(t))
			require.NoError(t, err)
			require.NotNil(t, result)
			require.False(t, result.Valid)

			for _, want := range tc.wantContains {
				require.True(t, containsSubstring(result.Errors, want))
			}
		})
	}
}

func TestValidateTemplatePattern_ConfigNonStandardNaming(t *testing.T) {
	t.Parallel()

	dir := createValidTemplateDir(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config", "custom-config.yml"), []byte("# custom\n"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateTemplatePattern(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Valid)
	require.True(t, containsSubstring(result.Warnings, "custom-config.yml"))
}

func TestFormatTemplatePatternResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *TemplatePatternResult
		contains []string
	}{
		{
			name:     "passing",
			result:   &TemplatePatternResult{Path: "/template", Valid: true},
			contains: []string{cryptoutilSharedMagic.TestStatusPass, "/template"},
		},
		{
			name: "failing with errors and warnings",
			result: &TemplatePatternResult{
				Path: "/template", Valid: false,
				Errors:   []string{"missing compose.yml"},
				Warnings: []string{"non-standard name"},
			},
			contains: []string{cryptoutilSharedMagic.TestStatusFail, "missing compose.yml", "non-standard name"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output := FormatTemplatePatternResult(tc.result)
			for _, s := range tc.contains {
				require.Contains(t, output, s)
			}
		})
	}
}

func TestValidateTemplatePattern_RealTemplate(t *testing.T) {
	t.Parallel()

	templatePath := filepath.Join(".", "..", "..", "..", "..", "deployments", cryptoutilSharedMagic.SkeletonTemplateServiceName)

	info, err := os.Stat(templatePath)
	if err != nil || !info.IsDir() {
		t.Skip("Real template directory not found - skipping integration test")
	}

	result, err := ValidateTemplatePattern(templatePath)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Valid, "Real template should pass validation. Errors: %v", result.Errors)
}
