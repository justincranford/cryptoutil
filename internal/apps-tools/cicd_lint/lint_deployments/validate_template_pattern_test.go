package lint_deployments

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const templateServiceComposeFile = "compose-cryptoutil-PRODUCT-SERVICE.yml"

// createValidTemplateDir sets up a minimal valid template directory structure.
func createValidTemplateDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	// Create required directories.
	require.NoError(t, os.MkdirAll(filepath.Join(dir, fileTypeConfig), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, validatorNameSecrets), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	// Compose files with placeholders.
	serviceContent := "name: PRODUCT-SERVICE\nservices:\n  PRODUCT-SERVICE-sqlite:\n    ports:\n      - \"XXXX:8080\"\n"
	productContent := "name: PRODUCT\nservices:\n  PRODUCT-sqlite:\n    ports:\n      - \"18000:8080\"\n"
	suiteContent := "name: cryptoutil\nservices:\n  sm-kms-sqlite:\n    ports:\n      - \"28000:8080\"\n"
	baseContent := "name: PRODUCT-SERVICE\nservices:\n  PRODUCT-SERVICE-sqlite:\n    image: local\n"

	require.NoError(t, os.WriteFile(filepath.Join(dir, templateServiceComposeFile), []byte(serviceContent), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose-cryptoutil-PRODUCT.yml"), []byte(productContent), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose-cryptoutil.yml"), []byte(suiteContent), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(dir, cryptoutilSharedMagic.COMPOSE_YML), []byte(baseContent), cryptoutilSharedMagic.CacheFilePermissions))

	// Config files.
	for _, f := range requiredTemplateConfigFiles {
		require.NoError(t, os.WriteFile(filepath.Join(dir, fileTypeConfig, f), []byte("# template config\n"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	// Secret files.
	for _, f := range requiredTemplateSecretFiles {
		require.NoError(t, os.WriteFile(filepath.Join(dir, validatorNameSecrets, f), []byte("secret-value"), cryptoutilSharedMagic.CacheFilePermissions))
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
				require.NoError(t, os.WriteFile(filepath.Join(dir, fileTypeConfig, valueReadmeMd), []byte("# readme\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: "config subdir ignored",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := createValidTemplateDir(t)
				require.NoError(t, os.MkdirAll(filepath.Join(dir, fileTypeConfig, "subdir"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

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
			name: testPathNotFound,
			setupFn: func(t *testing.T) string {
				t.Helper()

				return filepath.Join(t.TempDir(), "nonexistent")
			},
			wantContains: []string{testDoesNotExist},
		},
		{
			name: testPathIsFile,
			setupFn: func(t *testing.T) string {
				t.Helper()

				f := filepath.Join(t.TempDir(), "file.txt")
				require.NoError(t, os.WriteFile(f, []byte("data"), cryptoutilSharedMagic.CacheFilePermissions))

				return f
			},
			wantContains: []string{testNotADirectory},
		},
		{
			name: "missing compose files",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := createValidTemplateDir(t)
				require.NoError(t, os.Remove(filepath.Join(dir, cryptoutilSharedMagic.COMPOSE_YML)))
				require.NoError(t, os.Remove(filepath.Join(dir, templateServiceComposeFile)))

				return dir
			},
			wantContains: []string{cryptoutilSharedMagic.COMPOSE_YML, templateServiceComposeFile},
		},
		{
			name: "missing config dir",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := createValidTemplateDir(t)
				require.NoError(t, os.RemoveAll(filepath.Join(dir, fileTypeConfig)))

				return dir
			},
			wantContains: []string{"config/ directory"},
		},
		{
			name: testMissingConfigFile,
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := createValidTemplateDir(t)
				require.NoError(t, os.Remove(filepath.Join(dir, fileTypeConfig, "template-app-common.yml")))

				return dir
			},
			wantContains: []string{"template-app-common.yml"},
		},
		{
			name: "missing secrets dir",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := createValidTemplateDir(t)
				require.NoError(t, os.RemoveAll(filepath.Join(dir, validatorNameSecrets)))

				return dir
			},
			wantContains: []string{"secrets/ directory"},
		},
		{
			name: "missing secret file",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := createValidTemplateDir(t)
				require.NoError(t, os.Remove(filepath.Join(dir, validatorNameSecrets, valueUnseal3of5Secret)))

				return dir
			},
			wantContains: []string{valueUnseal3of5Secret},
		},
		{
			name: "missing service placeholder",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := createValidTemplateDir(t)
				content := "name: my-service\nservices:\n  PRODUCT-SERVICE-sqlite:\n    ports:\n      - \"8080:8080\"\n"
				require.NoError(t, os.WriteFile(filepath.Join(dir, templateServiceComposeFile), []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

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
				require.NoError(t, os.Remove(filepath.Join(dir, templateServiceComposeFile)))

				return dir
			},
			wantContains: []string{templateServiceComposeFile},
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
	require.NoError(t, os.WriteFile(filepath.Join(dir, fileTypeConfig, "custom-config.yml"), []byte("# custom\n"), cryptoutilSharedMagic.CacheFilePermissions))

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
			name:     testPassing,
			result:   &TemplatePatternResult{Path: testTemplate, Valid: true},
			contains: []string{cryptoutilSharedMagic.TestStatusPass, testTemplate},
		},
		{
			name: testFailingWithErrorsAndWarnings,
			result: &TemplatePatternResult{
				Path: testTemplate, Valid: false,
				Errors:   []string{testMissingComposeYml},
				Warnings: []string{testNonStandardName},
			},
			contains: []string{cryptoutilSharedMagic.TestStatusFail, testMissingComposeYml, testNonStandardName},
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
