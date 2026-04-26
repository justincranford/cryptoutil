package lint_deployments

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateAdmin_ValidCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setupFn func(t *testing.T) string
	}{
		{
			name: "valid deployment",
			setupFn: func(t *testing.T) string {
				t.Helper()

				return createAdminTestDeployment(t,
					"services:\n  my-app:\n    ports:\n      - \"8700:8080\"\n",
					map[string]string{
						"bind-private-address": cryptoutilSharedMagic.IPv4Loopback,
						"bind-private-port":    "9090",
					})
			},
		},
		{
			name: "admin port not exposed safe",
			setupFn: func(t *testing.T) string {
				t.Helper()

				return createAdminTestDeployment(t,
					"services:\n  my-app:\n    ports:\n      - \"8080:8080\"\n      - \"8081:8080\"\n", nil)
			},
		},
		{
			name: "no compose file",
			setupFn: func(t *testing.T) string {
				t.Helper()

				return t.TempDir()
			},
		},
		{
			name: "no config dir",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"),
					[]byte("services:\n  app:\n    ports:\n      - \"8080:8080\"\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: "invalid compose YAML",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"),
					[]byte("{{invalid yaml"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: "invalid config YAML",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				configDir := filepath.Join(dir, "config")
				require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.yml"),
					[]byte("{{invalid"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: "config non-int port",
			setupFn: func(t *testing.T) string {
				t.Helper()

				return createAdminTestDeployment(t, "", map[string]string{
					"bind-private-port": "\"not-a-number\"",
				})
			},
		},
		{
			name: "config non-string address",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				configDir := filepath.Join(dir, "config")
				require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.yml"),
					[]byte("bind-private-address: 12345\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: "config subdirectory skipped",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				configDir := filepath.Join(dir, "config")
				require.NoError(t, os.MkdirAll(filepath.Join(configDir, "subdir"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

				return dir
			},
		},
		{
			name: "config non-YAML skipped",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				configDir := filepath.Join(dir, "config")
				require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				require.NoError(t, os.WriteFile(filepath.Join(configDir, "readme.txt"),
					[]byte("not yaml"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: "compose container-only port",
			setupFn: func(t *testing.T) string {
				t.Helper()

				return createAdminTestDeployment(t,
					"services:\n  my-app:\n    ports:\n      - \"8080\"\n", nil)
			},
		},
		{
			name: "compose non-numeric port",
			setupFn: func(t *testing.T) string {
				t.Helper()

				return createAdminTestDeployment(t,
					"services:\n  my-app:\n    ports:\n      - \"abc:8080\"\n", nil)
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := tc.setupFn(t)

			result, err := ValidateAdmin(dir)
			require.NoError(t, err)
			assert.True(t, result.Valid)
			assert.Empty(t, result.Errors)
		})
	}
}

func TestValidateAdmin_Violations(t *testing.T) {
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

				return "/nonexistent/path"
			},
			wantContains: []string{"[ValidateAdmin]", "not found"},
		},
		{
			name: "path is file",
			setupFn: func(t *testing.T) string {
				t.Helper()

				f := filepath.Join(t.TempDir(), "file")
				require.NoError(t, os.WriteFile(f, []byte("x"), cryptoutilSharedMagic.CacheFilePermissions))

				return f
			},
			wantContains: []string{"[ValidateAdmin]", "not a directory"},
		},
		{
			name: "admin port exposed",
			setupFn: func(t *testing.T) string {
				t.Helper()

				return createAdminTestDeployment(t,
					"services:\n  my-app:\n    ports:\n      - \"9090:9090\"\n", nil)
			},
			wantContains: []string{"[ValidateAdmin]", "SECURITY VIOLATION", "9090", "ENG-HANDBOOK.md Section 5.3"},
		},
		{
			name: "wrong admin port",
			setupFn: func(t *testing.T) string {
				t.Helper()

				return createAdminTestDeployment(t, "", map[string]string{
					"bind-private-port": "8443",
				})
			},
			wantContains: []string{"[ValidateAdmin]", "bind-private-port is 8443", "ENG-HANDBOOK.md Section 5.3"},
		},
		{
			name: "wrong admin address",
			setupFn: func(t *testing.T) string {
				t.Helper()

				return createAdminTestDeployment(t, "", map[string]string{
					"bind-private-address": cryptoutilSharedMagic.IPv4AnyAddress,
				})
			},
			wantContains: []string{"[ValidateAdmin]", "bind-private-address", cryptoutilSharedMagic.IPv4AnyAddress, "ENG-HANDBOOK.md Section 5.3"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := tc.setupFn(t)

			result, err := ValidateAdmin(dir)
			require.NoError(t, err)
			assert.False(t, result.Valid)

			for _, want := range tc.wantContains {
				assert.Contains(t, result.Errors[0], want)
			}
		})
	}
}

func TestValidateAdmin_InternalEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   func(string, *AdminValidationResult)
		path string
	}{
		{name: "unreadable config dir", fn: validateAdminConfigSettings, path: "/nonexistent/dir"},
		{name: "unreadable compose file", fn: validateAdminNotExposed, path: "/nonexistent/compose.yml"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := &AdminValidationResult{Valid: true}
			tc.fn(tc.path, result)
			assert.True(t, result.Valid)
		})
	}
}

func TestValidateAdmin_ConfigUnreadableFile(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("os.Chmod 0o000 does not restrict access on Windows NTFS")
	}

	dir := t.TempDir()
	configDir := filepath.Join(dir, "config")
	require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.Symlink("/nonexistent/broken.yml",
		filepath.Join(configDir, "broken.yml")))

	result, err := ValidateAdmin(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
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
