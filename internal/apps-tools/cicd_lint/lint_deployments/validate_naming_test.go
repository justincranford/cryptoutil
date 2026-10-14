package lint_deployments

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestIsKebabCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  bool
	}{
		{cryptoutilSharedMagic.TEST_NEEDLE_HELLO, true},
		{testHelloWorldValueAFA27B44, true},
		{testService123, true},
		{testHelloWorld, false},
		{testHelloWorldValue35072C1A, false},
		{"-hello", false},
		{"hello-", false},
		{"hello--world", false},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, isKebabCase(tc.input))
		})
	}
}

func TestToKebabCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{testHelloWorld, "helloworld"},
		{testHelloWorldValue35072C1A, testHelloWorldValueAFA27B44},
		{"HELLO", cryptoutilSharedMagic.TEST_NEEDLE_HELLO},
		{"MyConfig.yml", "myconfig.yml"},
		{"My_Config.yaml", "my-config.yaml"},
		{"hello  world", testHelloWorldValueAFA27B44},
		{"__hello__", cryptoutilSharedMagic.TEST_NEEDLE_HELLO},
		{"___hello___world___", testHelloWorldValueAFA27B44},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, toKebabCase(tc.input))
		})
	}
}

func TestIsYAMLFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  bool
	}{
		{testConfigYml, true},
		{"config.yaml", true},
		{"readme.txt", false},
		{fileTypeConfig, false},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, isYAMLFile(tc.input))
		})
	}
}

func TestValidateNaming_ValidCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setupFn func(t *testing.T) string
	}{
		{
			name: "simple kebab dir",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "service-one"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

				return dir
			},
		},
		{
			name: "valid compose services",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				content := "services:\n  my-service:\n    image: nginx\n  another-service-123:\n    image: postgres\n"
				require.NoError(t, os.WriteFile(filepath.Join(dir, cryptoutilSharedMagic.COMPOSE_YML), []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: testNonYAMLFilesIgnored,
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(dir, valueReadmeMd), []byte("# readme"), cryptoutilSharedMagic.CacheFilePermissions))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte("{}"), cryptoutilSharedMagic.CacheFilePermissions))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "Makefile"), []byte("all:"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: testNestedDirectories,
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				nested := filepath.Join(dir, "level-one", "level-two", "level-three")
				require.NoError(t, os.MkdirAll(nested, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				require.NoError(t, os.WriteFile(filepath.Join(nested, testConfigYml), []byte("key: value"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: testEmptyDirectory,
			setupFn: func(t *testing.T) string {
				t.Helper()

				return t.TempDir()
			},
		},
		{
			name: "compose with empty services",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(dir, cryptoutilSharedMagic.COMPOSE_YML), []byte("services:\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: "skips template directory",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				templateDir := filepath.Join(dir, DeploymentTypeTemplate)
				require.NoError(t, os.MkdirAll(templateDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				require.NoError(t, os.WriteFile(
					filepath.Join(templateDir, "PRODUCT-SERVICE.yml"),
					[]byte("name: template\n"), cryptoutilSharedMagic.CacheFilePermissions,
				))

				return dir
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := tc.setupFn(t)

			result, err := ValidateNaming(dir)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.True(t, result.Valid)
			require.Empty(t, result.Errors)
		})
	}
}

func TestValidateNaming_Violations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setupFn      func(t *testing.T) string
		minErrors    int
		wantContains []string
	}{
		{
			name: "invalid PascalCase dir",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "ServiceOne"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

				return dir
			},
			wantContains: []string{testValidateNaming, "ServiceOne", validatorNameKebabCase, testEngHandbookMdSection441},
		},
		{
			name: "invalid snake_case dir",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "service_one"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

				return dir
			},
		},
		{
			name: "invalid compose service names",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				content := "services:\n  my_service:\n    image: nginx\n  AnotherService:\n    image: postgres\n"
				require.NoError(t, os.WriteFile(filepath.Join(dir, cryptoutilSharedMagic.DOCKER_COMPOSE_YML), []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
			minErrors: 2,
		},
		{
			name: testPathDoesNotExist,
			setupFn: func(t *testing.T) string {
				t.Helper()

				return filepath.Join(t.TempDir(), "nonexistent")
			},
			wantContains: []string{testValidateNaming, testDoesNotExist},
		},
		{
			name: testPathIsFile,
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				filePath := filepath.Join(dir, "somefile.txt")
				require.NoError(t, os.WriteFile(filePath, []byte("content"), cryptoutilSharedMagic.CacheFilePermissions))

				return filePath
			},
			wantContains: []string{testValidateNaming, testNotADirectory},
		},
		{
			name: "compose file read error",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.MkdirAll(filepath.Join(dir, cryptoutilSharedMagic.COMPOSE_YML), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

				return dir
			},
		},
		{
			name: "invalid file names",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "MyConfig.yml"), []byte("key: value"), cryptoutilSharedMagic.CacheFilePermissions))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "my_config.yaml"), []byte("key: value"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
			minErrors: 2,
		},
		{
			name: "mixed valid and invalid dirs",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "valid-name"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "Invalid_Name"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "another-valid"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

				return dir
			},
			wantContains: []string{"Invalid_Name"},
		},
		{
			name: "UPPERCASE dir",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "UPPER"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

				return dir
			},
			wantContains: []string{testValidateNaming, "UPPER", validatorNameKebabCase, testEngHandbookMdSection441},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := tc.setupFn(t)

			result, err := ValidateNaming(dir)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.False(t, result.Valid)
			require.True(t, len(result.Errors) > 0)

			if tc.minErrors > 0 {
				require.True(t, len(result.Errors) >= tc.minErrors)
			}

			for _, want := range tc.wantContains {
				require.Contains(t, result.Errors[0], want)
			}
		})
	}
}

func TestValidateNaming_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		dirname   string
		wantValid bool
	}{
		{name: testLeadingHyphen, dirname: "-invalid-", wantValid: false},
		{name: "trailing hyphen", dirname: "invalid-", wantValid: false},
		{name: "consecutive hyphens", dirname: "my--service", wantValid: false},
		{name: "numbers", dirname: testService123, wantValid: true},
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

func TestFormatNamingValidationResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *NamingValidationResult
		contains []string
	}{
		{
			name:     testValid,
			result:   &NamingValidationResult{Path: testTest, Valid: true},
			contains: []string{testTest, cryptoutilSharedMagic.TestStatusPass},
		},
		{
			name: testErrorsAndWarnings,
			result: &NamingValidationResult{
				Path: testTest, Valid: false,
				Errors:   []string{testError1},
				Warnings: []string{testWarning1},
			},
			contains: []string{cryptoutilSharedMagic.TestStatusFail, testError1, testWarning1},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output := FormatNamingValidationResult(tc.result)
			for _, want := range tc.contains {
				require.Contains(t, output, want)
			}
		})
	}
}

func TestValidateNaming_WalkErrorPermission(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("os.Chmod 0o000 does not restrict access on Windows NTFS")
	}

	dir := t.TempDir()
	subDir := filepath.Join(dir, "sub-dir")
	require.NoError(t, os.MkdirAll(subDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, testTestYml), []byte("key: value"), cryptoutilSharedMagic.CacheFilePermissions))

	// Remove read permission from subdirectory to trigger walk error.
	require.NoError(t, os.Chmod(subDir, 0o000))

	t.Cleanup(func() {
		// Restore permissions for cleanup.
		_ = os.Chmod(subDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
	})

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.Valid)
	require.True(t, len(result.Errors) > 0)
	require.Contains(t, result.Errors[0], testValidateNaming)
	require.Contains(t, result.Errors[0], "error accessing path")
}

func TestValidateNaming_ComposeInvalidYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	invalidYAML := []byte("services:\n  - invalid: [yaml: {broken")
	require.NoError(t, os.WriteFile(filepath.Join(dir, cryptoutilSharedMagic.COMPOSE_YML), invalidYAML, cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, len(result.Warnings) > 0)
	require.Contains(t, result.Warnings[0], "Failed to parse")
}

func TestValidateNaming_ComposeReadableFileError(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("os.Chmod 0o000 does not restrict access on Windows NTFS")
	}

	dir := t.TempDir()
	composePath := filepath.Join(dir, cryptoutilSharedMagic.COMPOSE_YML)
	require.NoError(t, os.WriteFile(composePath, []byte("services:\n  web: {}\n"), cryptoutilSharedMagic.CacheFilePermissions))

	require.NoError(t, os.Chmod(composePath, 0o000))

	t.Cleanup(func() {
		_ = os.Chmod(composePath, cryptoutilSharedMagic.CacheFilePermissions)
	})

	result, err := ValidateNaming(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, len(result.Warnings) > 0 || len(result.Errors) > 0)
}
