package lint_deployments

import (
"os"
"path/filepath"
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
)

// createValidTemplateDir sets up a minimal valid template directory structure.
func createValidTemplateDir(t *testing.T) string {
t.Helper()

dir := t.TempDir()

// Create required directories.
require.NoError(t, os.MkdirAll(filepath.Join(dir, "config"), 0o755))
require.NoError(t, os.MkdirAll(filepath.Join(dir, "secrets"), 0o755))

// Compose files with placeholders.
serviceContent := "name: PRODUCT-SERVICE\nservices:\n  PRODUCT-SERVICE-sqlite:\n    ports:\n      - \"XXXX:8080\"\n"
productContent := "name: PRODUCT\nservices:\n  PRODUCT-sqlite:\n    ports:\n      - \"18000:8080\"\n"
suiteContent := "name: cryptoutil\nservices:\n  sm-kms-sqlite:\n    ports:\n      - \"28000:8080\"\n"
baseContent := "name: PRODUCT-SERVICE\nservices:\n  PRODUCT-SERVICE-sqlite:\n    image: local\n"

require.NoError(t, os.WriteFile(filepath.Join(dir, "compose-cryptoutil-PRODUCT-SERVICE.yml"), []byte(serviceContent), 0o600))
require.NoError(t, os.WriteFile(filepath.Join(dir, "compose-cryptoutil-PRODUCT.yml"), []byte(productContent), 0o600))
require.NoError(t, os.WriteFile(filepath.Join(dir, "compose-cryptoutil.yml"), []byte(suiteContent), 0o600))
require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(baseContent), 0o600))

// Config files.
for _, f := range requiredTemplateConfigFiles {
require.NoError(t, os.WriteFile(filepath.Join(dir, "config", f), []byte("# template config\n"), 0o600))
}

// Secret files.
for _, f := range requiredTemplateSecretFiles {
require.NoError(t, os.WriteFile(filepath.Join(dir, "secrets", f), []byte("secret-value"), 0o600))
}

return dir
}

func TestValidateTemplatePattern_ValidTemplate(t *testing.T) {
t.Parallel()

dir := createValidTemplateDir(t)
result, err := ValidateTemplatePattern(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid)
assert.Empty(t, result.Errors)
}

func TestValidateTemplatePattern_PathNotFound(t *testing.T) {
t.Parallel()

result, err := ValidateTemplatePattern(filepath.Join(t.TempDir(), "nonexistent"))
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.True(t, containsSubstring(result.Errors, "does not exist"))
}

func TestValidateTemplatePattern_PathIsFile(t *testing.T) {
t.Parallel()

f := filepath.Join(t.TempDir(), "file.txt")
require.NoError(t, os.WriteFile(f, []byte("data"), 0o600))

result, err := ValidateTemplatePattern(f)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.True(t, containsSubstring(result.Errors, "not a directory"))
}

func TestValidateTemplatePattern_MissingComposeFiles(t *testing.T) {
t.Parallel()

dir := createValidTemplateDir(t)
require.NoError(t, os.Remove(filepath.Join(dir, "compose.yml")))
require.NoError(t, os.Remove(filepath.Join(dir, "compose-cryptoutil-PRODUCT-SERVICE.yml")))

result, err := ValidateTemplatePattern(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.True(t, containsSubstring(result.Errors, "compose.yml"))
assert.True(t, containsSubstring(result.Errors, "compose-cryptoutil-PRODUCT-SERVICE.yml"))
}

func TestValidateTemplatePattern_MissingConfigDir(t *testing.T) {
t.Parallel()

dir := createValidTemplateDir(t)
require.NoError(t, os.RemoveAll(filepath.Join(dir, "config")))

result, err := ValidateTemplatePattern(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.True(t, containsSubstring(result.Errors, "config/ directory"))
}

func TestValidateTemplatePattern_MissingConfigFile(t *testing.T) {
t.Parallel()

dir := createValidTemplateDir(t)
require.NoError(t, os.Remove(filepath.Join(dir, "config", "template-app-common.yml")))

result, err := ValidateTemplatePattern(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.True(t, containsSubstring(result.Errors, "template-app-common.yml"))
}

func TestValidateTemplatePattern_MissingSecretsDir(t *testing.T) {
t.Parallel()

dir := createValidTemplateDir(t)
require.NoError(t, os.RemoveAll(filepath.Join(dir, "secrets")))

result, err := ValidateTemplatePattern(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.True(t, containsSubstring(result.Errors, "secrets/ directory"))
}

func TestValidateTemplatePattern_MissingSecretFile(t *testing.T) {
t.Parallel()

dir := createValidTemplateDir(t)
require.NoError(t, os.Remove(filepath.Join(dir, "secrets", "unseal_3of5.secret")))

result, err := ValidateTemplatePattern(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.True(t, containsSubstring(result.Errors, "unseal_3of5.secret"))
}

func TestValidateTemplatePattern_MissingServicePlaceholder(t *testing.T) {
t.Parallel()

dir := createValidTemplateDir(t)
// Overwrite service compose template without XXXX placeholder.
content := "name: my-service\nservices:\n  PRODUCT-SERVICE-sqlite:\n    ports:\n      - \"8080:8080\"\n"
require.NoError(t, os.WriteFile(filepath.Join(dir, "compose-cryptoutil-PRODUCT-SERVICE.yml"), []byte(content), 0o600))

result, err := ValidateTemplatePattern(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.True(t, containsSubstring(result.Errors, "XXXX"))
}

func TestValidateTemplatePattern_MissingProductPlaceholder(t *testing.T) {
t.Parallel()

dir := createValidTemplateDir(t)
// Overwrite product compose template without PRODUCT placeholder.
content := "name: my-app\nservices:\n  my-sqlite:\n    ports:\n      - \"18000:8080\"\n"
require.NoError(t, os.WriteFile(filepath.Join(dir, "compose-cryptoutil-PRODUCT.yml"), []byte(content), 0o600))

result, err := ValidateTemplatePattern(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.True(t, containsSubstring(result.Errors, "PRODUCT"))
}

func TestValidateTemplatePattern_ConfigNonStandardNaming(t *testing.T) {
t.Parallel()

dir := createValidTemplateDir(t)
// Add a config file that doesn't follow template-app-*.yml naming.
require.NoError(t, os.WriteFile(filepath.Join(dir, "config", "custom-config.yml"), []byte("# custom\n"), 0o600))

result, err := ValidateTemplatePattern(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid) // Warnings don't affect validity.
assert.True(t, containsSubstring(result.Warnings, "custom-config.yml"))
}

func TestValidateTemplatePattern_NonYAMLConfigIgnored(t *testing.T) {
t.Parallel()

dir := createValidTemplateDir(t)
// Add a non-YAML file in config - should be ignored.
require.NoError(t, os.WriteFile(filepath.Join(dir, "config", "README.md"), []byte("# readme\n"), 0o600))

result, err := ValidateTemplatePattern(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid)
assert.Empty(t, result.Warnings)
}

func TestValidateTemplatePattern_ConfigSubdirIgnored(t *testing.T) {
t.Parallel()

dir := createValidTemplateDir(t)
// Add a subdirectory inside config - should be ignored.
require.NoError(t, os.MkdirAll(filepath.Join(dir, "config", "subdir"), 0o755))

result, err := ValidateTemplatePattern(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid)
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
contains: []string{"PASS", "/template"},
},
{
name: "failing with errors and warnings",
result: &TemplatePatternResult{
Path: "/template", Valid: false,
Errors:   []string{"missing compose.yml"},
Warnings: []string{"non-standard name"},
},
contains: []string{"FAIL", "missing compose.yml", "non-standard name"},
},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

output := FormatTemplatePatternResult(tc.result)
for _, s := range tc.contains {
assert.Contains(t, output, s)
}
})
}
}

func TestValidateTemplatePattern_RealTemplate(t *testing.T) {
t.Parallel()

// Validate the actual template directory in the repository.
templatePath := filepath.Join(".", "..", "..", "..", "..", "deployments", "template")

info, err := os.Stat(templatePath)
if err != nil || !info.IsDir() {
t.Skip("Real template directory not found - skipping integration test")
}

result, err := ValidateTemplatePattern(templatePath)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid, "Real template should pass validation. Errors: %v", result.Errors)
}

func TestValidateTemplatePattern_MissingProductComposeForPlaceholders(t *testing.T) {
t.Parallel()

dir := createValidTemplateDir(t)
// Remove only the product compose file to hit the ReadFile error path.
require.NoError(t, os.Remove(filepath.Join(dir, "compose-cryptoutil-PRODUCT.yml")))

result, err := ValidateTemplatePattern(dir)
require.NoError(t, err)
require.NotNil(t, result)
// Missing file reported by validateRequiredTemplateFiles, placeholder check skipped.
assert.False(t, result.Valid)
assert.True(t, containsSubstring(result.Errors, "compose-cryptoutil-PRODUCT.yml"))
}

func TestValidateTemplatePattern_MissingServiceComposeForPlaceholders(t *testing.T) {
t.Parallel()

dir := createValidTemplateDir(t)
// Remove only the service compose file.
require.NoError(t, os.Remove(filepath.Join(dir, "compose-cryptoutil-PRODUCT-SERVICE.yml")))

result, err := ValidateTemplatePattern(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.True(t, containsSubstring(result.Errors, "compose-cryptoutil-PRODUCT-SERVICE.yml"))
}
