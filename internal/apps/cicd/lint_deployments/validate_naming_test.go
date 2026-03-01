package lint_deployments

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateNaming_Simple(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "service-one"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Valid)
}

func TestValidateNaming_InvalidPascalCase(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "ServiceOne"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.Valid)
	require.True(t, len(result.Errors) > 0)
	require.Contains(t, result.Errors[0], "[ValidateNaming]")
	require.Contains(t, result.Errors[0], "ServiceOne")
	require.Contains(t, result.Errors[0], "kebab-case")
	require.Contains(t, result.Errors[0], "ARCHITECTURE.md Section 4.4.1")
}

func TestValidateNaming_InvalidSnakeCase(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "service_one"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.Valid)
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
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.Valid)
	require.True(t, len(result.Errors) >= 2)
}

func TestIsKebabCase(t *testing.T) {
	t.Parallel()

	require.True(t, isKebabCase("hello"))
	require.True(t, isKebabCase("hello-world"))
	require.True(t, isKebabCase("service-123"))
	require.False(t, isKebabCase("HelloWorld"))
	require.False(t, isKebabCase("hello_world"))
	require.False(t, isKebabCase("-hello"))
	require.False(t, isKebabCase("hello-"))
	require.False(t, isKebabCase("hello--world"))
}

func TestToKebabCase(t *testing.T) {
	t.Parallel()

	require.Equal(t, "helloworld", toKebabCase("HelloWorld"))
	require.Equal(t, "hello-world", toKebabCase("hello_world"))
	require.Equal(t, "hello", toKebabCase("HELLO"))
	require.Equal(t, "myconfig.yml", toKebabCase("MyConfig.yml"))
	require.Equal(t, "my-config.yaml", toKebabCase("My_Config.yaml"))
}

func TestIsYAMLFile(t *testing.T) {
	t.Parallel()

	require.True(t, isYAMLFile("config.yml"))
	require.True(t, isYAMLFile("config.yaml"))
	require.False(t, isYAMLFile("readme.txt"))
	require.False(t, isYAMLFile("config"))
}

func TestFormatNamingValidationResult(t *testing.T) {
	t.Parallel()

	result := &NamingValidationResult{Path: "/test", Valid: true}
	output := FormatNamingValidationResult(result)
	require.Contains(t, output, "/test")
	require.Contains(t, output, cryptoutilSharedMagic.TestStatusPass)

	result2 := &NamingValidationResult{
		Path: "/test", Valid: false,
		Errors:   []string{"error1"},
		Warnings: []string{"warning1"},
	}
	output2 := FormatNamingValidationResult(result2)
	require.Contains(t, output2, cryptoutilSharedMagic.TestStatusFail)
	require.Contains(t, output2, "error1")
	require.Contains(t, output2, "warning1")
}

func TestValidateNaming_PathDoesNotExist(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "nonexistent")

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.Valid)
	require.True(t, len(result.Errors) > 0)
	require.Contains(t, result.Errors[0], "[ValidateNaming]")
	require.Contains(t, result.Errors[0], "does not exist")
}

func TestValidateNaming_PathIsFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "somefile.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("content"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateNaming(filePath)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.Valid)
	require.True(t, len(result.Errors) > 0)
	require.Contains(t, result.Errors[0], "[ValidateNaming]")
	require.Contains(t, result.Errors[0], "not a directory")
}

func TestValidateNaming_ComposeFileReadError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Create a directory with compose name (not readable as file).
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "compose.yml"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	// Directory named compose.yml violates kebab-case validation, so result should be invalid.
	// The compose service name validation won't run because it can't read a directory as a file.
	require.False(t, result.Valid)
}

func TestValidateNaming_InvalidFileNames(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "MyConfig.yml"), []byte("key: value"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "my_config.yaml"), []byte("key: value"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.Valid)
	require.True(t, len(result.Errors) >= 2)
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
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Valid)
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
			require.NoError(t, os.MkdirAll(filepath.Join(dir, tc.dirname), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

			result, err := ValidateNaming(dir)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, tc.wantValid, result.Valid)
		})
	}
}

func TestToKebabCase_EdgeCases(t *testing.T) {
	t.Parallel()

	require.Equal(t, "hello-world", toKebabCase("hello  world"))
	require.Equal(t, "hello", toKebabCase("__hello__"))
	require.Equal(t, "hello-world", toKebabCase("___hello___world___"))
}

func TestValidateNaming_WalkErrorPermission(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	subDir := filepath.Join(dir, "sub-dir")
	require.NoError(t, os.MkdirAll(subDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "test.yml"), []byte("key: value"), cryptoutilSharedMagic.CacheFilePermissions))

	// Remove read permission from subdirectory to trigger walk error.
	require.NoError(t, os.Chmod(subDir, 0o000))

	t.Cleanup(func() {
		// Restore permissions for cleanup.
		_ = os.Chmod(subDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
	})

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	// Walk error should be recorded.
	require.False(t, result.Valid)
	require.True(t, len(result.Errors) > 0)
	require.Contains(t, result.Errors[0], "[ValidateNaming]")
	require.Contains(t, result.Errors[0], "error accessing path")
}

func TestValidateNaming_ComposeInvalidYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	invalidYAML := []byte("services:\n  - invalid: [yaml: {broken")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), invalidYAML, cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	// Invalid YAML produces warning, not error.
	require.True(t, len(result.Warnings) > 0)
	require.Contains(t, result.Warnings[0], "Failed to parse")
}

func TestValidateNaming_NonYAMLFilesIgnored(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Non-YAML files should be ignored regardless of name.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# readme"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte("{}"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Makefile"), []byte("all:"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Valid)
	require.Empty(t, result.Errors)
}

func TestValidateNaming_MixedValidInvalidDirs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "valid-name"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "Invalid_Name"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "another-valid"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.Valid)
	require.Equal(t, 1, len(result.Errors))
	require.Contains(t, result.Errors[0], "Invalid_Name")
}

func TestValidateNaming_NestedDirectories(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Create deeply nested structure with valid kebab-case.
	nested := filepath.Join(dir, "level-one", "level-two", "level-three")
	require.NoError(t, os.MkdirAll(nested, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(nested, "config.yml"), []byte("key: value"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Valid)
	require.Empty(t, result.Errors)
}

func TestValidateNaming_EmptyDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Valid)
	require.Empty(t, result.Errors)
}

func TestValidateNaming_ComposeWithEmptyServices(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := "services:\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Valid)
	require.Empty(t, result.Errors)
}

func TestValidateNaming_UPPERCASEDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "UPPER"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.Valid)
	require.Contains(t, result.Errors[0], "[ValidateNaming]")
	require.Contains(t, result.Errors[0], "UPPER")
	require.Contains(t, result.Errors[0], "kebab-case")
	require.Contains(t, result.Errors[0], "ARCHITECTURE.md Section 4.4.1")
}

func TestValidateNaming_ComposeReadableFileError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	composePath := filepath.Join(dir, "compose.yml")
	require.NoError(t, os.WriteFile(composePath, []byte("services:\n  web: {}\n"), cryptoutilSharedMagic.CacheFilePermissions))

	// Make file unreadable to trigger os.ReadFile error in validateComposeServiceNames.
	require.NoError(t, os.Chmod(composePath, 0o000))

	t.Cleanup(func() {
		_ = os.Chmod(composePath, cryptoutilSharedMagic.CacheFilePermissions)
	})

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	// Should have warning about failed read.
	require.True(t, len(result.Warnings) > 0 || len(result.Errors) > 0)
}
