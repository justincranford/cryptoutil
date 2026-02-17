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
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(content), 0o600))

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
		Errors:   []string{"error1"},
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

func TestValidateNaming_PathIsFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "somefile.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("content"), 0o600))

	result, err := ValidateNaming(filePath)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.True(t, len(result.Errors) > 0)
	assert.Contains(t, result.Errors[0], "not a directory")
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
	require.NoError(t, os.WriteFile(filepath.Join(dir, "MyConfig.yml"), []byte("key: value"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "my_config.yaml"), []byte("key: value"), 0o600))

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
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(content), 0o600))

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

func TestValidateNaming_WalkErrorPermission(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	subDir := filepath.Join(dir, "sub-dir")
	require.NoError(t, os.MkdirAll(subDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "test.yml"), []byte("key: value"), 0o600))

	// Remove read permission from subdirectory to trigger walk error.
	require.NoError(t, os.Chmod(subDir, 0o000))

	t.Cleanup(func() {
		// Restore permissions for cleanup.
		_ = os.Chmod(subDir, 0o755)
	})

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	// Walk error should be recorded.
	assert.False(t, result.Valid)
	assert.True(t, len(result.Errors) > 0)
	assert.Contains(t, result.Errors[0], "error accessing path")
}

func TestValidateNaming_ComposeInvalidYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	invalidYAML := []byte("services:\n  - invalid: [yaml: {broken")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), invalidYAML, 0o600))

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	// Invalid YAML produces warning, not error.
	assert.True(t, len(result.Warnings) > 0)
	assert.Contains(t, result.Warnings[0], "Failed to parse")
}

func TestValidateNaming_NonYAMLFilesIgnored(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Non-YAML files should be ignored regardless of name.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# readme"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte("{}"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Makefile"), []byte("all:"), 0o600))

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestValidateNaming_MixedValidInvalidDirs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "valid-name"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "Invalid_Name"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "another-valid"), 0o755))

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.Equal(t, 1, len(result.Errors))
	assert.Contains(t, result.Errors[0], "Invalid_Name")
}

func TestValidateNaming_NestedDirectories(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Create deeply nested structure with valid kebab-case.
	nested := filepath.Join(dir, "level-one", "level-two", "level-three")
	require.NoError(t, os.MkdirAll(nested, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(nested, "config.yml"), []byte("key: value"), 0o600))

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestValidateNaming_EmptyDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestValidateNaming_ComposeWithEmptyServices(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := "services:\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(content), 0o600))

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestValidateNaming_UPPERCASEDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "UPPER"), 0o755))

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0], "UPPER")
	assert.Contains(t, result.Errors[0], "kebab-case")
}

func TestValidateNaming_ComposeReadableFileError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	composePath := filepath.Join(dir, "compose.yml")
	require.NoError(t, os.WriteFile(composePath, []byte("services:\n  web: {}\n"), 0o600))

	// Make file unreadable to trigger os.ReadFile error in validateComposeServiceNames.
	require.NoError(t, os.Chmod(composePath, 0o000))

	t.Cleanup(func() {
		_ = os.Chmod(composePath, 0o600)
	})

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	// Should have warning about failed read.
	assert.True(t, len(result.Warnings) > 0 || len(result.Errors) > 0)
}
