package lint_deployments

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateAdmin_ValidDeployment(t *testing.T) {
	t.Parallel()

	dir := createAdminTestDeployment(t,
		"services:\n  my-app:\n    ports:\n      - \"8700:8080\"\n",
		map[string]string{
			"bind-private-address": cryptoutilSharedMagic.IPv4Loopback,
			"bind-private-port":    "9090",
		})

	result, err := ValidateAdmin(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestValidateAdmin_PathNotFound(t *testing.T) {
	t.Parallel()

	result, err := ValidateAdmin("/nonexistent/path")
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0], "[ValidateAdmin]")
	assert.Contains(t, result.Errors[0], "not found")
}

func TestValidateAdmin_PathIsFile(t *testing.T) {
	t.Parallel()

	f := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(f, []byte("x"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateAdmin(f)
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0], "[ValidateAdmin]")
	assert.Contains(t, result.Errors[0], "not a directory")
}

func TestValidateAdmin_AdminPortExposed(t *testing.T) {
	t.Parallel()

	compose := "services:\n  my-app:\n    ports:\n      - \"9090:9090\"\n"
	dir := createAdminTestDeployment(t, compose, nil)

	result, err := ValidateAdmin(dir)
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0], "[ValidateAdmin]")
	assert.Contains(t, result.Errors[0], "SECURITY VIOLATION")
	assert.Contains(t, result.Errors[0], "9090")
	assert.Contains(t, result.Errors[0], "ARCHITECTURE.md Section 5.3")
}

func TestValidateAdmin_AdminPortNotExposedSafe(t *testing.T) {
	t.Parallel()

	compose := "services:\n  my-app:\n    ports:\n      - \"8080:8080\"\n      - \"8081:8080\"\n"
	dir := createAdminTestDeployment(t, compose, nil)

	result, err := ValidateAdmin(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateAdmin_WrongAdminPort(t *testing.T) {
	t.Parallel()

	dir := createAdminTestDeployment(t, "", map[string]string{
		"bind-private-port": "8443",
	})

	result, err := ValidateAdmin(dir)
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0], "[ValidateAdmin]")
	assert.Contains(t, result.Errors[0], "bind-private-port is 8443")
	assert.Contains(t, result.Errors[0], "ARCHITECTURE.md Section 5.3")
}

func TestValidateAdmin_WrongAdminAddress(t *testing.T) {
	t.Parallel()

	dir := createAdminTestDeployment(t, "", map[string]string{
		"bind-private-address": cryptoutilSharedMagic.IPv4AnyAddress,
	})

	result, err := ValidateAdmin(dir)
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0], "[ValidateAdmin]")
	assert.Contains(t, result.Errors[0], "bind-private-address")
	assert.Contains(t, result.Errors[0], cryptoutilSharedMagic.IPv4AnyAddress)
	assert.Contains(t, result.Errors[0], "ARCHITECTURE.md Section 5.3")
}

func TestValidateAdmin_NoComposeFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	result, err := ValidateAdmin(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateAdmin_NoConfigDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"),
		[]byte("services:\n  app:\n    ports:\n      - \"8080:8080\"\n"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateAdmin(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateAdmin_InvalidComposeYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"),
		[]byte("{{invalid yaml"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateAdmin(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid) // Invalid YAML silently skipped.
}

func TestValidateAdmin_InvalidConfigYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configDir := filepath.Join(dir, "config")
	require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.yml"),
		[]byte("{{invalid"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateAdmin(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateAdmin_ConfigNonIntPort(t *testing.T) {
	t.Parallel()

	dir := createAdminTestDeployment(t, "", map[string]string{
		"bind-private-port": "\"not-a-number\"",
	})

	result, err := ValidateAdmin(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid) // Non-integer port silently skipped.
}

func TestValidateAdmin_ConfigNonStringAddress(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configDir := filepath.Join(dir, "config")
	require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.yml"),
		[]byte("bind-private-address: 12345\n"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateAdmin(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid) // Non-string address silently skipped.
}

func TestValidateAdmin_ConfigSubdirectorySkipped(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configDir := filepath.Join(dir, "config")
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, "subdir"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	result, err := ValidateAdmin(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateAdmin_ConfigNonYAMLSkipped(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configDir := filepath.Join(dir, "config")
	require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "readme.txt"),
		[]byte("not yaml"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateAdmin(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateAdmin_ConfigUnreadableFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configDir := filepath.Join(dir, "config")
	require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.Symlink("/nonexistent/broken.yml",
		filepath.Join(configDir, "broken.yml")))

	result, err := ValidateAdmin(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateAdmin_UnreadableConfigDir(t *testing.T) {
	t.Parallel()

	result := &AdminValidationResult{Valid: true}
	validateAdminConfigSettings("/nonexistent/dir", result)
	assert.True(t, result.Valid) // ReadDir error silently handled.
}

func TestValidateAdmin_UnreadableComposeFile(t *testing.T) {
	t.Parallel()

	result := &AdminValidationResult{Valid: true}
	validateAdminNotExposed("/nonexistent/compose.yml", result)
	assert.True(t, result.Valid) // ReadFile error silently handled.
}

func TestValidateAdmin_ComposeContainerOnlyPort(t *testing.T) {
	t.Parallel()

	compose := "services:\n  my-app:\n    ports:\n      - \"8080\"\n"
	dir := createAdminTestDeployment(t, compose, nil)

	result, err := ValidateAdmin(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid) // Container-only port has no host mapping.
}

func TestValidateAdmin_ComposeNonNumericPort(t *testing.T) {
	t.Parallel()

	compose := "services:\n  my-app:\n    ports:\n      - \"abc:8080\"\n"
	dir := createAdminTestDeployment(t, compose, nil)

	result, err := ValidateAdmin(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid) // Non-numeric ports silently skipped.
}

func TestValidateAdmin_RealSmIM(t *testing.T) {
	t.Parallel()

	deploymentDir := filepath.Join("testdata", "deployments", cryptoutilSharedMagic.OTLPServiceSMIM)
	if _, err := os.Stat(deploymentDir); err != nil {
		deploymentDir = filepath.Join("..", "..", "..", "..", "deployments", cryptoutilSharedMagic.OTLPServiceSMIM)
	}

	if _, err := os.Stat(deploymentDir); err != nil {
		t.Skip("Real sm-im deployment not found")
	}

	result, err := ValidateAdmin(deploymentDir)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Valid, "Real sm-im admin validation failed: %v", result.Errors)
}

func TestFormatAdminValidationResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *AdminValidationResult
		contains []string
	}{
		{
			name:     "nil result",
			result:   nil,
			contains: []string{"No admin validation result"},
		},
		{
			name:     "valid",
			result:   &AdminValidationResult{Valid: true},
			contains: []string{"PASSED"},
		},
		{
			name: "errors and warnings",
			result: &AdminValidationResult{
				Valid:    false,
				Errors:   []string{"err1"},
				Warnings: []string{"warn1"},
			},
			contains: []string{cryptoutilSharedMagic.TaskFailed, "ERROR: err1", "WARNING: warn1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output := FormatAdminValidationResult(tc.result)
			for _, want := range tc.contains {
				assert.Contains(t, output, want)
			}
		})
	}
}

// createAdminTestDeployment creates a temp deployment with optional compose and config.
func createAdminTestDeployment(t *testing.T, composeContent string, configFields map[string]string) string {
	t.Helper()

	dir := t.TempDir()

	if composeContent != "" {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"),
			[]byte(composeContent), cryptoutilSharedMagic.CacheFilePermissions))
	}

	if configFields != nil {
		configDir := filepath.Join(dir, "config")
		require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

		var content string
		for k, v := range configFields {
			content += k + ": " + v + "\n"
		}

		require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.yml"),
			[]byte(content), cryptoutilSharedMagic.CacheFilePermissions))
	}

	return dir
}
