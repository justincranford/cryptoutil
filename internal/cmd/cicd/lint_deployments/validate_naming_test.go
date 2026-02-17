package lint_deployments

import (
"os"
"path/filepath"
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
)

func TestValidateNaming_Simple(t *testing.T) {
t.Parallel()

dir := t.TempDir()
require.NoError(t, os.MkdirAll(filepath.Join(dir, "service-one"), 0o755))

result, err := ValidateNaming(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid)
}

func TestValidateNaming_InvalidPascalCase(t *testing.T) {
t.Parallel()

dir := t.TempDir()
require.NoError(t, os.MkdirAll(filepath.Join(dir, "ServiceOne"), 0o755))

result, err := ValidateNaming(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.True(t, len(result.Errors) > 0)
assert.Contains(t, result.Errors[0], "ServiceOne")
assert.Contains(t, result.Errors[0], "kebab-case")
}

func TestValidateNaming_InvalidSnakeCase(t *testing.T) {
t.Parallel()

dir := t.TempDir()
require.NoError(t, os.MkdirAll(filepath.Join(dir, "service_one"), 0o755))

result, err := ValidateNaming(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
}

func TestValidateNaming_InvalidComposeServiceNames(t *testing.T) {
t.Parallel()

dir := t.TempDir()
content := `services:
  my_service:
    image: nginx
  AnotherService:
    image: postgres
`
require.NoError(t, os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(content), 0o644))

result, err := ValidateNaming(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.True(t, len(result.Errors) >= 2)
}

func TestIsKebabCase(t *testing.T) {
t.Parallel()

assert.True(t, isKebabCase("hello"))
assert.True(t, isKebabCase("hello-world"))
assert.True(t, isKebabCase("service-123"))
assert.False(t, isKebabCase("HelloWorld"))
assert.False(t, isKebabCase("hello_world"))
assert.False(t, isKebabCase("-hello"))
assert.False(t, isKebabCase("hello-"))
assert.False(t, isKebabCase("hello--world"))
}

func TestToKebabCase(t *testing.T) {
t.Parallel()

assert.Equal(t, "helloworld", toKebabCase("HelloWorld"))
assert.Equal(t, "hello-world", toKebabCase("hello_world"))
assert.Equal(t, "hello", toKebabCase("HELLO"))
assert.Equal(t, "myconfig.yml", toKebabCase("MyConfig.yml"))
assert.Equal(t, "my-config.yaml", toKebabCase("My_Config.yaml"))
}

func TestIsYAMLFile(t *testing.T) {
t.Parallel()

assert.True(t, isYAMLFile("config.yml"))
assert.True(t, isYAMLFile("config.yaml"))
assert.False(t, isYAMLFile("readme.txt"))
assert.False(t, isYAMLFile("config"))
}

func TestFormatNamingValidationResult(t *testing.T) {
t.Parallel()

result := &NamingValidationResult{Path: "/test", Valid: true}
output := FormatNamingValidationResult(result)
assert.Contains(t, output, "/test")
assert.Contains(t, output, "PASS")

result2 := &NamingValidationResult{
Path: "/test", Valid: false,
Errors: []string{"error1"},
Warnings: []string{"warning1"},
}
output2 := FormatNamingValidationResult(result2)
assert.Contains(t, output2, "FAIL")
assert.Contains(t, output2, "error1")
assert.Contains(t, output2, "warning1")
}

func TestValidateNaming_PathDoesNotExist(t *testing.T) {
t.Parallel()

dir := filepath.Join(t.TempDir(), "nonexistent")

result, err := ValidateNaming(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.True(t, len(result.Errors) > 0)
assert.Contains(t, result.Errors[0], "does not exist")
}

func TestValidateNaming_ComposeFileReadError(t *testing.T) {
t.Parallel()

dir := t.TempDir()
// Create a directory with compose name (not readable as file).
require.NoError(t, os.MkdirAll(filepath.Join(dir, "compose.yml"), 0o755))

result, err := ValidateNaming(dir)
require.NoError(t, err)
require.NotNil(t, result)
	// Directory named compose.yml violates kebab-case validation, so result should be invalid.
	// The compose service name validation won't run because it can't read a directory as a file.
	assert.False(t, result.Valid)
}

func TestValidateNaming_InvalidFileNames(t *testing.T) {
t.Parallel()

dir := t.TempDir()
require.NoError(t, os.WriteFile(filepath.Join(dir, "MyConfig.yml"), []byte("key: value"), 0o644))
require.NoError(t, os.WriteFile(filepath.Join(dir, "my_config.yaml"), []byte("key: value"), 0o644))

result, err := ValidateNaming(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.False(t, result.Valid)
assert.True(t, len(result.Errors) >= 2)
}

func TestValidateNaming_ValidComposeFile(t *testing.T) {
t.Parallel()

dir := t.TempDir()
content := `services:
  my-service:
    image: nginx
  another-service-123:
    image: postgres
`
require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(content), 0o644))

result, err := ValidateNaming(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.True(t, result.Valid)
}

func TestValidateNaming_EdgeCases(t *testing.T) {
t.Parallel()

tests := []struct {
name      string
dirname   string
wantValid bool
}{
{name: "leading hyphen", dirname: "-invalid-", wantValid: false},
{name: "trailing hyphen", dirname: "invalid-", wantValid: false},
{name: "consecutive hyphens", dirname: "my--service", wantValid: false},
{name: "numbers", dirname: "service-123", wantValid: true},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

dir := t.TempDir()
require.NoError(t, os.MkdirAll(filepath.Join(dir, tc.dirname), 0o755))

result, err := ValidateNaming(dir)
require.NoError(t, err)
require.NotNil(t, result)
assert.Equal(t, tc.wantValid, result.Valid)
})
}
}

func TestToKebabCase_EdgeCases(t *testing.T) {
t.Parallel()

assert.Equal(t, "hello-world", toKebabCase("hello  world"))
assert.Equal(t, "hello", toKebabCase("__hello__"))
assert.Equal(t, "hello-world", toKebabCase("___hello___world___"))
}
